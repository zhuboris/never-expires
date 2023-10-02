CREATE TABLE IF NOT EXISTS shared_types_of_items (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS private_types_of_items (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID NOT NULL
);

CREATE UNIQUE INDEX idx_lower_name_shared_types_of_items ON shared_types_of_items (LOWER(name));
CREATE UNIQUE INDEX idx_lower_name_private_types_of_items ON private_types_of_items (LOWER(name), user_id);

CREATE TABLE IF NOT EXISTS storages (
    id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    owner_id UUID NOT NULL,

    UNIQUE (name, owner_id),
    PRIMARY KEY (owner_id, id)
);

CREATE TABLE IF NOT EXISTS users_default_storages (
    user_id UUID NOT NULL,
    storage_id UUID NOT NULL,

    PRIMARY KEY (user_id, storage_id),
    CONSTRAINT storage_fk FOREIGN KEY (user_id, storage_id) REFERENCES storages(owner_id, id)
);

CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storage_id UUID NOT NULL,

    CONSTRAINT storage_fk FOREIGN KEY (storage_id) REFERENCES storages(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS items_info (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    is_opened BOOLEAN NOT NULL DEFAULT FALSE,
    best_before TIMESTAMPTZ NOT NULL,
    expiration_date TIMESTAMPTZ,
    hours_after_opening INTEGER NOT NULL DEFAULT 0,
    added_date TIMESTAMPTZ NOT NULL DEFAULT now(),
    note TEXT,

    CONSTRAINT id_fk FOREIGN KEY (id) REFERENCES items(id) ON DELETE CASCADE
);

CREATE INDEX idx_added_date ON items_info (added_date);

CREATE TABLE IF NOT EXISTS ios_devices (
    token TEXT PRIMARY KEY,
    user_id UUID NOT NULL
);

CREATE OR REPLACE FUNCTION set_expiration_date()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.expiration_date IS NULL THEN
        NEW.expiration_date := NEW.best_before;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER add_item_trigger
    BEFORE INSERT ON items_info
    FOR EACH ROW
EXECUTE FUNCTION set_expiration_date();