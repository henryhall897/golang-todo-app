-- Migration to update the `title` column in the `tasks` table to `TEXT`

-- Step 1: Change the data type of `title` from VARCHAR(255) to TEXT
ALTER TABLE tasks
ALTER COLUMN title TYPE TEXT;

-- Step 2: Ensure the `title` column is NOT NULL to maintain data integrity
ALTER TABLE tasks
ALTER COLUMN title SET NOT NULL;
