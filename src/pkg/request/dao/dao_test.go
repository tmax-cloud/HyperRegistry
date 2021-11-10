// Copyright Request Harbor Authors
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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/request/models"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = New()
}

func (suite *DaoTestSuite) WithUser(f func(int64, string), usernames ...string) {
	var username string
	if len(usernames) > 0 {
		username = usernames[0]
	} else {
		username = suite.RandString(5)
	}

	o, err := orm.FromContext(orm.Context())
	if err != nil {
		suite.Fail("got error %v", err)
	}

	var userID int64

	email := fmt.Sprintf("%s@example.com", username)
	sql := "INSERT INTO harbor_user (username, realname, email, password) VALUES (?, ?, ?, 'Harbor12345') RETURNING user_id"
	suite.Nil(o.Raw(sql, username, username, email).QueryRow(&userID))

	defer func() {
		o.Raw("UPDATE harbor_user SET deleted=True, username=concat_ws('#', username, user_id), email=concat_ws('#', email, user_id) WHERE user_id = ?", userID).Exec()
	}()

	f(userID, username)
}

func (suite *DaoTestSuite) TestCreate() {
	{
		request := &models.Request{
			Name:    "foobar",
			OwnerID: 1,
		}

		requestID, err := suite.dao.Create(orm.Context(), request)
		suite.Nil(err)
		suite.dao.Delete(orm.Context(), requestID)
	}

	{
		// request name duplicated
		request := &models.Request{
			Name:    "library",
			OwnerID: 1,
		}

		requestID, err := suite.dao.Create(orm.Context(), request)
		suite.Error(err)
		suite.True(errors.IsConflictErr(err))
		suite.Equal(int64(0), requestID)
	}
}

func (suite *DaoTestSuite) TestCount() {
	request := &models.Request{
		Name:    "foobar",
		OwnerID: 1,
	}

	requestID, err := suite.dao.Create(orm.Context(), request)
	suite.Nil(err)

	count, err := suite.dao.Count(orm.Context(), q.New(q.KeyWords{"request_id": requestID}))
	suite.Nil(err)
	suite.Equal(int64(1), count)

	err = suite.dao.Delete(orm.Context(), requestID)
	suite.Nil(err)

	count, err = suite.dao.Count(orm.Context(), q.New(q.KeyWords{"request_id": requestID}))
	suite.Nil(err)
	suite.Equal(int64(0), count)
}

func (suite *DaoTestSuite) TestDelete() {
	request := &models.Request{
		Name:    "foobar",
		OwnerID: 1,
	}

	requestID, err := suite.dao.Create(orm.Context(), request)
	suite.Nil(err)

	p1, err := suite.dao.Get(orm.Context(), requestID)
	suite.Nil(err)
	suite.Equal("foobar", p1.Name)

	suite.dao.Delete(orm.Context(), requestID)
	suite.Nil(err)

	p2, err := suite.dao.Get(orm.Context(), requestID)
	suite.Error(err)
	suite.True(errors.IsNotFoundErr(err))
	suite.Nil(p2)
}

func (suite *DaoTestSuite) TestGet() {
	{
		request, err := suite.dao.Get(orm.Context(), 10000)
		suite.Error(err)
		suite.True(errors.IsNotFoundErr(err))
		suite.Nil(request)
	}
}

func (suite *DaoTestSuite) TestGetByName() {
	{
		// not found
		request, err := suite.dao.GetByName(orm.Context(), "no-exist")
		suite.Error(err)
		suite.True(errors.IsNotFoundErr(err))
		suite.Nil(request)
	}
}

func (suite *DaoTestSuite) TestList() {
	requestNames := []string{"foo1", "foo2", "foo3"}

	var requestIDs []int64
	for _, requestName := range requestNames {
		request := &models.Request{
			Name:    requestName,
			OwnerID: 1,
		}
		requestID, err := suite.dao.Create(orm.Context(), request)
		if suite.Nil(err) {
			requestIDs = append(requestIDs, requestID)
		}
	}

	defer func() {
		for _, requestID := range requestIDs {
			suite.dao.Delete(orm.Context(), requestID)
		}
	}()

	{
		requests, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"request_id__in": requestIDs}))
		suite.Nil(err)
		suite.Len(requests, len(requestNames))
	}
}

func (suite *DaoTestSuite) TestListByOwner() {
	{
		requests, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"owner": "owner-not-found"}))
		suite.Nil(err)
		suite.Len(requests, 0)
	}

	{
		// single quotes in owner
		suite.WithUser(func(userID int64, username string) {
			request := &models.Request{
				Name:    "request-owner-name-include-single-quotes",
				OwnerID: int(userID),
			}
			requestID, err := suite.dao.Create(orm.Context(), request)
			suite.Nil(err)

			defer suite.dao.Delete(orm.Context(), requestID)

			requests, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"owner": username}))
			suite.Nil(err)
			suite.Len(requests, 1)
		}, "owner include single quotes ' in it")
	}

	{
		// sql inject
		suite.WithUser(func(userID int64, username string) {
			request := &models.Request{
				Name:    "request-sql-inject",
				OwnerID: int(userID),
			}
			requestID, err := suite.dao.Create(orm.Context(), request)
			suite.Nil(err)

			defer suite.dao.Delete(orm.Context(), requestID)

			requests, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"owner": username}))
			suite.Nil(err)
			suite.Len(requests, 1)
		}, "'owner' OR 1=1")
	}
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
