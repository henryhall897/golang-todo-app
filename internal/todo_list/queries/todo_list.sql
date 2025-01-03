-- Create a new todo list
-- name: CreateTodoList :one
INSERT INTO todo_lists (user_id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- Retrieve todo lists with pagination
-- name: ListTodoListsWithPagination :many
SELECT *
FROM todo_lists
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- Delete a single todo list by ID for a specific user
-- name: DeleteTodoList :execrows
DELETE FROM todo_lists
WHERE id = $1 AND user_id = $2;

-- Bulk delete todo lists for a specific user
-- name: BulkDeleteTodoLists :execrows
DELETE FROM todo_lists
WHERE id = ANY($1::uuid[]) AND user_id = $2;

-- Retrieve a todo list by ID, ensuring it belongs to the user
-- name: GetTodoListByID :one
SELECT *
FROM todo_lists
WHERE id = $1 AND user_id = $2;

-- Update an existing todo list for a specific user
-- name: UpdateTodoList :one
UPDATE todo_lists
SET name = $3, description = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING *;