package schedule

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/response"
)

// ClassroomHandler обрабатывает запросы учебных кабинетов.
type ClassroomHandler struct {
	scheduleService *schedule.Service
	authHandler     *auth.Handler
}

func NewClassroomHandler(scheduleService *schedule.Service, authHandler *auth.Handler) *ClassroomHandler {
	return &ClassroomHandler{
		scheduleService: scheduleService,
		authHandler:     authHandler,
	}
}

func (h *ClassroomHandler) ListClassrooms(c echo.Context) error {
	_, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	classrooms, err := h.scheduleService.ListClassrooms(c.Request().Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list classrooms")
	}

	resp := make([]generated.ClassroomResponse, 0, len(classrooms))
	for _, cls := range classrooms {
		var desc *string
		if cls.Description != "" {
			desc = &cls.Description
		}
		resp = append(resp, generated.ClassroomResponse{
			Id:          cls.ID,
			Name:        cls.Name,
			Capacity:    cls.Capacity,
			Color:       cls.Color,
			Description: desc,
			CreatedAt:   cls.CreatedAt,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *ClassroomHandler) CreateClassroom(c echo.Context) error {
	claims, ok := h.authHandler.Authenticate(c)
	if !ok {
		return nil
	}

	if claims.Role != domain.RoleOwner {
		return response.Error(c, http.StatusForbidden, "FORBIDDEN", "only admin can create classrooms")
	}

	var req generated.CreateClassroomRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	var color, desc string
	if req.Color != nil {
		color = *req.Color
	}
	if req.Description != nil {
		desc = *req.Description
	}

	cls, err := h.scheduleService.CreateClassroom(c.Request().Context(), schedule.CreateClassroomInput{
		Name:        req.Name,
		Capacity:    req.Capacity,
		Color:       color,
		Description: desc,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidClassroomName) || errors.Is(err, domain.ErrInvalidClassroomCapacity) {
			return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create classroom")
	}

	var descPtr *string
	if cls.Description != "" {
		descPtr = &cls.Description
	}

	return c.JSON(http.StatusCreated, generated.ClassroomResponse{
		Id:          cls.ID,
		Name:        cls.Name,
		Capacity:    cls.Capacity,
		Color:       cls.Color,
		Description: descPtr,
		CreatedAt:   cls.CreatedAt,
	})
}

