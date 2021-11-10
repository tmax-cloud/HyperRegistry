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

package dao

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"time"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/request/models"
)

// DAO is the data access object interface for request
type DAO interface {
	// Create create a request instance
	Create(ctx context.Context, request *models.Request) (int64, error)
	// Count returns the total count of requests according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// Delete delete the request instance by id
	Delete(ctx context.Context, id int64) error
	// Get get request instance by id
	Get(ctx context.Context, id int64) (*models.Request, error)
	// GetByName get request instance by name
	GetByName(ctx context.Context, name string) (*models.Request, error)
	// List list requests
	List(ctx context.Context, query *q.Query) ([]*models.Request, error)
	// Update request
	Update(ctx context.Context, request *models.Request, props ...string) error
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

// Create create a request instance
func (d *dao) Create(ctx context.Context, request *models.Request) (int64, error) {
	var requestID int64

	h := func(ctx context.Context) error {
		o, err := orm.FromContext(ctx)
		if err != nil {
			return err
		}

		now := time.Now()
		request.CreationTime = now
		request.UpdateTime = now

		requestID, err = o.Insert(request)
		if err != nil {
			return orm.WrapConflictError(err, "The request named %s already exists", request.Name)
		}

		return nil
	}

	if err := orm.WithTransaction(h)(orm.SetTransactionOpNameToContext(ctx, "tx-create-request")); err != nil {
		return 0, err
	}

	return requestID, nil
}

// Count returns the total count of artifacts according to the query
func (d *dao) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	query = q.MustClone(query)
	query.Keywords["deleted"] = false
	qs, err := orm.QuerySetterForCount(ctx, &models.Request{}, query)
	if err != nil {
		return 0, err
	}

	return qs.Count()
}

// Delete delete the request instance by id
func (d *dao) Delete(ctx context.Context, id int64) error {
	request, err := d.Get(ctx, id)
	if err != nil {
		return err
	}

	request.Deleted = true
	request.Name = lib.Truncate(request.Name, fmt.Sprintf("#%d", request.RequestID), 255)

	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	_, err = o.Update(request, "deleted", "name")
	return err
}

// Get get request instance by id
func (d *dao) Get(ctx context.Context, id int64) (*models.Request, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	request := &models.Request{RequestID: id, Deleted: false}
	if err = o.Read(request, "request_id", "deleted"); err != nil {
		return nil, orm.WrapNotFoundError(err, "request %d not found", id)
	}
	return request, nil
}

// GetByName get request instance by name
func (d *dao) GetByName(ctx context.Context, name string) (*models.Request, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	request := &models.Request{Name: name, Deleted: false}
	if err := o.Read(request, "name", "deleted"); err != nil {
		return nil, orm.WrapNotFoundError(err, "request %s not found", name)
	}
	return request, nil
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*models.Request, error) {
	query = q.MustClone(query)
	query.Keywords["deleted"] = false

	qs, err := orm.QuerySetter(ctx, &models.Request{}, query)
	if err != nil {
		return nil, err
	}

	requests := []*models.Request{}
	if _, err := qs.All(&requests); err != nil {
		return nil, err
	}

	return requests, nil
}

func (d *dao) Update(ctx context.Context, request *models.Request, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(request, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("request with id %d not found", request.RequestID)
	}
	return nil
}
