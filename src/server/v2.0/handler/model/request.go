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

package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/request"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Project model
type Request struct {
	*request.Request
}

// ToSwagger converts the request to the swagger model
func (r *Request) ToSwagger() *models.Request {
	return &models.Request{
		CreationTime: strfmt.DateTime(r.CreationTime),
		Name:         r.Name,
		OwnerID:      int32(r.OwnerID),
		OwnerName:    r.OwnerName,
		RequestID:    int32(r.RequestID),
		UpdateTime:   strfmt.DateTime(r.UpdateTime),
		IsApproved:   int32(r.IsApproved),
		StorageQuota: &r.StorageQuota,
	}
}

// NewRequest ...
func NewRequest(p *request.Request) *Request {
	return &Request{p}
}
