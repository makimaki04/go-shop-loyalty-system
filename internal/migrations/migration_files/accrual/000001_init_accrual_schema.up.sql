CREATE TYPE accrual_status AS ENUM ('REGISTERED', 'INVALID', 'PROCESSING', 'PROCESSED');

CREATE TABLE accruals (
    id BIGSERIAL PRIMARY KEY,
    order_number TEXT NOT NULL UNIQUE CHECK(order_number ~ '^\d+$'),
    status accrual_status NOT NULL DEFAULT 'REGISTERED',
    accrual NUMERIC(20, 2),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    CONSTRAINT accrual_check_accrual_status
        CHECK (
            (status = 'PROCESSED' AND accrual IS NOT NULL AND accrual >= 0)
            OR 
            (status IN ('REGISTERED', 'INVALID', 'PROCESSING') AND accrual IS NULL)
        ),
    CONSTRAINT accrual_check_processed_at 
        CHECK (
            (status = 'PROCESSED' AND processed_at IS NOT NULL AND processed_at >= uploaded_at)
            OR
            (status IN ('REGISTERED', 'INVALID', 'PROCESSING') AND processed_at IS NULL)
        )
);

CREATE INDEX accrual_status_uploaded_idx ON accruals(status, uploaded_at DESC);