package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"freshman-adaptation/internal/apperrors"
	"freshman-adaptation/internal/middleware"
	"freshman-adaptation/internal/models"
	"freshman-adaptation/internal/repository"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	repo      *repository.Repository
	validator *validator.Validate
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{
		repo:      repo,
		validator: validator.New(),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /groups", h.CreateGroup)
	mux.HandleFunc("GET /groups", h.ListGroups)
	mux.HandleFunc("POST /students", h.CreateStudent)
	mux.HandleFunc("GET /students", h.ListStudents)
	mux.HandleFunc("GET /students/{id}/tasks", h.ListStudentTasks)
	mux.HandleFunc("PATCH /tasks/{id}", h.UpdateTaskStatus)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupRequest
	if err := h.decodeAndValidate(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	group, err := h.repo.CreateGroup(r.Context(), req.Name)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, group)
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.repo.ListGroups(r.Context())
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if groups == nil {
		groups = []models.Group{}
	}
	middleware.WriteJSON(w, http.StatusOK, groups)
}

func (h *Handler) CreateStudent(w http.ResponseWriter, r *http.Request) {
	var req CreateStudentRequest
	if err := h.decodeAndValidate(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	student, err := h.repo.CreateStudentWithDefaultTasks(
		r.Context(),
		req.FirstName,
		req.LastName,
		req.GroupID,
	)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, student)
}

func (h *Handler) ListStudents(w http.ResponseWriter, r *http.Request) {
	groupIDRaw := r.URL.Query().Get("group_id")
	if groupIDRaw == "" {
		middleware.WriteError(w, apperrors.Validation("group_id is required"))
		return
	}

	groupID, err := strconv.ParseInt(groupIDRaw, 10, 64)
	if err != nil || groupID <= 0 {
		middleware.WriteError(w, apperrors.Validation("group_id must be a positive integer"))
		return
	}

	students, err := h.repo.ListStudentsByGroup(r.Context(), groupID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if students == nil {
		students = []models.Student{}
	}
	middleware.WriteJSON(w, http.StatusOK, students)
}

func (h *Handler) ListStudentTasks(w http.ResponseWriter, r *http.Request) {
	studentID, err := parsePositiveID(r.PathValue("id"), "student id")
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	tasks, err := h.repo.ListTasksByStudent(r.Context(), studentID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}
	if tasks == nil {
		tasks = []models.Task{}
	}
	middleware.WriteJSON(w, http.StatusOK, tasks)
}

func (h *Handler) UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID, err := parsePositiveID(r.PathValue("id"), "task id")
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	var req UpdateTaskStatusRequest
	if err := h.decodeAndValidate(r, &req); err != nil {
		middleware.WriteError(w, err)
		return
	}

	task, err := h.repo.UpdateTaskStatus(
		r.Context(),
		taskID,
		models.TaskStatus(req.Status),
	)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, task)
}

func (h *Handler) decodeAndValidate(r *http.Request, dst any) error {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return apperrors.Validation("invalid request body")
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return apperrors.Validation("request body is required")
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return apperrors.Validation("invalid JSON")
	}

	if err := h.validator.Struct(dst); err != nil {
		var validationErr validator.ValidationErrors
		if errors.As(err, &validationErr) {
			return apperrors.Validation(formatValidationError(validationErr))
		}
		return apperrors.Validation(err.Error())
	}

	return nil
}

func parsePositiveID(raw, fieldName string) (int64, error) {
	if raw == "" {
		return 0, apperrors.Validationf("%s is required", fieldName)
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperrors.Validationf("%s must be a positive integer", fieldName)
	}

	return id, nil
}

func formatValidationError(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "validation failed"
	}

	first := errs[0]
	field := strings.ToLower(first.Field())

	switch first.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "oneof":
		return fmt.Sprintf("%s has invalid value", field)
	case "min":
		return fmt.Sprintf("%s must not be empty", field)
	case "gt":
		return fmt.Sprintf("%s must be greater than 0", field)
	default:
		return fmt.Sprintf("%s failed validation on '%s'", field, first.Tag())
	}
}
