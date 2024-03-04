CREATE TABLE relays (
    id SERIAL PRIMARY KEY,
    uri character varying NOT NULL UNIQUE,
    relay_name character varying NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone
);
--bun:split
INSERT INTO relays(id, uri, relay_name) SELECT 1, 'wss://dev-relay.nostrassets.com', 'internal-dev' WHERE NOT EXISTS (SELECT id FROM relays WHERE id = 1);
--bun:split

CREATE TABLE filters (
    relay_id bigint PRIMARY KEY,
    last_event_seen bigint DEFAULT extract(epoch from now()) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone,
    CONSTRAINT fk_relay
        FOREIGN KEY(relay_id)
        REFERENCES relays(id)
        ON DELETE CASCADE
);
--bun:split
INSERT INTO filters(relay_id) SELECT 1 WHERE NOT EXISTS (SELECT relay_id FROM filters WHERE relay_id = 1);
--bun:split
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    pubkey character varying NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone
);
--bun:split
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    event_id character varying NOT NULL UNIQUE,
    from_pubkey character varying NOT NULL,
    kind int NOT NULL,
    content character varying NOT NULL,
    created_at int NOT NULL
)
--bun:split
CREATE TABLE assets (
    id SERIAL PRIMARY KEY,
    ta_asset_id character varying NOT NULL UNIQUE,
    asset_name character varying NOT NULL UNIQUE,
    asset_type int DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone
);
--bun:split
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    user_id bigint NOT NULL,
    ta_asset_id character varying NOT NULL,
    type character varying NOT NULL,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_asset
        FOREIGN KEY(ta_asset_id)
        REFERENCES assets(ta_asset_id)
        ON DELETE NO ACTION
);
--bun:split
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    type character varying,
    user_id bigint,
    ta_asset_id character varying NOT NULL,
    amount bigint,
    memo character varying,
    description_hash character varying,
    payment_request character varying,
    destination_pubkey_hex character varying NOT NULL,
    r_hash character varying,
    preimage character varying,
    internal boolean,
    state character varying DEFAULT 'initialized',
    error_message character varying,
    add_index bigint,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at timestamp with time zone,
    updated_at timestamp with time zone,
    settled_at timestamp with time zone,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_asset
        FOREIGN KEY(ta_asset_id)
        REFERENCES assets(ta_asset_id)
        ON DELETE NO ACTION
);
--bun:split
CREATE TABLE addresses (
    id SERIAL PRIMARY KEY,
    user_id bigint NOT NULL,
    ta_asset_id character varying NOT NULL,
    amount bigint,
    addr character varying NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_asset
        FOREIGN KEY(ta_asset_id)
        REFERENCES assets(ta_asset_id)
        ON DELETE NO ACTION
);
--bun:split
CREATE TABLE transaction_entries (
    id SERIAL PRIMARY KEY,
    user_id bigint NOT NULL,
    invoice_id bigint,
    parent_id bigint,
    addr character varying,
    credit_account_id bigint NOT NULL,
    debit_account_id bigint NOT NULL,
    amount bigint NOT NULL,
    outpoint character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_credit_account
        FOREIGN KEY(credit_account_id)
        REFERENCES accounts(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_debit_account
        FOREIGN KEY(debit_account_id)
        REFERENCES accounts(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_addr
        FOREIGN KEY(addr)
        REFERENCES addresses(addr)
        ON DELETE NO ACTION
);
--bun:split
