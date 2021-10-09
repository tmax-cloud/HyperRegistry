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

package request

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/user"

	//commonmodels "github.com/goharbor/harbor/src/common/models"
	//event "github.com/goharbor/harbor/src/controller/event/metadata"
	_ "github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	_ "github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/request"
	"github.com/goharbor/harbor/src/pkg/request/models"
)

var (
	// Ctl is a global project controller instance
	Ctl = NewController()
)

// Request alias to models.Request
type Request = models.Request

// Controller defines the operations related with blobs
type Controller interface {
	// Create create project instance
	Create(ctx context.Context, project *models.Request) (int64, error)
	// Count returns the total count of requests according to the query
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Delete delete the project by project id
	Delete(ctx context.Context, id int64) error
	// Exists returns true when the specific project exists
	Exists(ctx context.Context, requestIDOrName interface{}) (bool, error)
	// Get get the project by project id or name
	Get(ctx context.Context, requestIDOrName interface{}, options ...Option) (*models.Request, error)
	// GetByName get the request by request name
	GetByName(ctx context.Context, requestName string, options ...Option) (*models.Request, error)
	// List list requests
	List(ctx context.Context, query *q.Query, options ...Option) ([]*models.Request, error)
	// Update update the project
	Update(ctx context.Context, project *models.Request) error
}

// NewController creates an instance of the default project controller
func NewController() Controller {
	return &controller{
		requestMgr: request.Mgr,
		userMgr:    user.Mgr,
	}
}

type controller struct {
	requestMgr request.Manager
	userMgr    user.Manager
}

func (c *controller) Create(ctx context.Context, project *models.Request) (int64, error) {
	var requestID int64
	h := func(ctx context.Context) (err error) {
		requestID, err = c.requestMgr.Create(ctx, project)
		if err != nil {
			return err
		}
		return nil
	}

	if err := orm.WithTransaction(h)(orm.SetTransactionOpNameToContext(ctx, "tx-create-request")); err != nil {
		return 0, err
	}

	// TODO: fire event
	//e := &event.CreateProjectEventMetadata{
	//	ProjectID: requestID,
	//	Project:   project.Name,
	//	Operator:  operator.FromContext(ctx),
	//}
	//notification.AddEvent(ctx, e)

	return requestID, nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.requestMgr.Count(ctx, query)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	//proj, err := c.Get(ctx, id)
	//if err != nil {
	//	return err
	//}

	if err := c.requestMgr.Delete(ctx, id); err != nil {
		return err
	}

	// TODO:
	//e := &event.DeleteProjectEventMetadata{
	//	ProjectID: proj.ProjectID,
	//	Project:   proj.Name,
	//	Operator:  operator.FromContext(ctx),
	//}
	//notification.AddEvent(ctx, e)

	return nil
}

func (c *controller) Exists(ctx context.Context, requestIDOrName interface{}) (bool, error) {
	_, err := c.requestMgr.Get(ctx, requestIDOrName)
	if err == nil {
		return true, nil
	} else if errors.IsNotFoundErr(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (c *controller) Get(ctx context.Context, requestIDOrName interface{}, options ...Option) (*models.Request, error) {
	r, err := c.requestMgr.Get(ctx, requestIDOrName)
	if err != nil {
		return nil, err
	}

	if err := c.assembleRequests(ctx, models.Requests{r}); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *controller) GetByName(ctx context.Context, requestName string, options ...Option) (*models.Request, error) {
	if requestName == "" {
		return nil, errors.BadRequestError(nil).WithMessage("project name required")
	}

	p, err := c.requestMgr.Get(ctx, requestName)
	if err != nil {
		return nil, err
	}

	if err := c.assembleRequests(ctx, models.Requests{p}); err != nil {
		return nil, err
	}

	return p, nil
}

func (c *controller) List(ctx context.Context, query *q.Query, options ...Option) ([]*models.Request, error) {
	requests, err := c.requestMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return requests, nil
	}

	if err := c.assembleRequests(ctx, requests, options...); err != nil {
		return nil, err
	}

	return requests, nil
}

func (c *controller) Update(ctx context.Context, p *models.Request) error {

	return nil
}

func (c *controller) assembleRequests(ctx context.Context, requests models.Requests, options ...Option) error {
	opts := newOptions(options...)

	if opts.WithOwner {
		if err := c.loadOwners(ctx, requests); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) loadOwners(ctx context.Context, requests models.Requests) error {
	if len(requests) == 0 {
		return nil
	}

	owners, err := c.userMgr.List(ctx, q.New(q.KeyWords{"user_id__in": requests.OwnerIDs()}))
	if err != nil {
		return err
	}
	m := owners.MapByUserID()
	for _, p := range requests {
		owner, ok := m[p.OwnerID]
		if !ok {
			log.G(ctx).Warningf("the owner of project %s is not found, owner id is %d", p.Name, p.OwnerID)
			continue
		}

		p.OwnerName = owner.Username
	}

	return nil
}
