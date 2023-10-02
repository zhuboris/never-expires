CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS emails(
    email VARCHAR PRIMARY KEY,
    owner_id  UUID NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_confirmed BOOLEAN DEFAULT FALSE,

    CONSTRAINT id_fk FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS passwords(
    user_id UUID PRIMARY KEY,
    encrypted_password VARCHAR(100) NOT NULL,

    CONSTRAINT id_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sessions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    device TEXT DEFAULT 'unknown device',
    refresh_jwt TEXT UNIQUE NOT NULL,
    start_time timestamptz DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,

    CONSTRAINT id_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS mail_confirmation_tokens(
    token VARCHAR PRIMARY KEY,
    email VARCHAR NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    expiration timestamptz NOT NULL,

    CONSTRAINT email_fk FOREIGN KEY (email) REFERENCES emails(email) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS google_ids(
    user_id UUID PRIMARY KEY,
    id TEXT UNIQUE NOT NULL,

    CONSTRAINT id_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS apple_ids(
    user_id UUID PRIMARY KEY,
    id TEXT UNIQUE NOT NULL,

    CONSTRAINT id_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS apple_refresh_tokens (
    token TEXT PRIMARY KEY,
    user_id UUID NOT NULL,

    CONSTRAINT id_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS password_restoration_tokens(
     token VARCHAR PRIMARY KEY,
     user_email VARCHAR NOT NULL,
     is_used BOOLEAN DEFAULT FALSE,
     expiration timestamptz NOT NULL,

     CONSTRAINT email_fk FOREIGN KEY (user_email) REFERENCES emails(email) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users_to_delete (
    id UUID PRIMARY KEY
);

CREATE OR REPLACE FUNCTION disable_users_emails()
    RETURNS TRIGGER AS $$
BEGIN
    UPDATE emails
    SET is_active = FALSE
    WHERE owner_id = NEW.owner_id
    AND is_active = TRUE;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER disable_previous_users_email
BEFORE INSERT ON emails
FOR EACH ROW
EXECUTE FUNCTION disable_users_emails();

CREATE UNIQUE INDEX active_email_per_user
ON emails (owner_id)
WHERE is_active;