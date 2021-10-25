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

package metadata

import (
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"time"
)

// ApproveRequestEventMetadata is the metadata from which the create project event can be resolved
type ApproveRequestEventMetadata struct {
	ProjectID int64
	Project   string
	Operator  string
}

// Resolve to the event from the metadata
func (c *ApproveRequestEventMetadata) Resolve(event *event.Event) error {
	event.Topic = event2.TopicApproveRequest
	event.Data = &event2.ApproveRequestEvent{
		&event2.RequestEvent{
			EventType: event2.TopicApproveRequest,
			Project:   c.Project,
			Operator:  c.Operator,
			OccurAt:   time.Now(),
		},
	}
	return nil
}

// RejectRequestEventMetadata is the metadata from which the delete project event can be resolved
type RejectRequestEventMetadata struct {
	ProjectID int64
	Project   string
	Operator  string
}

// Resolve to the event from the metadata
func (d *RejectRequestEventMetadata) Resolve(event *event.Event) error {
	event.Topic = event2.TopicRejectRequest
	event.Data = &event2.RejectRequestEvent{
		&event2.RequestEvent{
			EventType: event2.TopicRejectRequest,
			Project:   d.Project,
			Operator:  d.Operator,
			OccurAt:   time.Now(),
		},
	}
	return nil
}
