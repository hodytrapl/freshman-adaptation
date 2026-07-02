package repository

import (
	"context"
	"database/sql"
	"errors"
	"freshman-adaptation/internal/apperrors"
	"freshman-adaptation/internal/models"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateGroup(ctx context.Context, name string) (*models.Group, error) {
	var group models.Group
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO groups (name) VALUES ($1) RETURNING id, name`,
		name,
	).Scan(&group.ID, &group.Name)
	if err != nil {
		return nil, apperrors.Internal("internal server error")
	}
	return &group, nil
}

func (r *Repository) ListGroups(ctx context.Context) ([]models.Group, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM groups ORDER BY id`)
	if err != nil {
		return nil, apperrors.Internal("internal server error")
	}
	defer rows.Close()

	groups := make([]models.Group, 0)
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.Name); err != nil {
			return nil, apperrors.Internal("internal server error")
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Internal("internal server error")
	}
	return groups, nil
}

func (r *Repository) GroupExists(ctx context.Context, groupID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM groups WHERE id = $1)`,
		groupID,
	).Scan(&exists)
	if err != nil {
		return false, apperrors.Internal("internal server error")
	}
	return exists, nil
}

func (r *Repository) CreateStudentWithDefaultTasks(
	ctx context.Context,
	firstName, lastName string,
	groupID int64,
) (*models.Student, error) {
	exists, err := r.GroupExists(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.NotFound("group not found")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperrors.Internal("internal server error")
	}
	defer tx.Rollback()

	var student models.Student
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO students (first_name, last_name, group_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, first_name, last_name, group_id, created_at`,
		firstName,
		lastName,
		groupID,
	).Scan(&student.ID, &student.FirstName, &student.LastName, &student.GroupID, &student.CreatedAt)
	if err != nil {
		return nil, apperrors.Internal("internal server error")
	}

	if err := createDefaultTasks(ctx, tx, student.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, apperrors.Internal("internal server error")
	}

	return &student, nil
}

func createDefaultTasks(ctx context.Context, tx *sql.Tx, studentID int64) error {
	defaultTitles := []string{
		"Связаться с куратором",
		"Пройти профессиональное тестирование",
	}

	for _, title := range defaultTitles {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO tasks (student_id, title, status) VALUES ($1, $2, $3)`,
			studentID,
			title,
			models.TaskStatusTodo,
		)
		if err != nil {
			return apperrors.Internal("internal server error")
		}
	}

	return nil
}

func (r *Repository) ListStudentsByGroup(ctx context.Context, groupID int64) ([]models.Student, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, first_name, last_name, group_id, created_at
		 FROM students
		 WHERE group_id = $1
		 ORDER BY id`,
		groupID,
	)
	if err != nil {
		return nil, apperrors.Internal("internal server error")
	}
	defer rows.Close()

	students := make([]models.Student, 0)
	for rows.Next() {
		var student models.Student
		if err := rows.Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.GroupID,
			&student.CreatedAt,
		); err != nil {
			return nil, apperrors.Internal("internal server error")
		}
		students = append(students, student)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Internal("internal server error")
	}
	return students, nil
}

func (r *Repository) StudentExists(ctx context.Context, studentID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM students WHERE id = $1)`,
		studentID,
	).Scan(&exists)
	if err != nil {
		return false, apperrors.Internal("internal server error")
	}
	return exists, nil
}

func (r *Repository) ListTasksByStudent(ctx context.Context, studentID int64) ([]models.Task, error) {
	exists, err := r.StudentExists(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.NotFound("student not found")
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, student_id, title, status, created_at
		 FROM tasks
		 WHERE student_id = $1
		 ORDER BY id`,
		studentID,
	)
	if err != nil {
		return nil, apperrors.Internal("internal server error")
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var task models.Task
		var status string
		if err := rows.Scan(&task.ID, &task.StudentID, &task.Title, &status, &task.CreatedAt); err != nil {
			return nil, apperrors.Internal("internal server error")
		}
		task.Status = models.TaskStatus(status)
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Internal("internal server error")
	}
	return tasks, nil
}

func (r *Repository) UpdateTaskStatus(
	ctx context.Context,
	taskID int64,
	status models.TaskStatus,
) (*models.Task, error) {
	var task models.Task
	var currentStatus string

	err := r.db.QueryRowContext(
		ctx,
		`UPDATE tasks
		 SET status = $1
		 WHERE id = $2
		 RETURNING id, student_id, title, status, created_at`,
		status,
		taskID,
	).Scan(&task.ID, &task.StudentID, &task.Title, &currentStatus, &task.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NotFound("task not found")
	}
	if err != nil {
		return nil, apperrors.Internal("internal server error")
	}

	task.Status = models.TaskStatus(currentStatus)
	return &task, nil
}
