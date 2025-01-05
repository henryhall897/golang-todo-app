-- 20250102120000_rename_description_columns.down.sql

-- Revert todo_desc back to description in todo_lists table
ALTER TABLE todo_lists
RENAME COLUMN todo_desc TO description;

-- Revert task_desc back to description in tasks table
ALTER TABLE tasks
RENAME COLUMN task_desc TO description;
