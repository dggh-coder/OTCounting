CREATE SCHEMA IF NOT EXISTS staffinfo;
CREATE SCHEMA IF NOT EXISTS otdriverstd;

CREATE TABLE IF NOT EXISTS staffinfo.staffinfo (
    id           BIGSERIAL PRIMARY KEY,
    staffid      VARCHAR(64) UNIQUE NOT NULL,
    nameeng      VARCHAR(255) NOT NULL,
    namechi      VARCHAR(255) NOT NULL,
    displayname  VARCHAR(255) NOT NULL,
    domainname   VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS otdriverstd.otperiod (
    id         BIGSERIAL PRIMARY KEY,
    date       DATE NOT NULL,
    otstaffid  VARCHAR(64) NOT NULL,
    period     CHAR(2) NOT NULL CHECK (period IN ('00', '01', '02')),
    remarks    VARCHAR(600),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_otperiod_staff_date_period UNIQUE (otstaffid, date, period)
);

CREATE TABLE IF NOT EXISTS otdriverstd.otdetails (
    id         BIGSERIAL PRIMARY KEY,
    otid       BIGINT NOT NULL REFERENCES otdriverstd.otperiod(id) ON DELETE CASCADE,
    type       CHAR(2) NOT NULL CHECK (type IN ('00', '01')),
    starttime  TIME NOT NULL,
    endtime    TIME NOT NULL,
    inputby    VARCHAR(64),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_otdetails_otid ON otdriverstd.otdetails (otid);

CREATE TABLE IF NOT EXISTS otdriverstd.periodresult (
    id          VARCHAR(10) PRIMARY KEY,
    otstaffid   VARCHAR(64) NOT NULL,
    date_label  DATE NOT NULL,
    process20txt TEXT NOT NULL,
    process15txt TEXT NOT NULL,
    hours20     INTEGER NOT NULL DEFAULT 0,
    hours15     INTEGER NOT NULL DEFAULT 0,
    mins20      INTEGER NOT NULL DEFAULT 0,
    mins15      INTEGER NOT NULL DEFAULT 0,
    totalhrs20  INTEGER NOT NULL DEFAULT 0,
    totalhrs15  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_periodresult_staff_date ON otdriverstd.periodresult (otstaffid, date_label);
