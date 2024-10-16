-- +goose Up

CREATE TABLE subs(
    channel TEXT NOT NULL,
    chat TEXT NOT NULL,
    UNIQUE(channel, chat)
);

-- +goose Down

DROP TABLE subs;
