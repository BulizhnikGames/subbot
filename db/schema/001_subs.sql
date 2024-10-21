-- +goose Up

CREATE TABLE subs(
    channel BIGINT NOT NULL,
    chat BIGINT NOT NULL,
    UNIQUE(channel, chat)
);

-- +goose Down

DROP TABLE subs;
