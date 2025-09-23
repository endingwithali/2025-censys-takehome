CREATE TABLE snapshot (
    uuid        UUID PRIMARY KEY,
    timestamp   TIMESTAMP NOT NULL,
    host_ip     VARCHAR(255) NOT NULL,
    file_pwd    TEXT NOT NULL,
    file_name   TEXT NOT NULL
);

CREATE TABLE snapshot_differences (
    host_ip     VARCHAR(255) NOT NULL,
    timestamp1  TIMESTAMP NOT NULL,
    timestamp2  TIMESTAMP NOT NULL,
    json_data   TEXT NOT NULL,
    PRIMARY KEY (host_ip, timestamp1, timestamp2)
);