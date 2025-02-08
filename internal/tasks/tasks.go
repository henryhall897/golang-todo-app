package tasks

import (
	"context"
	"errors"
	"fmt"
	"golang-todo-app/internal/tasks/gen"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

// New initializes a new Store with the provided connection pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// CreateTask inserts a new task into the database and returns the created Task.
func (s *Store) CreateTask(ctx context.Context, lid uuid.UUID, title string, desc, status *string, due time.Time, prio int32) (FullTask, error) {
	query := gen.New(s.pool)

	params := CreateTaskParams{
		ListID:   lid,
		Title:    &title,
		TaskDesc: desc,
		Status:   status,
		DueDate:  &due,
		Priority: prio,
	}
	// Transform the Go struct to a database-compatible struct
	dbTask, err := toDBCreateTask(params)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to transform task: %w", err)
	}

	// Execute the query
	createdTask, err := query.CreateTask(ctx, dbTask)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to create task: %w", err)
	}

	// Convert the database result back to the application-compatible struct
	result, err := toFullTask(createdTask)
	if err != nil {
		return result, fmt.Errorf("failed to transform task from database: %w", err)
	}
	return result, nil
}

// UpdateTask updates an existing task in the database and returns the updated Task.
func (s *Store) UpdateTask(ctx context.Context, params UpdateTaskParams) (FullTask, error) {
	query := gen.New(s.pool)

	// Check if priority was updated
	if params.Priority != nil && *params.Priority != 0 {
		// Use the dedicated priority update function instead of general update
		return s.UpdateTaskPriority(ctx, params)
	}
	if params.Status != nil && *params.Status == "completed" {
		// If the status is being updated to "completed", call the specialized MarkTaskCompleted function
		return s.MarkTaskCompleted(ctx, params)
	}
	// Transform the Go struct to a database-compatible struct
	dbParams, err := toDBUpdateTaskParams(params)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to transform update task params: %w", err)
	}

	// Execute the query
	updatedTask, err := query.UpdateTask(ctx, dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FullTask{}, fmt.Errorf("no task found to update with the provided parameters: %w", err)
		}
		return FullTask{}, fmt.Errorf("failed to update task: %w", err)
	}

	// Convert the database result back to the application-compatible struct
	result, err := toFullTask(updatedTask)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to transform task from database: %w", err)
	}

	return result, nil
}

// DeleteTasks deletes one or more tasks from the database and returns the deleted tasks.
func (s *Store) DeleteTasks(ctx context.Context, params DeleteTasksParams) ([]FullTask, error) {
	query := gen.New(s.pool)

	// Transform the Go struct to a database-compatible struct
	dbParams, err := toDBDeleteTasksParams(params)
	if err != nil {
		return nil, fmt.Errorf("failed to transform delete tasks params: %w", err)
	}

	// Execute the query
	deletedTasks, err := query.DeleteTasks(ctx, dbParams)
	if err != nil {
		return nil, fmt.Errorf("failed to delete tasks: %w", err)
	}

	// Convert each deleted task to FullTask
	var results []FullTask
	for _, dbTask := range deletedTasks {
		task, err := toFullTask(dbTask)
		if err != nil {
			return nil, fmt.Errorf("failed to transform deleted task: %w", err)
		}
		results = append(results, task)
	}

	return results, nil
}

