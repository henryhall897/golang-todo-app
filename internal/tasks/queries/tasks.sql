-- name: CreateTask :one
INSERT INTO tasks (list_id, title, description, status, due_date, priority)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET title = COALESCE($3, tasks.title),
    description = COALESCE($4, tasks.description),
    status = COALESCE($5, status),
    due_date = COALESCE($6, due_date),
    priority = COALESCE($7, priority),
    updated_at = CURRENT_TIMESTAMP,
    completed_at = COALESCE($8, completed_at)
FROM todolists
WHERE tasks.id = $1
  AND tasks.list_id = todolists.id
  AND todolists.user_id = $2
RETURNING tasks.*;

-- name: DeleteTasks :many
DELETE FROM tasks
USING todolists
WHERE tasks.id = ANY($1::uuid[])
  AND tasks.list_id = todolists.id
  AND todolists.user_id = $2
RETURNING tasks.*;

-- name: ListTasks :many
SELECT tasks.*
FROM tasks
JOIN todolists ON tasks.list_id = todolists.id
WHERE todolists.id = $1
  AND todolists.user_id = $2
ORDER BY tasks.priority ASC, tasks.due_date ASC;

-- name: MarkTaskCompleted :exec
UPDATE tasks
SET 
    status = 'completed',               
    priority = NULL,                      
    completed_at = CURRENT_TIMESTAMP,    
    updated_at = CURRENT_TIMESTAMP       
FROM todolists
WHERE 
    tasks.id = $1
    AND tasks.list_id = todolists.id
    AND todolists.user_id = $2;         

-- name: ListOverdueTasks :many
SELECT tasks.*
FROM tasks
JOIN todolists ON tasks.list_id = todolists.id
WHERE tasks.list_id = $1
  AND todolists.user_id = $2
  AND tasks.due_date < CURRENT_TIMESTAMP
  AND tasks.status != 'completed'
ORDER BY tasks.due_date ASC;

-- name: ListTasksByStatus :many
SELECT tasks.*
FROM tasks
JOIN todolists ON tasks.list_id = todolists.id
WHERE tasks.list_id = $1
  AND todolists.user_id = $2
  AND tasks.status = $3;

-- name: SearchTasks :many
SELECT tasks.*
FROM tasks
JOIN todolists ON tasks.list_id = todolists.id
WHERE tasks.list_id = $1
  AND todolists.user_id = $2
  AND (tasks.title ILIKE '%' || $3 || '%' OR tasks.description ILIKE '%' || $3 || '%')
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
WHERE tasks.id = RankedTasks.id;
