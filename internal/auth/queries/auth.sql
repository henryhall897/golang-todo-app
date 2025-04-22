-- name: CreateAuthIdentity :one
INSERT INTO auth_identities (auth_id, provider, user_id, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAuthIdentityByAuthID :one
SELECT * FROM auth_identities
WHERE auth_id = $1;

-- name: GetAuthIdentityByUserID :one
SELECT * FROM auth_identities
WHERE user_id = $1;

-- name: UpdateAuthIdentityRole :exec
UPDATE auth_identities
SET role = $2, updated_at = NOW()
WHERE auth_id = $1;

-- name: DeleteAuthIdentityByAuthID :exec
DELETE FROM auth_identities
WHERE auth_id = $1;
