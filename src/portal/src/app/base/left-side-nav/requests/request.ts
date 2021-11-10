export class Request {
    request_id: number;
    owner_id?: number;
    name: string;
    creation_time?: Date | string;
    deleted?: number;
    is_approved?: number;
    storage_quota?: number;
    owner_name?: string;
    update_time?: Date | string;
}
