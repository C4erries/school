package finance_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/C4erries/school/backend/internal/application/crm"
	"github.com/C4erries/school/backend/internal/application/finance"
	"github.com/C4erries/school/backend/internal/application/schedule"
	"github.com/C4erries/school/backend/internal/domain"
)

type MockPaymentRepo struct {
	mock.Mock
}

func (m *MockPaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	return m.Called(ctx, p).Error(0)
}

func (m *MockPaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	args := m.Called(ctx, id)
	if p := args.Get(0); p != nil {
		return p.(*domain.Payment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepo) List(ctx context.Context, filter finance.PaymentFilter) ([]*domain.Payment, error) {
	args := m.Called(ctx, filter)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Payment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepo) SumAmountByPeriod(ctx context.Context, teacherID uuid.UUID, from, to time.Time) (float64, error) {
	args := m.Called(ctx, teacherID, from, to)
	return args.Get(0).(float64), args.Error(1)
}

type MockPayoutRepo struct {
	mock.Mock
}

func (m *MockPayoutRepo) Create(ctx context.Context, p *domain.PartnerPayout) error {
	return m.Called(ctx, p).Error(0)
}

func (m *MockPayoutRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.PartnerPayout, error) {
	args := m.Called(ctx, id)
	if p := args.Get(0); p != nil {
		return p.(*domain.PartnerPayout), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPayoutRepo) GetByTagAndPeriod(ctx context.Context, teacherID, tagID uuid.UUID, periodMonth string) (*domain.PartnerPayout, error) {
	args := m.Called(ctx, teacherID, tagID, periodMonth)
	if p := args.Get(0); p != nil {
		return p.(*domain.PartnerPayout), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPayoutRepo) ListByPeriod(ctx context.Context, teacherID uuid.UUID, periodMonth string) ([]*domain.PartnerPayout, error) {
	args := m.Called(ctx, teacherID, periodMonth)
	if list := args.Get(0); list != nil {
		return list.([]*domain.PartnerPayout), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPayoutRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.PartnerPayout, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.PartnerPayout), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockClientRepo struct {
	mock.Mock
}

func (m *MockClientRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Client, error) {
	args := m.Called(ctx, id)
	if c := args.Get(0); c != nil {
		return c.(*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClientRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID, filter ...crm.ClientFilter) ([]*domain.Client, error) {
	var f crm.ClientFilter
	if len(filter) > 0 {
		f = filter[0]
	}
	args := m.Called(ctx, teacherID, f)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Client), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockSubRepo struct {
	mock.Mock
}

func (m *MockSubRepo) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.ClientSubscription, error) {
	args := m.Called(ctx, clientID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.ClientSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSubRepo) Create(ctx context.Context, sub *domain.ClientSubscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *MockSubRepo) Update(ctx context.Context, sub *domain.ClientSubscription) error {
	return m.Called(ctx, sub).Error(0)
}

type MockLessonRepo struct {
	mock.Mock
}

func (m *MockLessonRepo) List(ctx context.Context, filter schedule.LessonFilter) ([]*domain.Lesson, error) {
	args := m.Called(ctx, filter)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Lesson), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockTagRepo struct {
	mock.Mock
}

func (m *MockTagRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	args := m.Called(ctx, id)
	if t := args.Get(0); t != nil {
		return t.(*domain.Tag), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTagRepo) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Tag, error) {
	args := m.Called(ctx, teacherID)
	if list := args.Get(0); list != nil {
		return list.([]*domain.Tag), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestFinanceService_CreatePayment(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	clientID := uuid.New()

	client := &domain.Client{
		ID:        clientID,
		TeacherID: teacherID,
		Name:      "Алексей Смирнов",
	}

	t.Run("success_existing_subscription_updated", func(t *testing.T) {
		paymentRepo := new(MockPaymentRepo)
		payoutRepo := new(MockPayoutRepo)
		clientRepo := new(MockClientRepo)
		subRepo := new(MockSubRepo)
		lessonRepo := new(MockLessonRepo)
		tagRepo := new(MockTagRepo)

		svc := finance.NewService(paymentRepo, payoutRepo, clientRepo, subRepo, lessonRepo, tagRepo)

		existingSub := &domain.ClientSubscription{
			ID:       uuid.New(),
			ClientID: clientID,
			Format:   domain.SubscriptionFormatIndividual,
			Balance:  2.0,
		}

		clientRepo.On("GetByID", ctx, clientID).Return(client, nil)
		paymentRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.Payment) bool {
			return p.Amount == 5000 && p.Hours == 4 && p.Format == domain.SubscriptionFormatIndividual
		})).Return(nil)
		subRepo.On("GetByClientID", ctx, clientID).Return([]*domain.ClientSubscription{existingSub}, nil)
		subRepo.On("Update", ctx, mock.MatchedBy(func(sub *domain.ClientSubscription) bool {
			return sub.ID == existingSub.ID && sub.Balance == 6.0
		})).Return(nil)

		paidAt := time.Now().UTC()
		item, err := svc.CreatePayment(ctx, finance.CreatePaymentInput{
			TeacherID:     teacherID,
			ClientID:      clientID,
			Amount:        5000,
			Hours:         4,
			Format:        domain.SubscriptionFormatIndividual,
			PaymentMethod: domain.PaymentMethodTransfer,
			PaidAt:        &paidAt,
			Notes:         "Оплата 4 занятий",
			CallerRole:    domain.RoleTeacher,
		})

		require.NoError(t, err)
		assert.Equal(t, "Алексей Смирнов", item.ClientName)
		assert.Equal(t, 5000.0, item.Payment.Amount)
		assert.Equal(t, 4.0, item.Payment.Hours)
		assert.Equal(t, domain.PaymentMethodTransfer, item.Payment.PaymentMethod)
	})

	t.Run("success_new_subscription_created", func(t *testing.T) {
		paymentRepo := new(MockPaymentRepo)
		payoutRepo := new(MockPayoutRepo)
		clientRepo := new(MockClientRepo)
		subRepo := new(MockSubRepo)
		lessonRepo := new(MockLessonRepo)
		tagRepo := new(MockTagRepo)

		svc := finance.NewService(paymentRepo, payoutRepo, clientRepo, subRepo, lessonRepo, tagRepo)

		clientRepo.On("GetByID", ctx, clientID).Return(client, nil)
		paymentRepo.On("Create", ctx, mock.Anything).Return(nil)
		subRepo.On("GetByClientID", ctx, clientID).Return([]*domain.ClientSubscription{}, nil)
		subRepo.On("Create", ctx, mock.MatchedBy(func(sub *domain.ClientSubscription) bool {
			return sub.ClientID == clientID && sub.Format == domain.SubscriptionFormatPair && sub.Balance == 2.0
		})).Return(nil)

		item, err := svc.CreatePayment(ctx, finance.CreatePaymentInput{
			TeacherID:     teacherID,
			ClientID:      clientID,
			Amount:        2000,
			Hours:         2,
			Format:        domain.SubscriptionFormatPair,
			PaymentMethod: domain.PaymentMethodCash,
			CallerRole:    domain.RoleTeacher,
		})

		require.NoError(t, err)
		assert.Equal(t, 2000.0, item.Payment.Amount)
	})

	t.Run("fails_unauthorized", func(t *testing.T) {
		paymentRepo := new(MockPaymentRepo)
		clientRepo := new(MockClientRepo)
		svc := finance.NewService(paymentRepo, nil, clientRepo, nil, nil, nil)

		otherTeacherID := uuid.New()
		clientRepo.On("GetByID", ctx, clientID).Return(client, nil)

		_, err := svc.CreatePayment(ctx, finance.CreatePaymentInput{
			TeacherID:     otherTeacherID,
			ClientID:      clientID,
			Amount:        2000,
			Hours:         2,
			Format:        domain.SubscriptionFormatIndividual,
			CallerRole:    domain.RoleTeacher,
		})

		require.ErrorIs(t, err, finance.ErrUnauthorizedAction)
	})

	t.Run("fails_validation", func(t *testing.T) {
		paymentRepo := new(MockPaymentRepo)
		clientRepo := new(MockClientRepo)
		svc := finance.NewService(paymentRepo, nil, clientRepo, nil, nil, nil)

		clientRepo.On("GetByID", ctx, clientID).Return(client, nil).Times(3)

		// Amount <= 0
		_, err := svc.CreatePayment(ctx, finance.CreatePaymentInput{
			TeacherID:     teacherID,
			ClientID:      clientID,
			Amount:        0,
			Hours:         2,
			Format:        domain.SubscriptionFormatIndividual,
			CallerRole:    domain.RoleTeacher,
		})
		require.ErrorIs(t, err, domain.ErrInvalidPaymentAmount)

		// Hours <= 0
		_, err = svc.CreatePayment(ctx, finance.CreatePaymentInput{
			TeacherID:     teacherID,
			ClientID:      clientID,
			Amount:        2000,
			Hours:         0,
			Format:        domain.SubscriptionFormatIndividual,
			CallerRole:    domain.RoleTeacher,
		})
		require.ErrorIs(t, err, domain.ErrInvalidPaymentHours)

		// Invalid format
		_, err = svc.CreatePayment(ctx, finance.CreatePaymentInput{
			TeacherID:     teacherID,
			ClientID:      clientID,
			Amount:        2000,
			Hours:         2,
			Format:        domain.SubscriptionFormat("invalid"),
			CallerRole:    domain.RoleTeacher,
		})
		require.ErrorIs(t, err, domain.ErrInvalidPaymentFormat)
	})
}

func TestFinanceService_ListPayments(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	clientID := uuid.New()

	paymentRepo := new(MockPaymentRepo)
	clientRepo := new(MockClientRepo)
	svc := finance.NewService(paymentRepo, nil, clientRepo, nil, nil, nil)

	payments := []*domain.Payment{
		{
			ID:            uuid.New(),
			TeacherID:     teacherID,
			ClientID:      clientID,
			Amount:        4500,
			Hours:         3,
			Format:        domain.SubscriptionFormatIndividual,
			PaymentMethod: domain.PaymentMethodTransfer,
			PaidAt:        time.Now().UTC(),
		},
	}

	filter := finance.PaymentFilter{TeacherID: &teacherID}
	paymentRepo.On("List", ctx, filter).Return(payments, nil)
	clientRepo.On("ListByTeacherID", ctx, teacherID, crm.ClientFilter{}).Return([]*domain.Client{
		{ID: clientID, Name: "Мария Петрова"},
	}, nil)

	items, err := svc.ListPayments(ctx, filter)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Мария Петрова", items[0].ClientName)
	assert.Equal(t, 4500.0, items[0].Payment.Amount)
}

func TestFinanceService_GetFinanceSummary(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	client1ID := uuid.New()
	client2ID := uuid.New()

	paymentRepo := new(MockPaymentRepo)
	payoutRepo := new(MockPayoutRepo)
	clientRepo := new(MockClientRepo)
	subRepo := new(MockSubRepo)
	lessonRepo := new(MockLessonRepo)
	tagRepo := new(MockTagRepo)

	svc := finance.NewService(paymentRepo, payoutRepo, clientRepo, subRepo, lessonRepo, tagRepo)

	periodMonth := "2026-10"
	startOfMonth := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// 1. Поступило оплат
	paymentRepo.On("SumAmountByPeriod", ctx, teacherID, startOfMonth, endOfMonth).Return(30000.0, nil)

	// 2. Клиенты: клиент 1 в долгу (-2 ч по индивид. ставке 1500 = 3000 долг), клиент 2 активный (4 ч)
	tagPartner := domain.Tag{
		ID:            uuid.New(),
		Name:          "Партнерская школа",
		SchoolPercent: 20,
	}

	client1 := &domain.Client{
		ID:             client1ID,
		TeacherID:      teacherID,
		Name:           "Клиент 1 (должник)",
		RateIndividual: 1500,
		BaseRate:       1500,
		Balances: domain.ClientBalances{
			IndividualHours: -2.0,
			TotalHours:      -2.0,
		},
		Tags: []domain.Tag{tagPartner},
	}
	client2 := &domain.Client{
		ID:             client2ID,
		TeacherID:      teacherID,
		Name:           "Клиент 2 (активный)",
		RateIndividual: 2000,
		BaseRate:       2000,
		Balances: domain.ClientBalances{
			IndividualHours: 4.0,
			TotalHours:      4.0,
		},
		Tags: []domain.Tag{},
	}

	clientRepo.On("ListByTeacherID", ctx, teacherID, crm.ClientFilter{}).Return([]*domain.Client{client1, client2}, nil)

	// 3. Проведенные уроки в этом месяце
	statusCompleted := domain.StatusCompleted
	lesson1 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client1ID,
		Format:    domain.FormatIndividual,
		StartTime: startOfMonth.Add(10 * time.Hour),
		EndTime:   startOfMonth.Add(12 * time.Hour), // 2 часа * 1500 = 3000 gross. Комиссия 20% = 600
		Status:    domain.StatusCompleted,
	}
	lesson2 := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  client2ID,
		Format:    domain.FormatIndividual,
		StartTime: startOfMonth.Add(15 * time.Hour),
		EndTime:   startOfMonth.Add(16 * time.Hour), // 1 час * 2000 = 2000 gross. Комиссия 0%
		Status:    domain.StatusCompleted,
	}

	lessonRepo.On("List", ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		Status:    &statusCompleted,
		From:      &startOfMonth,
		To:        &endOfMonth,
	}).Return([]*domain.Lesson{lesson1, lesson2}, nil)

	summary, err := svc.GetFinanceSummary(ctx, teacherID, periodMonth)
	require.NoError(t, err)

	assert.Equal(t, "2026-10", summary.Month)
	assert.Equal(t, 30000.0, summary.TotalPayments)
	assert.Equal(t, 5000.0, summary.TotalEarned)      // 3000 + 2000
	assert.Equal(t, 3000.0, summary.TotalDebts)       // 2h * 1500
	assert.Equal(t, 600.0, summary.TotalCommissions)  // 20% of 3000
	assert.Equal(t, 1, summary.ActiveSubscriptionsCount)
	assert.Equal(t, 1, summary.DebtorsCount)
}

func TestFinanceService_ListPartnerSettlements(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	tagID := uuid.New()
	clientID := uuid.New()

	paymentRepo := new(MockPaymentRepo)
	payoutRepo := new(MockPayoutRepo)
	clientRepo := new(MockClientRepo)
	subRepo := new(MockSubRepo)
	lessonRepo := new(MockLessonRepo)
	tagRepo := new(MockTagRepo)

	svc := finance.NewService(paymentRepo, payoutRepo, clientRepo, subRepo, lessonRepo, tagRepo)

	periodMonth := "2026-10"
	startOfMonth := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	tag := &domain.Tag{
		ID:            tagID,
		TeacherID:     teacherID,
		Name:          "Фоксфорд",
		SchoolPercent: 25,
		Color:         "orange",
	}

	client := &domain.Client{
		ID:             clientID,
		TeacherID:      teacherID,
		Name:           "Ученик школы",
		RateIndividual: 2000,
		Tags:           []domain.Tag{*tag},
	}

	statusCompleted := domain.StatusCompleted
	lesson := &domain.Lesson{
		ID:        uuid.New(),
		TeacherID: teacherID,
		ClientID:  clientID,
		Format:    domain.FormatIndividual,
		StartTime: startOfMonth.Add(2 * time.Hour),
		EndTime:   startOfMonth.Add(4 * time.Hour), // 2 hours * 2000 = 4000 gross, commission = 1000
		Status:    domain.StatusCompleted,
	}

	payoutID := uuid.New()
	paidAt := time.Now().UTC()
	existingPayout := &domain.PartnerPayout{
		ID:               payoutID,
		TeacherID:        teacherID,
		TagID:            tagID,
		PeriodMonth:      periodMonth,
		GrossAmount:      4000,
		CommissionAmount: 1000,
		PaidAt:           paidAt,
	}

	tagRepo.On("ListByTeacherID", ctx, teacherID).Return([]*domain.Tag{tag}, nil)
	payoutRepo.On("ListByPeriod", ctx, teacherID, periodMonth).Return([]*domain.PartnerPayout{existingPayout}, nil)
	lessonRepo.On("List", ctx, schedule.LessonFilter{
		TeacherID: &teacherID,
		Status:    &statusCompleted,
		From:      &startOfMonth,
		To:        &endOfMonth,
	}).Return([]*domain.Lesson{lesson}, nil)
	clientRepo.On("ListByTeacherID", ctx, teacherID, crm.ClientFilter{}).Return([]*domain.Client{client}, nil)

	items, err := svc.ListPartnerSettlements(ctx, teacherID, periodMonth)
	require.NoError(t, err)
	require.Len(t, items, 1)

	item := items[0]
	assert.Equal(t, tagID, item.TagID)
	assert.Equal(t, "Фоксфорд", item.TagName)
	assert.Equal(t, 25, item.SchoolPercent)
	assert.Equal(t, 1, item.LessonsCount)
	assert.Equal(t, 4000.0, item.GrossAmount)
	assert.Equal(t, 1000.0, item.CommissionAmount)
	assert.True(t, item.IsPaid)
	assert.Equal(t, payoutID, *item.PayoutID)
}

func TestFinanceService_CreatePartnerPayout(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	tagID := uuid.New()

	tag := &domain.Tag{
		ID:        tagID,
		TeacherID: teacherID,
		Name:      "Онлайн-школа",
	}

	t.Run("success", func(t *testing.T) {
		paymentRepo := new(MockPaymentRepo)
		payoutRepo := new(MockPayoutRepo)
		tagRepo := new(MockTagRepo)
		svc := finance.NewService(paymentRepo, payoutRepo, nil, nil, nil, tagRepo)

		tagRepo.On("GetByID", ctx, tagID).Return(tag, nil)
		payoutRepo.On("GetByTagAndPeriod", ctx, teacherID, tagID, "2026-10").Return(nil, domain.ErrPartnerPayoutNotFound)
		payoutRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.PartnerPayout) bool {
			return p.TagID == tagID && p.PeriodMonth == "2026-10" && p.GrossAmount == 10000 && p.CommissionAmount == 2000
		})).Return(nil)

		payout, err := svc.CreatePartnerPayout(ctx, finance.CreatePartnerPayoutInput{
			TeacherID:        teacherID,
			TagID:            tagID,
			PeriodMonth:      "2026-10",
			GrossAmount:      10000,
			CommissionAmount: 2000,
			Notes:            "Выплачено за октябрь",
			CallerRole:       domain.RoleTeacher,
		})

		require.NoError(t, err)
		assert.Equal(t, "2026-10", payout.PeriodMonth)
		assert.Equal(t, 10000.0, payout.GrossAmount)
		assert.Equal(t, 2000.0, payout.CommissionAmount)
	})

	t.Run("fails_already_exists", func(t *testing.T) {
		paymentRepo := new(MockPaymentRepo)
		payoutRepo := new(MockPayoutRepo)
		tagRepo := new(MockTagRepo)
		svc := finance.NewService(paymentRepo, payoutRepo, nil, nil, nil, tagRepo)

		existing := &domain.PartnerPayout{
			ID:          uuid.New(),
			PeriodMonth: "2026-10",
		}
		tagRepo.On("GetByID", ctx, tagID).Return(tag, nil)
		payoutRepo.On("GetByTagAndPeriod", ctx, teacherID, tagID, "2026-10").Return(existing, nil)

		_, err := svc.CreatePartnerPayout(ctx, finance.CreatePartnerPayoutInput{
			TeacherID:        teacherID,
			TagID:            tagID,
			PeriodMonth:      "2026-10",
			GrossAmount:      10000,
			CommissionAmount: 2000,
			CallerRole:       domain.RoleTeacher,
		})

		require.ErrorIs(t, err, finance.ErrPayoutAlreadyExists)
	})

	t.Run("fails_invalid_month", func(t *testing.T) {
		paymentRepo := new(MockPaymentRepo)
		payoutRepo := new(MockPayoutRepo)
		tagRepo := new(MockTagRepo)
		svc := finance.NewService(paymentRepo, payoutRepo, nil, nil, nil, tagRepo)

		tagRepo.On("GetByID", ctx, tagID).Return(tag, nil)

		_, err := svc.CreatePartnerPayout(ctx, finance.CreatePartnerPayoutInput{
			TeacherID:        teacherID,
			TagID:            tagID,
			PeriodMonth:      "invalid-month",
			GrossAmount:      10000,
			CommissionAmount: 2000,
			CallerRole:       domain.RoleTeacher,
		})

		require.ErrorIs(t, err, domain.ErrInvalidPeriodMonth)
	})
}

