-- +goose Up
-- +goose StatementBegin
CREATE TYPE messenger AS ENUM ('telegram');

CREATE TABLE account
(
    id           serial PRIMARY KEY,
    messenger    messenger NOT NULL,
    messenger_id text      NOT NULL
);

CREATE TABLE portfolio
(
    id         serial PRIMARY KEY,
    label      text NOT NULL,
    account_id int REFERENCES account (id)
);

CREATE INDEX portfolio_account_id_idx ON portfolio (account_id);

CREATE TABLE asset
(
    id   serial PRIMARY KEY,
    code text NOT NULL UNIQUE
);

CREATE TABLE portfolio_position
(
    id             serial PRIMARY KEY,
    portfolio_id   int REFERENCES portfolio (id),
    asset_id       int REFERENCES asset (id),
    quantity       int       NOT NULL CHECK (quantity > 0),
    placement_time timestamp NOT NULL
);

CREATE INDEX portfolio_position_portfolio_id_idx ON portfolio_position (portfolio_id);

CREATE TABLE portfolio_value_history
(
    portfolio_id     int REFERENCES portfolio (id),
    "value"          numeric(1000, 2) NOT NULL,
    calculation_time timestamp        NOT NULL
);

CREATE INDEX portfolio_value_history_portfolio_id_idx ON portfolio_value_history (portfolio_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS portfolio_value_history_portfolio_id_idx;
DROP TABLE IF EXISTS portfolio_value_history;
DROP INDEX IF EXISTS portfolio_position_portfolio_id_idx;
DROP TABLE IF EXISTS portfolio_position;
DROP TABLE IF EXISTS asset;
DROP INDEX IF EXISTS portfolio_account_id_idx;
DROP TABLE IF EXISTS portfolio;
DROP TABLE IF EXISTS account;
DROP TYPE IF EXISTS messenger;
-- +goose StatementEnd
