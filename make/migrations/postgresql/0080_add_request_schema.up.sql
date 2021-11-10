CREATE TABLE request
(
    request_id    SERIAL PRIMARY KEY      NOT NULL,
    owner_id      int                     NOT NULL,
    name          varchar(255),
    deleted       boolean   DEFAULT false NOT NULL,
    creation_time timestamp default CURRENT_TIMESTAMP,
    update_time   timestamp default CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES harbor_user (user_id),
    UNIQUE (name)
);

ALTER TABLE request
    ADD COLUMN owner_name varchar(255) Default '-';
ALTER TABLE request
    ADD COLUMN is_approved int NOT NULL Default 0;
ALTER TABLE request
    ADD COLUMN storage_quota bigint NOT NULL Default 0;