func (s *Store) ListTasks(ctx context.Context, params TaskListParams) ([]FullTask, error) {
	query := gen.New(s.pool)

	// Transform params to DB params
	dbParams, err := toDBListTasksParams(params)
	if err != nil {
		return nil, err
	}

	// Execute the query
	dbTasks, err := query.ListTasks(ctx, dbParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Convert the results to FullTask
	return toFullTaskList(dbTasks)
}

// ListOverdueTasks retrieves all overdue tasks for a specific todo list and user.
func (s *Store) ListOverdueTasks(ctx context.Context, params TaskListParams) ([]FullTask, error) {
	query := gen.New(s.pool)

	// Transform params to DB params
	dbParams, err := toDBListOverdueTasksParams(params)
	if err != nil {
		return nil, err
	}

	// Execute the query
	dbTasks, err := query.ListOverdueTasks(ctx, dbParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list overdue tasks: %w", err)
	}

	// Convert the results to FullTask
	return toFullTaskList(dbTasks)
}

// MarkTaskCompleted updates a task's status to 'completed' and handles priority and completed_at updates within a transaction.
func (s *Store) MarkTaskCompleted(ctx context.Context, params UpdateTaskParams) (FullTask, error) {
	// Start a new transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p) // Re-panic after rollback
		} else if err != nil {
			tx.Rollback(ctx)
		}
	}()

	query := gen.New(tx)

	// Step 1: Update task status to 'completed' and set priority and completed_at
	dbParams, err := toDBUpdateTaskParams(params)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to transform update task params: %w", err)
	}

	updateParams, err := toMarkTaskCompletedParams(dbParams)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to transform mark task completed params: %w", err)
	}

	err = query.MarkTaskCompleted(ctx, updateParams)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to mark task as completed: %w", err)
	}

	// Step 2: Perform a general update for other fields (title, description, etc.)
	params.Priority = nil    // Exclude priority from general update
	params.Status = nil      // Exclude status from general update
	params.CompletedAt = nil // Exclude completed_at from general update

	genUpdateParams, err := toDBUpdateTaskParams(params)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to transform update task params: %w", err)
	}

	genTask, err := query.UpdateTask(ctx, genUpdateParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FullTask{}, fmt.Errorf("no task found to update with the provided parameters: %w", err)
		}
		return FullTask{}, fmt.Errorf("failed to update task: %w", err)
	}

	// Step 3: Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return FullTask{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Step 4: Convert the database result back to the application-compatible struct
	result, err := toFullTask(genTask)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to transform task from database: %w", err)
	}

	return result, nil
}

// ListTasksByStatus retrieves all tasks for a specific list and user with a given status.
func (s *Store) ListTasksByStatus(ctx context.Context, params CountTasksByStatusParams) ([]FullTask, error) {
	query := gen.New(s.pool)

	// Transform the Go struct to a database-compatible struct
	dbParams, err := toDBListTasksByStatusParams(params)
	if err != nil {
		return nil, fmt.Errorf("failed to transform count tasks by status params: %w", err)
	}

	// Execute the query
	dbTasks, err := query.ListTasksByStatus(ctx, dbParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks by status: %w", err)
	}

	// Convert the database result to FullTask using toFullTaskList
	fullTasks, err := toFullTaskList(dbTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to transform tasks from database: %w", err)
	}

	return fullTasks, nil
}

// SearchTasks retrieves tasks based on the provided search parameters.
func (s *Store) SearchTasks(ctx context.Context, params SearchTasksParams) ([]FullTask, error) {
	query := gen.New(s.pool)

	// Transform the Go struct (SearchTasksParams) into a database-compatible struct (SearchTasksParams)
	dbParams, err := toDBSearchTasksParams(params)
	if err != nil {
		return nil, fmt.Errorf("failed to transform search tasks params: %w", err)
	}

	// Execute the query
	dbTasks, err := query.SearchTasks(ctx, dbParams)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	// Convert the database result to FullTask using toFullTaskList
	fullTasks, err := toFullTaskList(dbTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to transform tasks from database: %w", err)
	}

	return fullTasks, nil
}

// UpdateTaskPriority handles updating priority and reordering tasks based on priority with transaction support.
func (s *Store) UpdateTaskPriority(ctx context.Context, params UpdateTaskParams) (FullTask, error) {
	// Start a new transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p) // Re-panic after rollback
		} else if err != nil {
			tx.Rollback(ctx)
		}
	}()

	query := gen.New(tx)

	// Step 1: Update the task priority
	updatePrioParams := toDBUpdatePriorityParams(params)
	err = query.UpdateTaskPriority(ctx, updatePrioParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FullTask{}, fmt.Errorf("task not found for priority update: %w", err)
		}
		return FullTask{}, fmt.Errorf("failed to update task priority: %w", err)
	}

	// Step 2: Update other fields except priority
	paramsWithoutPriority := params
	paramsWithoutPriority.Priority = nil
	noPrioParams, err := toDBUpdateTaskParams(paramsWithoutPriority)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to transform update task params: %w", err)
	}

	updatedTaskGeneral, err := query.UpdateTask(ctx, noPrioParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FullTask{}, fmt.Errorf("no task found to update with the provided parameters: %w", err)
		}
		return FullTask{}, fmt.Errorf("failed to update task: %w", err)
	}

	// Step 3: Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return FullTask{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Step 4: Convert the updated task to the application-compatible format
	result, err := toFullTask(updatedTaskGeneral)
	if err != nil {
		return FullTask{}, fmt.Errorf("failed to transform task from database: %w", err)
	}

	return result, nil
}
