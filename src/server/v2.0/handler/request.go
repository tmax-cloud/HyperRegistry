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

package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	robotSec "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/controller/request"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	pkgModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/request"
	"strconv"
	"strings"
)

type requestsAPI struct {
	BaseAPI
	requestCtl request.Controller
	projectCtl project.Controller
	quotaCtl   quota.Controller
	userCtl    user.Controller
	getAuth    func(ctx context.Context) (string, error) // For testing
}

func newRequestsAPI() *requestsAPI {
	return &requestsAPI{
		requestCtl: request.Ctl,
		projectCtl: project.Ctl,
		quotaCtl:   quota.Ctl,
		userCtl:    user.Ctl,
		getAuth:    config.AuthMode,
	}
}

func (a *requestsAPI) CreateRequest(ctx context.Context, params operation.CreateRequestParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}

	onlyAdmin, err := config.OnlyAdminCreateProject(ctx)
	if err != nil {
		return a.SendError(ctx, fmt.Errorf("failed to determine whether only admin can create projects: %v", err))
	}

	secCtx, _ := security.FromContext(ctx)
	if onlyAdmin && !(!a.isSysAdmin(ctx, rbac.ActionCreate) || secCtx.IsSolutionUser()) {
		log.Errorf("Only non sys admin can create request")
		return a.SendError(ctx, errors.ForbiddenError(nil).WithMessage("Only non system admin can create request"))
	}

	req := params.Request
	// validate the RegistryID and StorageLimit in the body of the request
	if err := a.validateRequestReq(ctx, req); err != nil {
		return a.SendError(ctx, err)
	}

	var ownerName string
	var ownerID int

	ownerName = secCtx.GetUsername()
	user, err := a.userCtl.GetByName(ctx, ownerName)
	if err != nil {
		return a.SendError(ctx, err)
	}
	ownerID = user.UserID

	p := &request.Request{
		Name:      req.Name,
		OwnerID:   ownerID,
		OwnerName: ownerName,
	}

	requestID, err := a.requestCtl.Create(ctx, p)
	if err != nil {
		return a.SendError(ctx, err)
	}

	var location string
	if lib.BoolValue(params.XResourceNameInLocation) {
		location = fmt.Sprintf("%s/%s", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), req.Name)
	} else {
		location = fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), requestID)
	}

	return operation.NewCreateRequestCreated().WithLocation(location)
}

func (a *requestsAPI) DeleteRequest(ctx context.Context, params operation.DeleteRequestParams) middleware.Responder {
	if !a.isSysAdmin(ctx, rbac.ActionDelete) {

	}
	requestNameOrID := parseRequestNameOrID(params.RequestNameOrID, params.XIsResourceName)
	//if err := a.RequireProjectAccess(ctx, requestNameOrID, rbac.ActionDelete); err != nil {
	//	return a.SendError(ctx, err)
	//}

	p, result, err := a.deletable(ctx, requestNameOrID)
	if err != nil {
		return a.SendError(ctx, err)
	}

	if !result.Deletable {
		return a.SendError(ctx, errors.PreconditionFailedError(errors.New(result.Message)))
	}

	if err := a.requestCtl.Delete(ctx, p.RequestID); err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewDeleteRequestOK()
}

func (a *requestsAPI) GetRequest(ctx context.Context, params operation.GetRequestParams) middleware.Responder {
	panic("implement me")
}

func (a *requestsAPI) ListRequests(ctx context.Context, params operation.ListRequestsParams) middleware.Responder {
	query, err := a.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}

	if name := lib.StringValue(params.Name); name != "" {
		query.Keywords["name"] = &q.FuzzyMatchValue{Value: name}
	}

	secCtx, ok := security.FromContext(ctx)
	if ok && secCtx.IsAuthenticated() {
		if !a.isSysAdmin(ctx, rbac.ActionList) && !secCtx.IsSolutionUser() {
			// authenticated but not system admin or solution user,
			// return public projects and projects that the user is member of
			if l, ok := secCtx.(*local.SecurityContext); ok {
				currentUser := l.User()
				member := &project.MemberQuery{
					UserID:   currentUser.UserID,
					GroupIDs: currentUser.GroupIDs,
				}

				query.Keywords["member"] = member
			} else if r, ok := secCtx.(*robotSec.SecurityContext); ok {
				// for the system level robot that covers all the project, see it as the system admin.
				var coverAll bool
				var names []string
				for _, p := range r.User().Permissions {
					if p.IsCoverAll() {
						coverAll = true
						break
					}
					names = append(names, p.Namespace)
				}
				if !coverAll {
					namesQuery := &pkgModels.NamesQuery{
						Names: names,
					}
					if public, ok := query.Keywords["public"]; !ok || lib.ToBool(public) {
						namesQuery.WithPublic = true
					}
					query.Keywords["names"] = namesQuery
				}
			}
		}
	}

	total, err := a.requestCtl.Count(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}

	if total == 0 {
		// no projects found for the query return directly
		return operation.NewListRequestsOK().WithXTotalCount(0).WithPayload([]*models.Request{})
	}

	requests, err := a.requestCtl.List(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}

	var payload []*models.Request
	for _, r := range requests {
		payload = append(payload, model.NewRequest(r).ToSwagger())
	}

	return operation.NewListRequestsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(payload)
}

