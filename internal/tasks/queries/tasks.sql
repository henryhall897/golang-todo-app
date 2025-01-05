-- name: CreateTask :one
INSERT INTO tasks (list_id, title, task_desc, status, due_date, priority)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET title = COALESCE($3, title),
    task_desc = COALESCE($4, task_desc),
    status = COALESCE($5, status),
    due_date = COALESCE($6, due_date),
    priority = COALESCE($7, priority),
    updated_at = CURRENT_TIMESTAMP,
    completed_at = COALESCE($8, completed_at)
FROM todo_lists
WHERE tasks.id = $1
  AND tasks.list_id = todo_lists.id
  AND todo_lists.user_id = $2
RETURNING tasks.*;


-- name: DeleteTasks :many
DELETE FROM tasks
USING todo_lists
WHERE tasks.id = ANY($1::uuid[])
  AND tasks.list_id = todo_lists.id
  AND todo_lists.user_id = $2
RETURNING tasks.*;

-- name: ListTasks :many
SELECT tasks.*
FROM tasks
JOIN todo_lists ON tasks.list_id = todo_lists.id
WHERE todo_lists.id = $1
  AND todo_lists.user_id = $2
ORDER BY tasks.priority ASC, tasks.due_date ASC;

-- name: MarkTaskCompleted :exec
UPDATE tasks
SET 
    status = 'completed',                -- Set status to 'completed'
    priority = NULL,                      -- Set priority to NULL when task is completed
    completed_at = CURRENT_TIMESTAMP,    -- Set completed_at to current timestamp
    updated_at = CURRENT_TIMESTAMP       -- Update the timestamp for the task
FROM todo_lists
WHERE 
    tasks.id = $1
    AND tasks.list_id = todo_lists.id
    AND todo_lists.user_id = $2;         -- Reference user_id from todo_lists table

-- name: ListOverdueTasks :many
SELECT tasks.*
FROM tasks
JOIN todo_lists ON tasks.list_id = todo_lists.id
WHERE tasks.list_id = $1
  AND todo_lists.user_id = $2
  AND tasks.due_date < CURRENT_TIMESTAMP
  AND tasks.status != 'completed'
ORDER BY tasks.due_date ASC;

-- name: ListTasksByStatus :many
SELECT tasks.*
FROM tasks
JOIN todo_lists ON tasks.list_id = todo_lists.id
WHERE tasks.list_id = $1
  AND todo_lists.user_id = $2
  AND tasks.status = $3;

-- name: SearchTasks :many
SELECT tasks.*
FROM tasks
JOIN todo_lists ON tasks.list_id = todo_lists.id
WHERE tasks.list_id = $1
  AND todo_lists.user_id = $2
  AND (tasks.title ILIKE '%' || $3 || '%' OR tasks.task_desc ILIKE '%' || $3 || '%')
ORDER BY tasks.priority ASC NULLS LAST, tasks.due_date ASC;

-- name: UpdateTaskPriority :exec
WITH RankedTasks AS (
    SELECT 
        tasks.id,
        ROW_NUMBER() OVER (ORDER BY tasks.priority DESC, tasks.due_date ASC) AS new_priority
    FROM tasks
    WHERE tasks.list_id = $1
)
UPDATE tasks
SET 
    priority = CASE 
        WHEN tasks.id = $2 THEN $3  -- Set the new priority for the specific task
        ELSE RankedTasks.new_priority -- Reorder the other tasks
    END,
    updated_at = CURRENT_TIMESTAMP
FROM RankedTasks
WHERE tasks.id = RankedTasks.id
;



