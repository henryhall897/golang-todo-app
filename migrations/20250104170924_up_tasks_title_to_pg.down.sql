-- Rollback the `title` column to its previous state (VARCHAR(255))

-- Step 1: Change the data type of `title` from TEXT back to VARCHAR(255)
ALTER TABLE tasks
ALTER COLUMN title TYPE VARCHAR(255);

-- Step 2: Ensure the `title` column remains NOT NULL
ALTER TABLE tasks
ALTER COLUMN title SET NOT NULL;
