import {ComponentFixture, TestBed, waitForAsync} from '@angular/core/testing';
import {CreateRequestComponent} from './create-request.component';
import {CUSTOM_ELEMENTS_SCHEMA} from '@angular/core';
import {MessageHandlerService} from '../../../../shared/services/message-handler.service';
import {of} from 'rxjs';
import {delay} from 'rxjs/operators';
import {InlineAlertComponent} from "../../../../shared/components/inline-alert/inline-alert.component";
import {SharedTestingModule} from "../../../../shared/shared.module";
import {RequestService} from "../../../../../../ng-swagger-gen/services/request.service";

describe('CreateRequestComponent', () => {
    let component: CreateRequestComponent;
    let fixture: ComponentFixture<CreateRequestComponent>;
    const mockRequestService = {
        listProjects: function (params: RequestService.ListRequestsParams) {
            if (params && params.q === encodeURIComponent('name=test')) {
                return of([true]).pipe(delay(10));
            } else {
                return of([]).pipe(delay(10));
            }
        },
        createProject: function () {
            return of(true);
        }
    };
    const mockMessageHandlerService = {
        showSuccess: function () {
        }
    };
    beforeEach(waitForAsync(() => {
        TestBed.configureTestingModule({
            imports: [
                SharedTestingModule
            ],
            declarations: [CreateRequestComponent, InlineAlertComponent],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            providers: [
                {provide: RequestService, useValue: mockRequestService},
                {provide: MessageHandlerService, useValue: mockMessageHandlerService},
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(CreateRequestComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open and close', async () => {
        let modelBody: HTMLDivElement;
        modelBody = fixture.nativeElement.querySelector(".modal-body");
        expect(modelBody).toBeFalsy();
        component.createRequestOpened = true;
        fixture.detectChanges();
        await fixture.whenStable();
        modelBody = fixture.nativeElement.querySelector(".modal-body");
        expect(modelBody).toBeTruthy();
        const cancelButton: HTMLButtonElement = fixture.nativeElement.querySelector("#new-request-cancel");
        cancelButton.click();
        fixture.detectChanges();
        await fixture.whenStable();
        modelBody = fixture.nativeElement.querySelector(".modal-body");
        expect(modelBody).toBeFalsy();
    });

    it('should check request name', async () => {
        fixture.autoDetectChanges(true);
        component.createRequestOpened = true;
        await fixture.whenStable();
        const nameInput: HTMLInputElement = fixture.nativeElement.querySelector("#create_request_name");
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        let el: HTMLSpanElement;
        el = fixture.nativeElement.querySelector('#name-error');
        expect(el).toBeTruthy();
        nameInput.value = "test";
        nameInput.dispatchEvent(new Event("input"));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        el = fixture.nativeElement.querySelector('#name-error');
        expect(el).toBeFalsy();
        nameInput.value = "test1";
        nameInput.dispatchEvent(new Event("input"));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        el = fixture.nativeElement.querySelector('#name-error');
        expect(el).toBeFalsy();
        const okButton: HTMLButtonElement = fixture.nativeElement.querySelector("#new-request-ok");
        okButton.click();
        await fixture.whenStable();
        const modelBody: HTMLDivElement = fixture.nativeElement.querySelector(".modal-body");
        expect(modelBody).toBeTruthy();
    });
});
