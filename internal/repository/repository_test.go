package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"freshman-adaptation/internal/apperrors"
	"freshman-adaptation/internal/models"
	"freshman-adaptation/internal/repository"
	"freshman-adaptation/internal/testdb"
	"net/http"
	"testing"
)

func TestCreateGroupAndListGroups(t *testing.T) {
	db := testdb.Open(t)
	repo := repository.New(db)

	group, err := repo.CreateGroup(context.Background(), "Группа 101")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if group.ID != 1 || group.Name != "Группа 101" {
		t.Fatalf("group = %#v", group)
	}

	groups, err := repo.ListGroups(context.Background())
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("groups count = %d, want 1", len(groups))
	}
}

func TestCreateStudentWithDefaultTasks(t *testing.T) {
	db := testdb.Open(t)
	repo := repository.New(db)

	group, err := repo.CreateGroup(context.Background(), "Группа 101")
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}

	student, err := repo.CreateStudentWithDefaultTasks(
		context.Background(),
		"Иван",
		"Петров",
		group.ID,
	)
	if err != nil {
		t.Fatalf("CreateStudentWithDefaultTasks: %v", err)
	}
	if student.GroupID != group.ID {
		t.Fatalf("student = %#v", student)
	}

	tasks, err := repo.ListTasksByStudent(context.Background(), student.ID)
	if err != nil {
		t.Fatalf("ListTasksByStudent: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("tasks count = %d, want 2", len(tasks))
	}
	for _, task := range tasks {
		if task.Status != models.TaskStatusTodo {
			t.Fatalf("task status = %q, want todo", task.Status)
		}
	}
}

func TestCreateStudentGroupNotFound(t *testing.T) {
	db := testdb.Open(t)
	repo := repository.New(db)

	_, err := repo.CreateStudentWithDefaultTasks(context.Background(), "Иван", "Петров", 999)
	assertAppError(t, err, http.StatusNotFound, "group not found")
}

func TestListTasksStudentNotFound(t *testing.T) {
	db := testdb.Open(t)
	repo := repository.New(db)

	_, err := repo.ListTasksByStudent(context.Background(), 999)
	assertAppError(t, err, http.StatusNotFound, "student not found")
}

func TestUpdateTaskStatus(t *testing.T) {
	db := testdb.Open(t)
	repo := repository.New(db)

	group, _ := repo.CreateGroup(context.Background(), "Группа 101")
	student, _ := repo.CreateStudentWithDefaultTasks(context.Background(), "Иван", "Петров", group.ID)
	tasks, _ := repo.ListTasksByStudent(context.Background(), student.ID)

	updated, err := repo.UpdateTaskStatus(
		context.Background(),
		tasks[0].ID,
		models.TaskStatusInProgress,
	)
	if err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}
	if updated.Status != models.TaskStatusInProgress {
		t.Fatalf("status = %q, want in_progress", updated.Status)
	}
}

func TestUpdateTaskStatusNotFound(t *testing.T) {
	db := testdb.Open(t)
	repo := repository.New(db)

	_, err := repo.UpdateTaskStatus(context.Background(), 999, models.TaskStatusDone)
	assertAppError(t, err, http.StatusNotFound, "task not found")
}

func TestCreateStudentTransactionRollback(t *testing.T) {
	db := testdb.Open(t)
	ctx := context.Background()

	groupID := createGroup(t, db, "Группа rollback")
	before := countStudents(t, db)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	var studentID int64
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO students (first_name, last_name, group_id)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		"Тест",
		"Откат",
		groupID,
	).Scan(&studentID)
	if err != nil {
		t.Fatalf("insert student: %v", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO tasks (student_id, title, status) VALUES ($1, $2, $3)`,
		studentID,
		"Задача 1",
		models.TaskStatusTodo,
	)
	if err != nil {
		t.Fatalf("insert first task: %v", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO tasks (student_id, title, status) VALUES ($1, $2, $3)`,
		studentID,
		"Задача 2",
		"invalid_status",
	)
	if err == nil {
		t.Fatal("expected CHECK constraint error")
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	after := countStudents(t, db)
	if after != before {
		t.Fatalf("students count changed: before=%d after=%d", before, after)
	}
}

func createGroup(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()

	var id int64
	err := db.QueryRow(
		`INSERT INTO groups (name) VALUES ($1) RETURNING id`,
		name,
	).Scan(&id)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	return id
}

func countStudents(t *testing.T, db *sql.DB) int {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&count); err != nil {
		t.Fatalf("count students: %v", err)
	}
	return count
}

func assertAppError(t *testing.T, err error, status int, message string) {
	t.Helper()

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Status != status {
		t.Fatalf("status = %d, want %d", appErr.Status, status)
	}
	if appErr.Error() != message {
		t.Fatalf("message = %q, want %q", appErr.Error(), message)
	}
}
