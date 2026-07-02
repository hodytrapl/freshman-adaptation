package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"freshman-adaptation/internal/api"
	"freshman-adaptation/internal/apperrors"
	"freshman-adaptation/internal/middleware"
	"freshman-adaptation/internal/repository"
	"freshman-adaptation/internal/testdb"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockServer(t *testing.T) (*httptest.Server, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	repo := repository.New(db)
	handler := api.NewHandler(repo)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	server := httptest.NewServer(middleware.Recover(mux))

	t.Cleanup(func() {
		server.Close()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("sql expectations: %v", err)
		}
		_ = db.Close()
	})

	return server, mock
}

func newIntegrationServer(t *testing.T) *httptest.Server {
	t.Helper()

	db := testdb.Open(t)
	repo := repository.New(db)
	handler := api.NewHandler(repo)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	server := httptest.NewServer(middleware.Recover(mux))
	t.Cleanup(server.Close)

	return server
}

func TestHealth(t *testing.T) {
	server, _ := newMockServer(t)

	resp := mustDo(t, server.Client(), http.MethodGet, server.URL+"/health", nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var payload map[string]string
	decodeJSON(t, resp.Body, &payload)
	if payload["status"] != "ok" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestAPIFlowIntegration(t *testing.T) {
	server := newIntegrationServer(t)
	client := server.Client()

	resp := mustDo(t, client, http.MethodPost, server.URL+"/groups", jsonBody(t, map[string]string{
		"name": "Группа 101",
	}))
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusCreated)

	var group struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	decodeJSON(t, resp.Body, &group)

	resp = mustDo(t, client, http.MethodPost, server.URL+"/students", jsonBody(t, map[string]any{
		"first_name": "Иван",
		"last_name":  "Петров",
		"group_id":   group.ID,
	}))
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusCreated)

	var student struct {
		ID int64 `json:"id"`
	}
	decodeJSON(t, resp.Body, &student)

	resp = mustDo(t, client, http.MethodGet, fmt.Sprintf("%s/students?group_id=%d", server.URL, group.ID), nil)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)

	var students []map[string]any
	decodeJSON(t, resp.Body, &students)
	if len(students) != 1 {
		t.Fatalf("students = %#v", students)
	}

	resp = mustDo(t, client, http.MethodGet, fmt.Sprintf("%s/students/%d/tasks", server.URL, student.ID), nil)
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)

	var tasks []struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	decodeJSON(t, resp.Body, &tasks)
	if len(tasks) != 2 {
		t.Fatalf("tasks = %#v", tasks)
	}

	resp = mustDo(t, client, http.MethodPatch, fmt.Sprintf("%s/tasks/%d", server.URL, tasks[0].ID), jsonBody(t, map[string]string{
		"status": "in_progress",
	}))
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)

	var updatedTask struct {
		Status string `json:"status"`
	}
	decodeJSON(t, resp.Body, &updatedTask)
	if updatedTask.Status != "in_progress" {
		t.Fatalf("updated task = %#v", updatedTask)
	}
}

func TestListStudentsMissingGroupID(t *testing.T) {
	server, _ := newMockServer(t)

	resp := mustDo(t, server.Client(), http.MethodGet, server.URL+"/students", nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorMessage(t, resp.Body, "group_id is required")
}

func TestPatchTaskInvalidStatus(t *testing.T) {
	server, _ := newMockServer(t)

	resp := mustDo(t, server.Client(), http.MethodPatch, server.URL+"/tasks/1", jsonBody(t, map[string]string{
		"status": "bad",
	}))
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest)
	assertErrorMessage(t, resp.Body, "status has invalid value")
}

func TestPatchTaskNotFound(t *testing.T) {
	server, mock := newMockServer(t)

	mock.ExpectQuery(`UPDATE tasks`).
		WithArgs("done", int64(99999)).
		WillReturnError(sql.ErrNoRows)

	resp := mustDo(t, server.Client(), http.MethodPatch, server.URL+"/tasks/99999", jsonBody(t, map[string]string{
		"status": "done",
	}))
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound)
	assertErrorMessage(t, resp.Body, "task not found")
}

func TestPatchTaskSuccess(t *testing.T) {
	server, mock := newMockServer(t)
	createdAt := time.Now().UTC()

	mock.ExpectQuery(`UPDATE tasks`).
		WithArgs("in_progress", int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "student_id", "title", "status", "created_at"}).
			AddRow(int64(1), int64(10), "Связаться с куратором", "in_progress", createdAt))

	resp := mustDo(t, server.Client(), http.MethodPatch, server.URL+"/tasks/1", jsonBody(t, map[string]string{
		"status": "in_progress",
	}))
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)

	var task struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
		Title  string `json:"title"`
	}
	decodeJSON(t, resp.Body, &task)
	if task.ID != 1 || task.Status != "in_progress" {
		t.Fatalf("task = %#v", task)
	}
}

func TestStudentTasksNotFound(t *testing.T) {
	server, mock := newMockServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(int64(99999)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	resp := mustDo(t, server.Client(), http.MethodGet, server.URL+"/students/99999/tasks", nil)
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound)
	assertErrorMessage(t, resp.Body, "student not found")
}

func TestCreateStudentGroupNotFound(t *testing.T) {
	server, mock := newMockServer(t)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(int64(99999)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	resp := mustDo(t, server.Client(), http.MethodPost, server.URL+"/students", jsonBody(t, map[string]any{
		"first_name": "Мария",
		"last_name":  "Сидорова",
		"group_id":   99999,
	}))
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusNotFound)
	assertErrorMessage(t, resp.Body, "group not found")
}

func TestCreateGroupValidation(t *testing.T) {
	server, _ := newMockServer(t)

	resp := mustDo(t, server.Client(), http.MethodPost, server.URL+"/groups", jsonBody(t, map[string]string{
		"name": "",
	}))
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest)
}

func TestCreateGroupSuccess(t *testing.T) {
	server, mock := newMockServer(t)

	mock.ExpectQuery(`INSERT INTO groups`).
		WithArgs("Группа 101").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "Группа 101"))

	resp := mustDo(t, server.Client(), http.MethodPost, server.URL+"/groups", jsonBody(t, map[string]string{
		"name": "Группа 101",
	}))
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusCreated)

	var group struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	decodeJSON(t, resp.Body, &group)
	if group.ID != 1 || group.Name != "Группа 101" {
		t.Fatalf("group = %#v", group)
	}
}

func TestAppErrorSentinelMapping(t *testing.T) {
	err := apperrors.NotFound("task not found")
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError")
	}
	if appErr.Status != http.StatusNotFound {
		t.Fatalf("status = %d", appErr.Status)
	}
}

func mustDo(t *testing.T, client *http.Client, method, url string, body io.Reader) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	return resp
}

func jsonBody(t *testing.T, payload any) *bytes.Buffer {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	return bytes.NewBuffer(data)
}

func decodeJSON(t *testing.T, body io.Reader, dst any) {
	t.Helper()

	if err := json.NewDecoder(body).Decode(dst); err != nil {
		t.Fatalf("Decode: %v", err)
	}
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()

	if resp.StatusCode != want {
		t.Fatalf("status = %d, want %d", resp.StatusCode, want)
	}
}

func assertErrorMessage(t *testing.T, body io.Reader, want string) {
	t.Helper()

	var payload struct {
		Error string `json:"error"`
	}
	decodeJSON(t, body, &payload)
	if payload.Error != want {
		t.Fatalf("error = %q, want %q", payload.Error, want)
	}
}
