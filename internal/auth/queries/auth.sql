-- name: CreateAuthIdentity :one
INSERT INTO auth_identities (auth_id, provider, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAuthIdentityByAuthID :one
SELECT * FROM auth_identities
WHERE auth_id = $1;

-- name: GetAuthIdentitiesByUserID :many
SELECT * FROM auth_identities
WHERE user_id = $1;

-- name: DeleteAuthIdentityByAuthID :execrows
DELETE FROM auth_identities
WHERE auth_id = $1;

