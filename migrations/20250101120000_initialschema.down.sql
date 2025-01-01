-- Drop the tasks table if it exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'tasks') THEN
        DROP TABLE tasks;
    END IF;
END $$;

-- Drop the todo_lists table if it exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'todo_lists') THEN
        DROP TABLE todo_lists;
    END IF;
END $$;

-- Drop the users table if it exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users') THEN
        DROP TABLE users;
    END IF;
END $$;

-- Drop the pgcrypto extension if it exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto') THEN
        DROP EXTENSION "pgcrypto";
    END IF;
END $$;
