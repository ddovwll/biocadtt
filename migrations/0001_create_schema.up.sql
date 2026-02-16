CREATE TABLE IF NOT EXISTS file_data
(
    id         UUID PRIMARY KEY,
    n          INTEGER     NOT NULL,
    mqtt       TEXT        NOT NULL DEFAULT '',
    invid      TEXT        NOT NULL,
    unit_guid  TEXT        NOT NULL,
    msg_id     TEXT        NOT NULL,
    text       TEXT        NOT NULL,
    context    TEXT        NOT NULL DEFAULT '',
    class      TEXT        NOT NULL,
    level      INTEGER     NOT NULL,
    area       TEXT        NOT NULL,
    addr       TEXT        NOT NULL,
    block      TEXT        NOT NULL DEFAULT '',
    type       TEXT        NOT NULL DEFAULT '',
    bit        TEXT        NOT NULL DEFAULT '',
    invert_bit TEXT        NOT NULL DEFAULT ''
);


CREATE TABLE IF NOT EXISTS processed_files
(
    id                     UUID PRIMARY KEY,
    file_name              TEXT        NOT NULL UNIQUE,
    processed_at           TIMESTAMPTZ NOT NULL,
    processed_successfully BOOLEAN     NOT NULL,
    error_message          TEXT        NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_processed_files_file_name
    ON processed_files (file_name);

CREATE INDEX IF NOT EXISTS idx_file_data_unit_guid
    ON file_data (unit_guid);
