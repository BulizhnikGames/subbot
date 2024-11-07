// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: subs.sql

package orm

import (
	"context"
)

const checkSubscription = `-- name: CheckSubscription :one
SELECT COUNT(1) FROM subs
WHERE chat = $1 AND channel = $2
`

type CheckSubscriptionParams struct {
	Chat    int64
	Channel int64
}

func (q *Queries) CheckSubscription(ctx context.Context, arg CheckSubscriptionParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, checkSubscription, arg.Chat, arg.Channel)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getSubsOfChannel = `-- name: GetSubsOfChannel :many
SELECT chat FROM subs
WHERE channel = $1
`

func (q *Queries) GetSubsOfChannel(ctx context.Context, channel int64) ([]int64, error) {
	rows, err := q.db.QueryContext(ctx, getSubsOfChannel, channel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int64
	for rows.Next() {
		var chat int64
		if err := rows.Scan(&chat); err != nil {
			return nil, err
		}
		items = append(items, chat)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listGroupSubs = `-- name: ListGroupSubs :many
SELECT channel FROM subs
WHERE chat = $1
`

func (q *Queries) ListGroupSubs(ctx context.Context, chat int64) ([]int64, error) {
	rows, err := q.db.QueryContext(ctx, listGroupSubs, chat)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int64
	for rows.Next() {
		var channel int64
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		items = append(items, channel)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const subscribe = `-- name: Subscribe :one
INSERT INTO subs(chat, channel)
VALUES ($1, $2)
RETURNING chat, channel
`

type SubscribeParams struct {
	Chat    int64
	Channel int64
}

func (q *Queries) Subscribe(ctx context.Context, arg SubscribeParams) (Sub, error) {
	row := q.db.QueryRowContext(ctx, subscribe, arg.Chat, arg.Channel)
	var i Sub
	err := row.Scan(&i.Chat, &i.Channel)
	return i, err
}

const unSubscribe = `-- name: UnSubscribe :exec
DELETE FROM subs
WHERE chat = $1 AND channel = $2
`

type UnSubscribeParams struct {
	Chat    int64
	Channel int64
}

func (q *Queries) UnSubscribe(ctx context.Context, arg UnSubscribeParams) error {
	_, err := q.db.ExecContext(ctx, unSubscribe, arg.Chat, arg.Channel)
	return err
}
