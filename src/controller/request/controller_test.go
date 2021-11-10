//// Copyright Request Harbor Authors
////
//// Licensed under the Apache License, Version 2.0 (the "License");
//// you may not use this file except in compliance with the License.
//// You may obtain a copy of the License at
////
////    http://www.apache.org/licenses/LICENSE-2.0
////
//// Unless required by applicable law or agreed to in writing, software
//// distributed under the License is distributed on an "AS IS" BASIS,
//// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//// See the License for the specific language governing permissions and
//// limitations under the License.
//
package request

import (
	"context"
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/request/models"
	usermodels "github.com/goharbor/harbor/src/pkg/user/models"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/request"
	"github.com/goharbor/harbor/src/testing/pkg/user"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
}

func (suite *ControllerTestSuite) TestCreate() {
	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	mgr := &request.Manager{}
	usrMgr := &user.Manager{}

	c := controller{requestMgr: mgr, userMgr: usrMgr}

	{
		mgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil).Once()
		requestID, err := c.Create(ctx, &models.Request{OwnerID: 1})
		suite.Nil(err)
		suite.Equal(int64(1), requestID)
	}

	{
		mgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil).Once()
		requestID, err := c.Create(ctx, &models.Request{OwnerID: 1, StorageQuota: 0})
		suite.Nil(err)
		suite.Equal(int64(1), requestID)
	}

	{
		mgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil).Once()
		requestID, err := c.Create(ctx, &models.Request{OwnerID: 1, StorageQuota: 1024 * 1024})
		suite.Nil(err)
		suite.Equal(int64(1), requestID)
	}
}

func (suite *ControllerTestSuite) TestGetByName() {
	ctx := context.TODO()

	mgr := &request.Manager{}
	mgr.On("Get", ctx, "existing").Return(&models.Request{RequestID: 1, Name: "existing"}, nil)
	mgr.On("Get", ctx, "not-existing").Return(nil, errors.NotFoundError(nil))
	mgr.On("Get", ctx, "oops").Return(nil, fmt.Errorf("oops"))

	c := controller{requestMgr: mgr}

	{
		p, err := c.GetByName(ctx, "existing")
		suite.Nil(err)
		suite.Equal("existing", p.Name)
		suite.Equal(int64(1), p.RequestID)
	}

	{
		p, err := c.GetByName(ctx, "not-existing")
		suite.Error(err)
		suite.True(errors.IsNotFoundErr(err))
		suite.Nil(p)
	}

	{
		p, err := c.GetByName(ctx, "oops")
		suite.Error(err)
		suite.False(errors.IsNotFoundErr(err))
		suite.Nil(p)
	}
}

func (suite *ControllerTestSuite) TestWithOwner() {
	ctx := context.TODO()

	mgr := &request.Manager{}
	mgr.On("Get", ctx, int64(1)).Return(&models.Request{RequestID: 1, OwnerID: 2, Name: "tmaxcloud"}, nil)
	mgr.On("Get", ctx, "tmaxcloud").Return(&models.Request{RequestID: 1, OwnerID: 2, Name: "tmaxcloud"}, nil)
	mgr.On("List", ctx, mock.Anything).Return([]*models.Request{
		{RequestID: 1, OwnerID: 2, Name: "tmaxcloud"},
	}, nil)

	userMgr := &user.Manager{}
	userMgr.On("List", ctx, mock.Anything).Return(usermodels.Users{
		&usermodels.User{UserID: 1, Username: "admin"},
		&usermodels.User{UserID: 2, Username: "dev"},
		&usermodels.User{UserID: 3, Username: "guest"},
	}, nil)

	c := controller{requestMgr: mgr, userMgr: userMgr}

	{
		req, err := c.Get(ctx, int64(1), WithOwner())
		suite.Nil(err)
		suite.Equal("dev", req.OwnerName)
	}

	{
		req, err := c.GetByName(ctx, "tmaxcloud", WithOwner())
		suite.Nil(err)
		suite.Equal("dev", req.OwnerName)
	}

	{
		req, err := c.List(ctx, q.New(q.KeyWords{"request_id__in": []int64{1}}), WithOwner())
		suite.Nil(err)
		suite.Len(req, 1)
		suite.Equal("dev", req[0].OwnerName)
	}
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