func TestFinanceService_CSVExports_WithUTF8BOM(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	clientID := uuid.New()

	t.Run("ExportClientsCSV", func(t *testing.T) {
		clientRepo := new(MockClientRepo)
		svc := finance.NewService(nil, nil, clientRepo, nil, nil, nil)

		phone := "+79991234567"
		clientRepo.On("ListByTeacherID", ctx, teacherID, crm.ClientFilter{}).Return([]*domain.Client{
			{
				ID:             clientID,
				TeacherID:      teacherID,
				Name:           "Ольга Иванова",
				Phone:          &phone,
				BaseRate:       1800,
				RateIndividual: 1800,
				Balances: domain.ClientBalances{
					IndividualHours: 5.5,
					TotalHours:      5.5,
				},
				Tags:      []domain.Tag{{Name: "ЕГЭ"}},
				CreatedAt: time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
			},
		}, nil)

		data, err := svc.ExportClientsCSV(ctx, teacherID)
		require.NoError(t, err)

		// Проверка UTF-8 BOM в начале
		require.True(t, bytes.HasPrefix(data, []byte("\xEF\xBB\xBF")), "must contain UTF-8 BOM prefix")

		// Парсим CSV (без BOM)
		reader := csv.NewReader(bytes.NewReader(data[3:]))
		records, err := reader.ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2) // Заголовок + 1 строка
		assert.Equal(t, "ID", records[0][0])
		assert.Equal(t, "Имя", records[0][1])
		assert.Equal(t, "Ольга Иванова", records[1][1])
		assert.Equal(t, "+79991234567", records[1][2])
		assert.Equal(t, "1800.00", records[1][4])
		assert.Equal(t, "5.50", records[1][7])
		assert.Equal(t, "ЕГЭ", records[1][11])
	})

	t.Run("ExportLessonsCSV", func(t *testing.T) {
		lessonRepo := new(MockLessonRepo)
		clientRepo := new(MockClientRepo)
		svc := finance.NewService(nil, nil, clientRepo, nil, lessonRepo, nil)

		start := time.Date(2026, 10, 5, 14, 0, 0, 0, time.UTC)
		end := start.Add(1 * time.Hour)

		lessonRepo.On("List", ctx, schedule.LessonFilter{
			TeacherID: &teacherID,
		}).Return([]*domain.Lesson{
			{
				ID:            uuid.New(),
				TeacherID:     teacherID,
				ClientID:      clientID,
				Title:         "Подготовка к профилю",
				Format:        domain.FormatIndividual,
				Status:        domain.StatusCompleted,
				StartTime:     start,
				EndTime:       end,
				LocationOrURL: "https://telemost.yandex.ru/j/123",
				Notes:         "Разобрали стереометрию",
			},
		}, nil)

		clientRepo.On("ListByTeacherID", ctx, teacherID, crm.ClientFilter{}).Return([]*domain.Client{
			{ID: clientID, Name: "Иван Сидоров"},
		}, nil)

		data, err := svc.ExportLessonsCSV(ctx, teacherID, nil, nil)
		require.NoError(t, err)

		require.True(t, bytes.HasPrefix(data, []byte("\xEF\xBB\xBF")))
		reader := csv.NewReader(bytes.NewReader(data[3:]))
		records, err := reader.ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2)
		assert.Equal(t, "Подготовка к профилю", records[1][3])
		assert.Equal(t, "Индивидуальное", records[1][4])
		assert.Equal(t, "Иван Сидоров", records[1][5])
		assert.Equal(t, "Завершен", records[1][6])
	})

	t.Run("ExportPaymentsCSV", func(t *testing.T) {
		paymentRepo := new(MockPaymentRepo)
		clientRepo := new(MockClientRepo)
		svc := finance.NewService(paymentRepo, nil, clientRepo, nil, nil, nil)

		paidAt := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
		paymentRepo.On("List", ctx, finance.PaymentFilter{
			TeacherID: &teacherID,
		}).Return([]*domain.Payment{
			{
				ID:            uuid.New(),
				TeacherID:     teacherID,
				ClientID:      clientID,
				Amount:        6000,
				Hours:         4,
				Format:        domain.SubscriptionFormatIndividual,
				PaymentMethod: domain.PaymentMethodTransfer,
				PaidAt:        paidAt,
				Notes:         "Оплата за октябрь",
				CreatedAt:     paidAt,
			},
		}, nil)

		clientRepo.On("ListByTeacherID", ctx, teacherID, crm.ClientFilter{}).Return([]*domain.Client{
			{ID: clientID, Name: "Дмитрий Соколов"},
		}, nil)

		data, err := svc.ExportPaymentsCSV(ctx, teacherID, nil, nil)
		require.NoError(t, err)

		require.True(t, bytes.HasPrefix(data, []byte("\xEF\xBB\xBF")))
		reader := csv.NewReader(bytes.NewReader(data[3:]))
		records, err := reader.ReadAll()
		require.NoError(t, err)

		require.Len(t, records, 2)
		assert.Equal(t, "Дмитрий Соколов", records[1][2])
		assert.Equal(t, "6000.00", records[1][3])
		assert.Equal(t, "4.00", records[1][4])
		assert.Equal(t, "Индивидуальное", records[1][5])
		assert.Equal(t, "Перевод / СБП", records[1][6])
	})
}
