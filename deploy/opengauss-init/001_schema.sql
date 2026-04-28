CREATE SCHEMA IF NOT EXISTS staffinfo;
CREATE SCHEMA IF NOT EXISTS otdriverstd;

CREATE TABLE IF NOT EXISTS staffinfo.staffinfo (
    id           BIGSERIAL PRIMARY KEY,
    staffid      VARCHAR(64) UNIQUE NOT NULL,
    nameeng      VARCHAR(255),
    namechi      VARCHAR(255),
    displayname  VARCHAR(255),
    domainname   VARCHAR(255)
);

ALTER TABLE staffinfo.staffinfo ALTER COLUMN nameeng DROP NOT NULL;
ALTER TABLE staffinfo.staffinfo ALTER COLUMN namechi DROP NOT NULL;
ALTER TABLE staffinfo.staffinfo ALTER COLUMN displayname DROP NOT NULL;
ALTER TABLE staffinfo.staffinfo ALTER COLUMN domainname DROP NOT NULL;

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
    starttime  VARCHAR(16) NOT NULL,
    endtime    VARCHAR(16) NOT NULL,
    inputby    VARCHAR(64),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE otdriverstd.otdetails
  ALTER COLUMN starttime TYPE VARCHAR(16)
  USING COALESCE(regexp_substr(starttime::text, '[0-9]{2}:[0-9]{2}'), starttime::text);
ALTER TABLE otdriverstd.otdetails
  ALTER COLUMN endtime TYPE VARCHAR(16)
  USING COALESCE(regexp_substr(endtime::text, '[0-9]{2}:[0-9]{2}'), endtime::text);

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
