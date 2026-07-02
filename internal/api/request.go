package api

// CreateGroupRequest тело для POST /groups
type CreateGroupRequest struct {
	Name string `json:"name" validate:"required,min=1"`
}

// CreateStudentRequest тело для POST /students
type CreateStudentRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1"`
	LastName  string `json:"last_name" validate:"required,min=1"`
	GroupID   int64  `json:"group_id" validate:"required,gt=0"`
}

// UpdateTaskStatusRequest тело для PATCH /tasks/{id}
type UpdateTaskStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=todo in_progress done"`
}
