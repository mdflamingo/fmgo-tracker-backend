-- ==============================================================================
-- EMAIL ПОЛЬЗОВАТЕЛЯ: NOT NULL -> NULL
-- ==============================================================================
-- Сервис-аккаунты (client_credentials) и часть обычных пользователей в Keycloak
-- не имеют email. При NOT NULL их невозможно было бы завести в "user", а
-- уникальный индекс по lower(email) делал коллизию lower('') для двоих без почты.

ALTER TABLE "user" ALTER COLUMN email DROP NOT NULL;

DROP INDEX idx_user_email_lower;

-- Частичный уникальный индекс: несколько NULL допустимы, дубли непустых адресов — нет
CREATE UNIQUE INDEX idx_user_email_lower ON "user" (lower(email)) WHERE email IS NOT NULL;
