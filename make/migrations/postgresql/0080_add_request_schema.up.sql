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
