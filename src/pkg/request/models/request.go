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

package models

import (
	"context"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/lib/pq"
	"strings"
	"time"
)

const (
	NotDetermined = 0
	Approved      = 1
	Rejected      = 2
)

const (
	// RequestTable is the table name for request
	RequestTable = "request"
)

func init() {
	orm.RegisterModel(&Request{})
}

// Request holds the details of a project.
type Request struct {
	RequestID    int64     `orm:"pk;auto;column(request_id)" json:"request_id"`
	OwnerID      int       `orm:"column(owner_id)" json:"owner_id"`
	Name         string    `orm:"column(name)" json:"name" sort:"default"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
	Deleted      bool      `orm:"column(deleted)" json:"deleted"`
	OwnerName    string    `orm:"column(owner_name)" json:"owner_name"`
	IsApproved   int       `orm:"column(is_approved)" json:"is_approved"`
}

// NamesQuery ...
type NamesQuery struct {
	Names []string // the names of request
}

// FilterByOwner returns orm.QuerySeter with owner filter
func (p *Request) FilterByOwner(ctx context.Context, qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
	username, ok := value.(string)
	if !ok {
		return qs
	}

	return qs.FilterRaw("owner_id", fmt.Sprintf("IN (SELECT user_id FROM harbor_user WHERE username = %s)", pq.QuoteLiteral(username)))
}

// FilterByNames returns orm.QuerySeter with name filter
func (p *Request) FilterByNames(ctx context.Context, qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
	query, ok := value.(*NamesQuery)
	if !ok {
		return qs
	}

	if len(query.Names) == 0 {
		return qs
	}

	var names []string
	for _, v := range query.Names {
		names = append(names, `'`+v+`'`)
	}
	subQuery := fmt.Sprintf("SELECT request_id FROM request WHERE name IN (%s)", strings.Join(names, ","))

	return qs.FilterRaw("request_id", fmt.Sprintf("IN (%s)", subQuery))
}

func isTrue(i interface{}) bool {
	switch value := i.(type) {
	case bool:
		return value
	case string:
		v := strings.ToLower(value)
		return v == "true" || v == "1"
	default:
		return false
	}
}

// TableName is required by beego orm to map Request to table project
func (p *Request) TableName() string {
	return RequestTable
}

// Requests the connection for Request
type Requests []*Request

// OwnerIDs returns all the owner ids from the projects
func (requests Requests) OwnerIDs() []int {
	var ownerIDs []int
	for _, req := range requests {
		ownerIDs = append(ownerIDs, req.OwnerID)
	}
	return ownerIDs
}
