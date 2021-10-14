import {Observable, of, throwError as observableThrowError} from "rxjs";
import {Injectable} from "@angular/core";
import {HttpClient, HttpParams, HttpResponse} from "@angular/common/http";
import {catchError} from "rxjs/operators";
import {Request} from "../../base/left-side-nav/requests/request";
import {
    buildHttpRequestOptionsWithObserveResponse,
    CURRENT_BASE_HREF,
    HTTP_GET_OPTIONS,
    HTTP_JSON_OPTIONS
} from "../units/utils";

/**
 * Define the service methods to handle the Request related things.
 *
 **
 * @abstract
 * class RequestService
 */
export abstract class RequestService {
    /**
     * Get Informations about a specific Request.
     *
     * @abstract
     *  ** deprecated param {string|number} [RequestId]
     * returns {(Observable<Request> )}
     *
     * @memberOf RequestService
     */
    abstract getRequest(
        requestId: number | string
    ): Observable<Request>;

    /**
     * Get all Requests
     *
     * @abstract
     *  ** deprecated param {string} name
     *  ** deprecated param {number} page
     *  ** deprecated param {number} pageSize
     * returns {(Observable<any>)}
     *
     */
    abstract listRequests(
        name: string,
        page?: number,
        pageSize?: number,
        sort?: string
    ): Observable<HttpResponse<Request[]>>;
    abstract createRequest(name: string): Observable<any>;
    abstract deleteRequest(requestId: number): Observable<any>;
    abstract approveRequest(requestId: number): Observable<any>;
    abstract rejectRequest(requestId: number): Observable<any>;
    abstract checkRequestExists(requestName: string): Observable<any>;
}

/**
 * Implement default service for Request.
 *
 **
 * class RequestDefaultService
 * extends {RequestService}
 */
@Injectable()
export class RequestDefaultService extends RequestService {
    constructor(private http: HttpClient) {
        super();
    }

    public getRequest(requestId: number | string): Observable<Request> {
        if (!requestId) {
            return observableThrowError("Bad argument");
        }
        let baseUrl: string = CURRENT_BASE_HREF + "/requests";
        return this.http
            .get<Request>(`${baseUrl}/${requestId}`, HTTP_GET_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public listRequests(name: string,
                        page?: number, pageSize?: number, sort?: string): Observable<HttpResponse<Request[]>> {
        let params = new HttpParams();
        if (page && pageSize) {
            params = params.set('page', page + '').set('page_size', pageSize + '');
        }
        if (name && name.trim() !== "") {
            params = params.set('name', name);
        }
        if (sort) {
            params = params.set('sort', sort);
        }
        return this.http
            .get<HttpResponse<Request[]>>(`${CURRENT_BASE_HREF}/requests`, buildHttpRequestOptionsWithObserveResponse(params)).pipe(
                catchError(error => observableThrowError(error)));
    }

    public createRequest(name: string): Observable<any> {
        return this.http
            .post(`${CURRENT_BASE_HREF}/requests`,
                JSON.stringify({
                    'request_name': name
                })
                , HTTP_JSON_OPTIONS).pipe(
                catchError(error => observableThrowError(error)));
    }

    public deleteRequest(requestId: number): Observable<any> {
        return this.http
            .delete(`${CURRENT_BASE_HREF}/request/${requestId}`)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public approveRequest(requestId: number): Observable<any> {
        return this.http
            .put(`${CURRENT_BASE_HREF}/request/${requestId}/_approve`, "")
            .pipe(catchError(error => observableThrowError(error)));
    }

    public rejectRequest(requestId: number): Observable<any> {
        return this.http
            .put(`${CURRENT_BASE_HREF}/request/${requestId}/_reject`, "")
            .pipe(catchError(error => observableThrowError(error)));
    }

    public checkRequestExists(requestName: string): Observable<any> {
        return this.http
            .head(`${CURRENT_BASE_HREF}/request/?request_name=${requestName}`).pipe(
                catchError(error => {
                    if (error && error.status === 404) {
                        return of(error);
                    }
                    return observableThrowError(error);
                }))
        ;
    }
}
