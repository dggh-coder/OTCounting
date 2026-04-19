CREATE SCHEMA IF NOT EXISTS ot_uat;
SET search_path TO ot_uat, public;

CREATE TABLE IF NOT EXISTS work_session (
    session_id       BIGINT PRIMARY KEY,
    session_date     DATE NOT NULL,
    period           CHAR(2) NOT NULL CHECK (period IN ('AM','PM')),
    status           VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    created_by       VARCHAR(64),
    note             TEXT,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_work_session_date_period
    ON work_session (session_date, period);

CREATE TABLE IF NOT EXISTS time_entry (
    id               UUID PRIMARY KEY,
    session_id       BIGINT NOT NULL,
    employee_id      CHAR(1) NOT NULL CHECK (employee_id IN ('A','B')),
    entry_type       VARCHAR(5) NOT NULL CHECK (entry_type IN ('OT','BREAK')),
    start_time       TIME NOT NULL,
    end_time         TIME NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_time_entry_session
        FOREIGN KEY (session_id)
        REFERENCES work_session(session_id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_time_entry_session_emp
    ON time_entry (session_id, employee_id);

CREATE INDEX IF NOT EXISTS idx_time_entry_session_type
    ON time_entry (session_id, entry_type);

CREATE TABLE IF NOT EXISTS session_result (
    session_id               BIGINT NOT NULL,
    employee_id              CHAR(1) NOT NULL CHECK (employee_id IN ('A','B')),
    date_label               VARCHAR(64) NOT NULL,
    rate20_rounded_hours     INTEGER NOT NULL DEFAULT 0,
    rate15_rounded_hours     INTEGER NOT NULL DEFAULT 0,
    rate20_minutes           INTEGER NOT NULL DEFAULT 0,
    rate15_minutes           INTEGER NOT NULL DEFAULT 0,
    calculated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT pk_session_result
        PRIMARY KEY (session_id, employee_id),

    CONSTRAINT fk_session_result_session
        FOREIGN KEY (session_id)
        REFERENCES work_session(session_id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_session_result_session
    ON session_result (session_id);
