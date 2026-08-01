BEGIN;
INSERT INTO users (username, email, password_hash)
VALUES ('somkiad', 'test@gmail.com', '$2a$14$gxWtRVIcHr75BWnjj7vPqufXLaRwoAWtwbe6Aa7xasOGgPMXbFO/u');
COMMIT;