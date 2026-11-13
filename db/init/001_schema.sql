-- Invoicely Lite schema

CREATE TABLE sellers (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    trn         TEXT        NOT NULL UNIQUE,
    peppol_id   TEXT        NOT NULL UNIQUE,
    emirate     TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE imports (
    id             UUID PRIMARY KEY,
    seller_id      BIGINT      NOT NULL REFERENCES sellers(id),
    filename       TEXT        NOT NULL,
    status         TEXT        NOT NULL,           -- Uploaded | Processing | Completed | Failed
    error          TEXT        NOT NULL DEFAULT '',
    row_count      INT         NOT NULL DEFAULT 0,
    invoice_count  INT         NOT NULL DEFAULT 0,
    duration_ms    BIGINT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at   TIMESTAMPTZ
);
CREATE INDEX imports_created_at_idx ON imports (created_at DESC);

CREATE TABLE invoices (
    id              UUID PRIMARY KEY,
    import_id       UUID          NOT NULL REFERENCES imports(id),
    seller_id       BIGINT        NOT NULL REFERENCES sellers(id),
    invoice_number  TEXT          NOT NULL,
    buyer_name      TEXT          NOT NULL DEFAULT '',
    issue_date      TEXT          NOT NULL DEFAULT '',
    currency        TEXT          NOT NULL DEFAULT '',
    total_excl_vat  NUMERIC(14,2) NOT NULL DEFAULT 0,
    total_vat       NUMERIC(14,2) NOT NULL DEFAULT 0,
    total_incl_vat  NUMERIC(14,2) NOT NULL DEFAULT 0,
    status          TEXT          NOT NULL,         -- Parsed | Validated | Submitted | Accepted | Rejected | Failed
    rejected_by     TEXT          NOT NULL DEFAULT '', -- portal | asp
    errors          JSONB         NOT NULL DEFAULT '[]',
    payload         JSONB         NOT NULL,
    asp_message_id  TEXT          NOT NULL DEFAULT '',
    asp_response    JSONB,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX invoices_import_idx ON invoices (import_id);
CREATE INDEX invoices_seller_number_idx ON invoices (seller_id, invoice_number);
CREATE INDEX invoices_status_idx ON invoices (status);

CREATE TABLE invoice_events (
    id          BIGSERIAL PRIMARY KEY,
    invoice_id  UUID        NOT NULL REFERENCES invoices(id),
    status      TEXT        NOT NULL,
    detail      TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX invoice_events_invoice_idx ON invoice_events (invoice_id, id);