func (a *requestsAPI) ApproveRequest(ctx context.Context, params operation.ApproveRequestParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}

	onlyAdmin, err := config.OnlyAdminCreateProject(ctx)
	if err != nil {
		return a.SendError(ctx, fmt.Errorf("failed to determine whether only admin can create projects: %v", err))
	}

	secCtx, _ := security.FromContext(ctx)
	if onlyAdmin && !(a.isSysAdmin(ctx, rbac.ActionCreate) || secCtx.IsSolutionUser()) {
		log.Errorf("Only sys admin can allowed to create request")
		return a.SendError(ctx, errors.ForbiddenError(nil).WithMessage("Only non system admin can create request"))
	}

	requestNameOrID := parseRequestNameOrID(params.RequestNameOrID, params.XIsResourceName)
	req, err := a.getRequest(ctx, requestNameOrID)
	if err != nil {
		return a.SendError(ctx, err)
	}

	projectID, err := a.projectCtl.Create(ctx, &project.Project{
		Name:    req.Name,
		OwnerID: req.OwnerID,
	})
	if err != nil {
		return a.SendError(ctx, err)
	}

	// populate storage limit
	if config.QuotaPerProjectEnable(ctx) {
		// the security context is not sys admin, set the StorageLimit the global StoragePerProject
		setting, err := config.QuotaSetting(ctx)
		if err != nil {
			log.Errorf("failed to get quota setting: %v", err)
			return a.SendError(ctx, fmt.Errorf("failed to get quota setting: %v", err))
		}
		referenceID := quota.ReferenceID(projectID)
		hardLimits := types.ResourceList{types.ResourceStorage: setting.StoragePerProject}
		if _, err := a.quotaCtl.Create(ctx, quota.ProjectReference, referenceID, hardLimits); err != nil {
			return a.SendError(ctx, fmt.Errorf("failed to create quota for project: %v", err))
		}
	}

	if err := a.requestCtl.Approve(ctx, req); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewApproveRequestOK()
}

func (a *requestsAPI) RejectRequest(ctx context.Context, params operation.RejectRequestParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}

	onlyAdmin, err := config.OnlyAdminCreateProject(ctx)
	if err != nil {
		return a.SendError(ctx, fmt.Errorf("failed to determine whether only admin can create projects: %v", err))
	}

	secCtx, _ := security.FromContext(ctx)
	if onlyAdmin && !(a.isSysAdmin(ctx, rbac.ActionCreate) || secCtx.IsSolutionUser()) {
		log.Errorf("Only sys admin can allowed to create request")
		return a.SendError(ctx, errors.ForbiddenError(nil).WithMessage("Only non system admin can create request"))
	}

	requestNameOrID := parseRequestNameOrID(params.RequestNameOrID, params.XIsResourceName)
	req, err := a.requestCtl.Get(ctx, requestNameOrID)
	if err != nil {
		return a.SendError(ctx, err)
	}

	if err := a.requestCtl.Reject(ctx, req); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewRejectRequestOK()
}

func (a *requestsAPI) isSysAdmin(ctx context.Context, action rbac.Action) bool {
	if err := a.RequireSystemAccess(ctx, action, rbac.ResourceRequest); err != nil {
		return false
	}
	return true
}

func (a *requestsAPI) validateRequestReq(ctx context.Context, req *models.Request) error {
	// TODO:
	//if req.StorageLimit != nil {
	//	hardLimits := types.ResourceList{types.ResourceStorage: *req.StorageLimit}
	//	if err := quota.Validate(ctx, quota.ProjectReference, hardLimits); err != nil {
	//		return errors.BadRequestError(err)
	//	}
	//}

	return nil
}

func (a *requestsAPI) getRequest(ctx context.Context, requestNameOrID interface{}, options ...request.Option) (*request.Request, error) {
	r, err := a.requestCtl.Get(ctx, requestNameOrID, options...)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (a *requestsAPI) deletable(ctx context.Context, requestNameOrID interface{}) (*request.Request, *models.ProjectDeletable, error) {
	r, err := a.getRequest(ctx, requestNameOrID)
	if err != nil {
		return nil, nil, err
	}

	result := &models.ProjectDeletable{Deletable: true}
	//if p.RepoCount > 0 {
	//	result.Deletable = false
	//	result.Message = "the project contains repositories, can not be deleted"
	//} else if p.ChartCount > 0 {
	//	result.Deletable = false
	//	result.Message = "the project contains helm charts, can not be deleted"
	//}

	return r, result, nil
}

func parseRequestNameOrID(str string, isResourceName *bool) interface{} {
	if lib.BoolValue(isResourceName) {
		// always as projectName
		return str
	}

	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		// it's projectName
		return str
	}

	return v // requestID
}
