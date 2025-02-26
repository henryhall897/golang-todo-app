-- Create a new user
-- name: CreateUser :one
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING *;

-- Retrieve a user by ID
-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- Retrieve a user by email
-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- Update user details
-- name: UpdateUser :one
UPDATE users
SET 
    name = $2,
    email = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;



-- Delete a user by ID
-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = $1;

-- List all users with pagination
-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
