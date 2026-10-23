-- Откат 000002: возвращаем email в NOT NULL и общий уникальный индекс.
-- Перед откатом нужно заполнить пустые email, иначе ALTER COLUMN ... SET NOT NULL
-- упадёт на существующих NULL.

UPDATE "user" SET email = 'unknown@invalid.local' WHERE email IS NULL;

ALTER TABLE "user" ALTER COLUMN email SET NOT NULL;

DROP INDEX idx_user_email_lower;

CREATE UNIQUE INDEX idx_user_email_lower ON "user" (lower(email));
