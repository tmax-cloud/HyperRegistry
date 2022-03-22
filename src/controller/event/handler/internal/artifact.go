// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/email"
	"github.com/goharbor/harbor/src/lib/config"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event"
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
)

// Handler preprocess artifact event data
type Handler struct {
}

// Name ...
func (a *Handler) Name() string {
	return "InternalArtifact"
}

// Handle ...
func (a *Handler) Handle(ctx context.Context, value interface{}) error {
	switch v := value.(type) {
	case *event.PullArtifactEvent:
		return a.onPull(ctx, v.ArtifactEvent)
	case *event.PushArtifactEvent:
		return a.onPush(ctx, v.ArtifactEvent)
	case *event.ApproveRequestEvent:
		return a.onApprove(ctx, v.RequestEvent)
	case *event.RejectRequestEvent:
		return a.onReject(ctx, v.RequestEvent)

	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	return nil
}

// IsStateful ...
func (a *Handler) IsStateful() bool {
	return false
}

func (a *Handler) onPull(ctx context.Context, event *event.ArtifactEvent) error {
	if !config.PullTimeUpdateDisable(ctx) {
		go func() { a.updatePullTime(ctx, event) }()
	}
	if !config.PullCountUpdateDisable(ctx) {
		go func() { a.addPullCount(ctx, event) }()
	}
	return nil
}

func (a *Handler) updatePullTime(ctx context.Context, event *event.ArtifactEvent) {
	var tagID int64
	if len(event.Tags) != 0 {
		tags, err := tag.Ctl.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"ArtifactID": event.Artifact.ID,
				"Name":       event.Tags[0],
			},
		}, nil)
		if err != nil {
			log.Infof("failed to list tags when to update pull time, %v", err)
		} else {
			if len(tags) != 0 {
				tagID = tags[0].ID
			}
		}
	}
	if err := artifact.Ctl.UpdatePullTime(ctx, event.Artifact.ID, tagID, time.Now()); err != nil {
		log.Debugf("failed to update pull time form artifact %d, %v", event.Artifact.ID, err)
	}
	return
}

func (a *Handler) addPullCount(ctx context.Context, event *event.ArtifactEvent) {
	if err := repository.Ctl.AddPullCount(ctx, event.Artifact.RepositoryID); err != nil {
		log.Debugf("failed to add pull count repository %d, %v", event.Artifact.RepositoryID, err)
	}
	return
}

func (a *Handler) onPush(ctx context.Context, event *event.ArtifactEvent) error {
	go func() {
		if err := autoScan(ctx, &artifact.Artifact{Artifact: *event.Artifact}, event.Tags...); err != nil {
			log.Errorf("scan artifact %s@%s failed, error: %v", event.Artifact.RepositoryName, event.Artifact.Digest, err)
		}
	}()

	return nil
}

func (a *Handler) onApprove(ctx context.Context, event *event.RequestEvent) error {
	go func() {
		if err := a.sendMail(ctx, event); err != nil {
			log.Errorf("send mail %s@%s failed, error: %v", event.EventType, event.Project, err)
		}
	}()

	return nil
}

func (a *Handler) onReject(ctx context.Context, event *event.RequestEvent) error {
	go func() {
		if err := a.sendMail(ctx, event); err != nil {
			log.Errorf("send mail %s@%s failed, error: %v", event.EventType, event.Project, err)
		}
	}()

	return nil
}

func (a *Handler) sendMail(ctx context.Context, event *event.RequestEvent) error {
	meta, err := config.Email(ctx)
	if err != nil {
		return err
	}

	addr := strings.TrimSpace(strings.Join([]string{meta.Host, strconv.Itoa(meta.Port)}, ":"))
	owner, err := user.Ctl.Get(ctx, event.OwnerID, &user.Option{})
	if err != nil {
		log.Errorf("cannot get (%d)'s user mail info\n", event.OwnerID)
		return err
	}

	log.Infof("host: %s/ identity: %s/ user: %s/ password: *****/ ssl: %v/ insecure: %v/ from: %s/ to: %s\n",
		addr, meta.Identity, meta.Username, meta.SSL, meta.Insecure, meta.From, owner.Email)

	var subject, message string
	switch event.EventType {
	case event2.TopicApproveRequest:
		subject = fmt.Sprintf("[HyperRegistry] Approved request")
		message = fmt.Sprintf("Hey %s! The project named %s has been created by request.", owner.Username, event.Project)
	case event2.TopicRejectRequest:
		subject = fmt.Sprintf("[HyperRegistry] Rejected request")
		message = fmt.Sprintf("Sorry %s. Please contact admin %s.", owner.Username, event.Operator)
	default:
		log.Errorf("undefined event type: %s", event.EventType)
		return nil
	}

	err = email.Send(addr, meta.Identity, meta.Username, meta.Password, 5, meta.SSL, meta.Insecure, meta.From,
		[]string{owner.Email}, subject, message)

	return err
}
