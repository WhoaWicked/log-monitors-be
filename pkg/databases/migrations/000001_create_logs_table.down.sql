BEGIN;

DROP INDEX IF EXISTS idx_logs_service;
DROP INDEX IF EXISTS idx_logs_level;
DROP INDEX IF EXISTS idx_logs_timestamp;
DROP TABLE logs;

COMMIT;