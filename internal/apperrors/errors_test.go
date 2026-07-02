package apperrors_test

import (
	"errors"
	"freshman-adaptation/internal/apperrors"
	"net/http"
	"testing"
)

func TestNotFound(t *testing.T) {
	err := apperrors.NotFound("task not found")

	if err.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", err.Status, http.StatusNotFound)
	}
	if err.Error() != "task not found" {
		t.Fatalf("message = %q", err.Error())
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatal("expected ErrNotFound sentinel")
	}
}

func TestValidation(t *testing.T) {
	err := apperrors.Validation("group_id is required")

	if err.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", err.Status, http.StatusBadRequest)
	}
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatal("expected ErrValidation sentinel")
	}
}

func TestInternal(t *testing.T) {
	err := apperrors.Internal("")

	if err.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", err.Status, http.StatusInternalServerError)
	}
	if err.Error() != "internal server error" {
		t.Fatalf("message = %q", err.Error())
	}
}

func TestFromSentinelAppError(t *testing.T) {
	original := apperrors.NotFound("student not found")
	mapped := apperrors.FromSentinel(original)

	if mapped != original {
		t.Fatal("AppError should pass through unchanged")
	}
}

func TestFromSentinelUnknown(t *testing.T) {
	mapped := apperrors.FromSentinel(errors.New("boom"))

	if mapped.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", mapped.Status, http.StatusInternalServerError)
	}
	if mapped.Error() != "internal server error" {
		t.Fatalf("message = %q", mapped.Error())
	}
}
