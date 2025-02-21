-- Create a new todo list
-- name: CreateTodoList :one
INSERT INTO todolists (user_id, title, description)
VALUES ($1, $2, $3)
RETURNING *;

-- Retrieve todo lists with pagination
-- name: ListTodoListsWithPagination :many
SELECT *
FROM todolists
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- Delete one, multiple, or all todo lists for a specific user
-- name: DeleteTodoLists :execrows
DELETE FROM todolists
WHERE user_id = $1
AND ($2::uuid[] IS NULL OR id = ANY($2::uuid[]));

-- Retrieve a todo list by ID, ensuring it belongs to the user
-- name: GetTodoListByID :one
SELECT *
FROM todolists
WHERE id = $1 AND user_id = $2;

-- Update an existing todo list for a specific user
-- name: UpdateTodoList :one
UPDATE todolists
SET 
    title = COALESCE($3, title),
    description = COALESCE($4, description),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING *;
