-- Remove the priority column from the tasks table
ALTER TABLE tasks
DROP COLUMN IF EXISTS priority;
