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
	"regexp"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/request/dao"
	"github.com/goharbor/harbor/src/pkg/request/models"
)

var (
	// Mgr is the global project manager
	Mgr = New()
)

// Manager is used for project management
type Manager interface {
	// Create create project instance
	Create(ctx context.Context, request *models.Request) (int64, error)

	// Count returns the total count of requests according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// Delete delete the request instance by id
	Delete(ctx context.Context, id int64) error

	// Get the request specified by the ID or name
	Get(ctx context.Context, idOrName interface{}) (*models.Request, error)

	// List requests according to the query
	List(ctx context.Context, query *q.Query) ([]*models.Request, error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{dao: dao.New()}
}

const requestNameMaxLen int = 255
const requestNameMinLen int = 1
const restrictedNameChars = `[a-z0-9]+(?:[._-][a-z0-9]+)*`

var (
	validRequestName = regexp.MustCompile(`^` + restrictedNameChars + `$`)
)

type manager struct {
	dao dao.DAO
}

// Create create request instance
func (m *manager) Create(ctx context.Context, request *models.Request) (int64, error) {
	if request.OwnerID <= 0 {
		return 0, errors.BadRequestError(nil).WithMessage("Owner is missing when creating request %s", request.Name)
	}

	if utils.IsIllegalLength(request.Name, requestNameMinLen, requestNameMaxLen) {
		format := "Request name %s is illegal in length. (greater than %d or less than %d)"
		return 0, errors.BadRequestError(nil).WithMessage(format, request.Name, requestNameMaxLen, requestNameMinLen)
	}

	legal := validRequestName.MatchString(request.Name)
	if !legal {
		return 0, errors.BadRequestError(nil).WithMessage("request name is not in lower case or contains illegal characters")
	}

	return m.dao.Create(ctx, request)
}

// Count returns the total count of requests according to the query
func (m *manager) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	return m.dao.Count(ctx, query)
}

// Delete delete the request instance by id
func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

// Get the request specified by the ID
func (m *manager) Get(ctx context.Context, idOrName interface{}) (*models.Request, error) {
	id, ok := idOrName.(int64)
	if ok {
		return m.dao.Get(ctx, id)
	}
	name, ok := idOrName.(string)
	if ok {
		return m.dao.GetByName(ctx, name)
	}
	return nil, errors.Errorf("invalid parameter: %v, should be ID(int64) or name(string)", idOrName)
}

// List requests according to the query
func (m *manager) List(ctx context.Context, query *q.Query) ([]*models.Request, error) {
	return m.dao.List(ctx, query)
}
