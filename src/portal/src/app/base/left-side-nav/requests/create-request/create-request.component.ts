// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import {debounceTime, distinctUntilChanged, filter, map, switchMap} from 'rxjs/operators';
import {
    AfterViewInit,
    Component,
    ElementRef,
    EventEmitter,
    Input,
    OnChanges,
    OnDestroy,
    OnInit,
    Output,
    SimpleChanges,
    ViewChild
} from "@angular/core";
import {NgForm, Validators} from "@angular/forms";
import {fromEvent, Subscription} from "rxjs";
import {TranslateService} from "@ngx-translate/core";
import {MessageHandlerService} from "../../../../shared/services/message-handler.service";
import {Request} from "../request";
import {QuotaUnits, QuotaUnlimited} from "../../../../shared/entities/shared.const";
import {QuotaHardInterface} from '../../../../shared/services';
import {clone, getByte, GetIntegerAndUnit, validateLimit} from "../../../../shared/units/utils";
import {InlineAlertComponent} from "../../../../shared/components/inline-alert/inline-alert.component";
import {RequestService} from "../../../../../../ng-swagger-gen/services/request.service";

const PAGE_SIZE: number = 100;

@Component({
    selector: "create-request",
    templateUrl: "create-request.component.html",
    styleUrls: ["create-request.scss"]
})
export class CreateRequestComponent implements OnInit, AfterViewInit, OnChanges, OnDestroy {

    requestForm: NgForm;

    @ViewChild("requestForm", {static: true})
    currentForm: NgForm;
    quotaUnits = QuotaUnits;
    request: Request = new Request();
    storageLimit: number;
    storageLimitUnit: string = QuotaUnits[3].UNIT;
    storageDefaultLimit: number;
    storageDefaultLimitUnit: string;
    initVal: Request = new Request();

    createRequestOpened: boolean;

    hasChanged: boolean;
    isSubmitOnGoing = false;

    staticBackdrop = true;
    closable = false;
    isNameExisted: boolean = false;
    nameTooltipText = "REQUEST.NAME_TOOLTIP";
    checkOnGoing = false;
    endpoint: string = "";
    @Output() create = new EventEmitter<boolean>();
    @Input() quotaObj: QuotaHardInterface;
    @Input() isSystemAdmin: boolean;
    @ViewChild(InlineAlertComponent, {static: true})
    inlineAlert: InlineAlertComponent;
    @ViewChild('requestName') requestNameInput: ElementRef;
    checkNameSubscribe: Subscription;

    constructor(private requestService: RequestService,
                private translateService: TranslateService,
                private messageHandlerService: MessageHandlerService) {
    }

    ngOnInit(): void {

    }

    ngAfterViewInit(): void {
        if (!this.checkNameSubscribe) {
            this.checkNameSubscribe = fromEvent(this.requestNameInput.nativeElement, 'input').pipe(
                map((e: any) => e.target.value),
                debounceTime(300),
                distinctUntilChanged(),
                filter(name => {
                    return this.currentForm.controls["create_request_name"].valid && name.length > 0;
                }),
                switchMap(name => {
                    // Check exiting from backend
                    this.checkOnGoing = true;
                    this.isNameExisted = false;
                    return this.requestService.ListRequests({
                        q: encodeURIComponent(`name=${name}`)
                    });
                })).subscribe(response => {
                // Project existing
                if (response && response.length) {
                    this.isNameExisted = true;
                }
                this.checkOnGoing = false;
            }, error => {
                this.checkOnGoing = false;
                this.isNameExisted = false;
            });
        }
    }

    get isNameValid(): boolean {
        if (!this.currentForm || !this.currentForm.controls || !this.currentForm.controls["create_request_name"]) {
            return true;
        }
        if (!(this.currentForm.controls["create_request_name"].dirty || this.currentForm.controls["create_request_name"].touched)) {
            return true;
        }
        if (this.checkOnGoing) {
            return true;
        }
        if (this.currentForm.controls["create_request_name"].errors) {
            this.nameTooltipText = 'PROJECT.NAME_TOOLTIP';
            return false;
        }
        if (this.isNameExisted) {
            this.nameTooltipText = 'PROJECT.NAME_ALREADY_EXISTS';
            return false;
        }
        return true;
    }

    ngOnChanges(changes: SimpleChanges): void {
        if (changes && changes["quotaObj"] && changes["quotaObj"].currentValue) {
            this.storageLimit = GetIntegerAndUnit(this.quotaObj.storage_per_project, clone(QuotaUnits), 0, clone(QuotaUnits)).partNumberHard;
            this.storageLimitUnit = this.storageLimit === QuotaUnlimited ? QuotaUnits[3].UNIT
                : GetIntegerAndUnit(this.quotaObj.storage_per_project, clone(QuotaUnits), 0, clone(QuotaUnits)).partCharacterHard;

            this.storageDefaultLimit = this.storageLimit;
            this.storageDefaultLimitUnit = this.storageLimitUnit;
            if (this.isSystemAdmin) {
                this.currentForm.form.controls['create_project_storage_limit'].setValidators(
                    [
                        Validators.required,
                        Validators.pattern('(^-1$)|(^([1-9]+)([0-9]+)*$)'),
                        validateLimit(this.currentForm.form.controls['create_project_storage_limit_unit'])
                    ]);
            }
            this.currentForm.form.valueChanges
                .pipe(distinctUntilChanged((a, b) => JSON.stringify(a) === JSON.stringify(b)))
                .subscribe((data) => {
                    ['create_project_storage_limit', 'create_project_storage_limit_unit'].forEach(fieldName => {
                        if (this.currentForm.form.get(fieldName) && this.currentForm.form.get(fieldName).value !== null) {
                            this.currentForm.form.get(fieldName).updateValueAndValidity();
                        }
                    });
                });
        }
    }

    ngOnDestroy(): void {
        if (this.checkNameSubscribe) {
            this.checkNameSubscribe.unsubscribe();
            this.checkNameSubscribe = null;
        }
    }

    onSubmit() {
        if (this.isSubmitOnGoing) {
            return;
        }
        this.isSubmitOnGoing = true;
        const storageByte = +this.storageLimit === QuotaUnlimited ? this.storageLimit : getByte(+this.storageLimit, this.storageLimitUnit);
        this.requestService
            .createRequest({
                request: {
                    name: this.request.name
                }
            })
            .subscribe(
                status => {
                    this.isSubmitOnGoing = false;

                    this.create.emit(true);
                    this.messageHandlerService.showSuccess("REQUEST.CREATED_SUCCESS");
                    this.createRequestOpened = false;
                },
                error => {
                    this.isSubmitOnGoing = false;
                    this.inlineAlert.showInlineError(error);
                });
    }

    onCancel() {
        this.createRequestOpened = false;
    }

    newRequest() {
        this.request = new Request();
        this.hasChanged = false;
        this.createRequestOpened = true;
        if (this.currentForm && this.currentForm.controls && this.currentForm.controls["create_request_name"]) {
            this.currentForm.controls["create_request_name"].reset();
        }
        this.inlineAlert.close();
        this.storageLimit = this.storageDefaultLimit;
        this.storageLimitUnit = this.storageDefaultLimitUnit;
    }

    public get isValid(): boolean {
        return this.currentForm &&
            this.currentForm.valid &&
            !this.isSubmitOnGoing &&
            this.isNameValid &&
            !this.checkOnGoing;
    }
}
