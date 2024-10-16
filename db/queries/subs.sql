-- name: Subscribe :one
INSERT INTO subs(channel, chat)
VALUES ($1, $2)
    RETURNING *;

-- name: GetSubCnt :one
SELECT COUNT(*) FROM subs
WHERE chat = $1;

-- name: GetSubsOfChannel :many
SELECT chat FROM subs
WHERE channel = $1;
