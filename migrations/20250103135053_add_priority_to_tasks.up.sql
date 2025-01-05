-- Add a priority column to the tasks table
ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 0;

