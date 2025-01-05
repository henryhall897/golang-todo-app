-- 20250102120000_rename_description_columns.up.sql

-- Rename description column in todo_lists table to todo_desc
ALTER TABLE todo_lists
RENAME COLUMN description TO todo_desc;

-- Rename description column in tasks table to task_desc
ALTER TABLE tasks
RENAME COLUMN description TO task_desc;
