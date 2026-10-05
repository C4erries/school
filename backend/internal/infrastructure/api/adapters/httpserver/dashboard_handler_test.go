package httpserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/dashboard"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/httpserver/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

func TestDashboardHandler_Metrics(t *testing.T) {
	clientRepo := new(MockCRMClientRepo)
	lessonRepo := new(MockLessonRepo)
	dashboardSvc := dashboard.NewService(lessonRepo, clientRepo)
	tokenMgr := new(MockTokenManager)

	teacherID := uuid.New()
	teacherClaims := &security.UserClaims{
		UserID: teacherID,
		Role:   domain.RoleTeacher,
	}
	tokenMgr.On("ValidateAccessToken", "teacher_token").Return(teacherClaims, nil)

	handler := httpserver.NewAPIHandler(nil, nil, nil, dashboardSvc, tokenMgr, "v1")

	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC)

	clientID := uuid.New()
	client := &domain.Client{
		ID:               clientID,
		TeacherID:        teacherID,
		BaseRate:         1500,
		SchoolPercentTag: 10,
	}

	lessons := []*domain.Lesson{
		{
			ID:        uuid.New(),
			TeacherID: teacherID,
			ClientID:  clientID,
			StartTime: from.Add(time.Hour),
			EndTime:   from.Add(2 * time.Hour),
			Status:    domain.StatusCompleted,
		},
	}

	lessonRepo.On("List", mock.Anything, schedule.LessonFilter{
		TeacherID: &teacherID,
		From:      &from,
		To:        &to,
	}).Return(lessons, nil).Once()

	clientRepo.On("GetByID", mock.Anything, clientID).Return(client, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics?from="+from.Format(time.RFC3339)+"&to="+to.Format(time.RFC3339), nil)
	req.Header.Set("Authorization", "Bearer teacher_token")
	rec := httptest.NewRecorder()

	handler.GetDashboardMetrics(rec, req, generated.GetDashboardMetricsParams{
		From: from,
		To:   to,
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp generated.DashboardMetricsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float32(1500), resp.GrossPotentialRevenue)
	assert.Equal(t, float32(1350), resp.NetIncome) // 1500 - 10% (150) = 1350
	assert.Equal(t, float32(1500), resp.AverageRate)
}
