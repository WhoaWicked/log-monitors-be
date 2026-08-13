BEGIN;

ALTER TABLE alert_rules ADD CONSTRAINT unique_service_level UNIQUE (service, level);

COMMIT;