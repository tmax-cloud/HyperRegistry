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
import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { CreateRequestComponent } from './create-request/create-request.component';
import { ListRequestComponent } from './list-request/list-request.component';
import { ConfigurationService } from '../../../services/config.service';
import { SessionService } from "../../../shared/services/session.service";
import { RequestService, QuotaHardInterface } from "../../../shared/services";
import { Configuration } from "../config/config";
import { FilterComponent } from '../../../shared/components/filter/filter.component';
import { Subscription } from 'rxjs';
import { debounceTime, distinctUntilChanged, finalize, switchMap } from 'rxjs/operators';
import { Request } from './request';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { getSortingString } from "../../../shared/units/utils";

@Component({
    selector: 'requests',
    templateUrl: 'requests.component.html',
    styleUrls: ['./requests.component.scss']
})
export class RequestsComponent implements OnInit, OnDestroy {
    quotaObj: QuotaHardInterface;
    @ViewChild(CreateRequestComponent)
    creationRequest: CreateRequestComponent;

    @ViewChild(ListRequestComponent)
    listRequest: ListRequestComponent;

    requestName: string = "";

    loading: boolean = true;

    @ViewChild(FilterComponent, {static: true})
    filterComponent: FilterComponent;
    searchSub: Subscription;

    constructor(
        public configService: ConfigurationService,
        private session: SessionService,
        private reqService: RequestService,
        private msgHandler: MessageHandlerService,
    ) { }

    ngOnInit(): void {
        if (this.isSystemAdmin) {
            this.getConfigration();
        }
        if (!this.searchSub) {
            this.searchSub = this.filterComponent.filterTerms.pipe(
                debounceTime(500),
                distinctUntilChanged(),
                switchMap(projectName => {
                    // reset project list
                    this.listRequest.currentPage = 1;
                    this.listRequest.searchKeyword = projectName;
                    this.listRequest.selectedRow = [];
                    this.loading = true;
                    return this.reqService.listRequests( this.listRequest.searchKeyword,
                        this.listRequest.currentPage, this.listRequest.pageSize, getSortingString(this.listRequest.state))
                        .pipe(finalize(() => {
                            this.loading = false;
                        }));
                })).subscribe(response => {
                // Get total count
                if (response.headers) {
                    let xHeader: string = response.headers.get("X-Total-Count");
                    if (xHeader) {
                        this.listRequest.totalCount = parseInt(xHeader, 0);
                    }
                }
                this.listRequest.requests = response.body as Request[];
            }, error => {
                this.msgHandler.handleError(error);
            });
        }
    }

    ngOnDestroy() {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
            this.searchSub = null;
        }
    }

    getConfigration() {
        this.configService.getConfiguration()
            .subscribe((configurations: Configuration) => {
                this.quotaObj = {
                    storage_per_project: configurations.storage_per_project ? configurations.storage_per_project.value : -1
                };
            });
    }

    public get isSystemAdmin(): boolean {
        let account = this.session.getCurrentUser();
        return account != null && account.has_admin_role;
    }
    openModal(): void {
        this.creationRequest.newRequest();
    }

    createRequest(created: boolean) {
        if (created) {
            this.refresh();
        }
    }

    refresh(): void {
        this.requestName = "";
        this.listRequest.refresh();
    }

}
