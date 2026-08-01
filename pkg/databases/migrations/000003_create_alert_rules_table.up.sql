BEGIN;

CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service VARCHAR(255) NOT NULL,
    level VARCHAR(50) NOT NULL,
    threshold INT NOT NULL,
    window_seconds INT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;