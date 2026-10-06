package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/domain"
	httpauth "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/auth"
	httpcrm "github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/crm"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/security"
)

type MockTagRepoForHandler struct {
	mock.Mock
}

func (m *MockTagRepoForHandler) Create(ctx context.Context, tag *domain.Tag) error {
	return m.Called(ctx, tag).Error(0)
}
func (m *MockTagRepoForHandler) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	args := m.Called(ctx, id)
	if t := args.Get(0); t != nil {
		return t.(*domain.Tag), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTagRepoForHandler) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Tag, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Tag), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTagRepoForHandler) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockTagRepoForHandler) AssignToClient(ctx context.Context, clientID, tagID uuid.UUID) error {
	return m.Called(ctx, clientID, tagID).Error(0)
}
func (m *MockTagRepoForHandler) RemoveFromClient(ctx context.Context, clientID, tagID uuid.UUID) error {
	return m.Called(ctx, clientID, tagID).Error(0)
}
func (m *MockTagRepoForHandler) SetClientTags(ctx context.Context, clientID uuid.UUID, tagIDs []uuid.UUID) error {
	return m.Called(ctx, clientID, tagIDs).Error(0)
}

func TestTagHandler(t *testing.T) {
	tagRepo := new(MockTagRepoForHandler)
	crmSvc := crm.NewService(nil, nil, tagRepo)
	tokenMgr := new(MockTokenManager)

	teacherID := uuid.New()
	teacherClaims := &security.UserClaims{
		UserID: teacherID,
		Role:   domain.RoleTeacher,
	}
	tokenMgr.On("ValidateAccessToken", "teacher_token").Return(teacherClaims, nil)

	authHandler := httpauth.NewHandler(nil, tokenMgr)
	tagHandler := httpcrm.NewTagHandler(crmSvc, authHandler)

	t.Run("CreateTag success", func(t *testing.T) {
		tagRepo.On("Create", mock.Anything, mock.MatchedBy(func(tag *domain.Tag) bool {
			return tag.TeacherID == teacherID && tag.Name == "Школа №12" && tag.SchoolPercent == 30
		})).Return(nil).Once()

		reqBody := `{"name":"Школа №12","school_percent":30,"color":"emerald"}`
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/tags", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer teacher_token")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := tagHandler.CreateTag(c)
		require.NoError(t, err)

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp generated.TagResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Школа №12", resp.Name)
		assert.Equal(t, 30, resp.SchoolPercent)
		assert.Equal(t, "emerald", resp.Color)
		tagRepo.AssertExpectations(t)
	})

	t.Run("ListTags success", func(t *testing.T) {
		tags := []*domain.Tag{
			{ID: uuid.New(), TeacherID: teacherID, Name: "Школа №12", SchoolPercent: 30, Color: "emerald", CreatedAt: time.Now()},
		}
		tagRepo.On("ListByTeacherID", mock.Anything, teacherID).Return(tags, nil).Once()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/tags", nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := tagHandler.ListTags(c)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp []generated.TagResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Школа №12", resp[0].Name)
		tagRepo.AssertExpectations(t)
	})

	t.Run("DeleteTag success", func(t *testing.T) {
		tagID := uuid.New()
		tagRepo.On("GetByID", mock.Anything, tagID).Return(&domain.Tag{
			ID:        tagID,
			TeacherID: teacherID,
		}, nil).Once()
		tagRepo.On("Delete", mock.Anything, tagID).Return(nil).Once()

		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/tags/"+tagID.String(), nil)
		req.Header.Set("Authorization", "Bearer teacher_token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := tagHandler.DeleteTag(c, tagID)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		tagRepo.AssertExpectations(t)
	})
}
