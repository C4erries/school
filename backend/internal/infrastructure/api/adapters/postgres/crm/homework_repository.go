package crm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/postgres"
)

// HomeworkRepository реализует сохранение и выборку домашних заданий в PostgreSQL.
type HomeworkRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

// NewHomeworkRepository создает новый экземпляр HomeworkRepository.
func NewHomeworkRepository(db *sql.DB) *HomeworkRepository {
	return &HomeworkRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *HomeworkRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create сохраняет новое домашнее задание.
func (r *HomeworkRepository) Create(ctx context.Context, hw *domain.HomeworkAssignment) error {
	query, args, err := r.sb.Insert("homework_assignments").
		Columns(
			"id",
			"client_id",
			"teacher_id",
			"assigned_lesson_id",
			"title",
			"description",
			"due_date",
			"status",
			"review_notes",
			"created_at",
			"updated_at",
		).
		Values(
			hw.ID,
			hw.ClientID,
			hw.TeacherID,
			hw.AssignedLessonID,
			hw.Title,
			hw.Description,
			hw.DueDate,
			string(hw.Status),
			hw.ReviewNotes,
			hw.CreatedAt,
			hw.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert homework query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert homework: %w", err)
	}

	return nil
}

// GetByID находит домашнее задание по идентификатору.
func (r *HomeworkRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.HomeworkAssignment, error) {
	query, args, err := r.sb.Select(
		"id",
		"client_id",
		"teacher_id",
		"assigned_lesson_id",
		"title",
		"description",
		"due_date",
		"status",
		"review_notes",
		"created_at",
		"updated_at",
	).
		From("homework_assignments").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select homework query: %w", err)
	}

	hw, err := r.scanHomework(r.getDBTX(ctx).QueryRowContext(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrHomeworkNotFound
		}
		return nil, fmt.Errorf("scan homework: %w", err)
	}

	return hw, nil
}

// Update обновляет поля домашнего задания.
func (r *HomeworkRepository) Update(ctx context.Context, hw *domain.HomeworkAssignment) error {
	query, args, err := r.sb.Update("homework_assignments").
		Set("title", hw.Title).
		Set("description", hw.Description).
		Set("due_date", hw.DueDate).
		Set("status", string(hw.Status)).
		Set("review_notes", hw.ReviewNotes).
		Set("assigned_lesson_id", hw.AssignedLessonID).
		Set("updated_at", hw.UpdatedAt).
		Where(sq.Eq{"id": hw.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update homework query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec update homework: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected update homework: %w", err)
	}
	if rows == 0 {
		return domain.ErrHomeworkNotFound
	}

	return nil
}

// Delete удаляет задание с проверкой принадлежности репетитору.
func (r *HomeworkRepository) Delete(ctx context.Context, id, teacherID uuid.UUID) error {
	query, args, err := r.sb.Delete("homework_assignments").
		Where(sq.Eq{
			"id":         id,
			"teacher_id": teacherID,
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete homework query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec delete homework: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected delete homework: %w", err)
	}
	if rows == 0 {
		// Проверяем, существует ли запись
		var existingTeacher uuid.UUID
		checkQuery, checkArgs, _ := r.sb.Select("teacher_id").
			From("homework_assignments").
			Where(sq.Eq{"id": id}).
			ToSql()
		if scanErr := r.getDBTX(ctx).QueryRowContext(ctx, checkQuery, checkArgs...).Scan(&existingTeacher); scanErr != nil {
			return domain.ErrHomeworkNotFound
		}
		return domain.ErrForbidden
	}

	return nil
}

// ListByClientID возвращает все домашние задания ученика с опциональным фильтром по статусу.
func (r *HomeworkRepository) ListByClientID(
	ctx context.Context,
	clientID, teacherID uuid.UUID,
	statusFilter *domain.HomeworkStatus,
) ([]*domain.HomeworkAssignment, error) {
	builder := r.sb.Select(
		"id",
		"client_id",
		"teacher_id",
		"assigned_lesson_id",
		"title",
		"description",
		"due_date",
		"status",
		"review_notes",
		"created_at",
		"updated_at",
	).
		From("homework_assignments").
		Where(sq.Eq{
			"client_id":  clientID,
			"teacher_id": teacherID,
		})

	if statusFilter != nil && statusFilter.IsValid() {
		builder = builder.Where(sq.Eq{"status": string(*statusFilter)})
	}

	query, args, err := builder.OrderBy("created_at DESC").ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list client homework query: %w", err)
	}

	return r.queryHomeworks(ctx, query, args)
}

// ListByAssignedLessonID возвращает задания, выданные на конкретном уроке.
func (r *HomeworkRepository) ListByAssignedLessonID(ctx context.Context, lessonID uuid.UUID) ([]*domain.HomeworkAssignment, error) {
	query, args, err := r.sb.Select(
		"id",
		"client_id",
		"teacher_id",
		"assigned_lesson_id",
		"title",
		"description",
		"due_date",
		"status",
		"review_notes",
		"created_at",
		"updated_at",
	).
		From("homework_assignments").
		Where(sq.Eq{"assigned_lesson_id": lessonID}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list by assigned lesson query: %w", err)
	}

	return r.queryHomeworks(ctx, query, args)
}

// ListDueByLessonOrDate возвращает задания, срок сдачи которых приходится на targetDate (или просроченные активные).
func (r *HomeworkRepository) ListDueByLessonOrDate(
	ctx context.Context,
	clientID uuid.UUID,
	targetDate time.Time,
) ([]*domain.HomeworkAssignment, error) {
	dateStr := targetDate.Format("2006-01-02")
	query, args, err := r.sb.Select(
		"id",
		"client_id",
		"teacher_id",
		"assigned_lesson_id",
		"title",
		"description",
		"due_date",
		"status",
		"review_notes",
		"created_at",
		"updated_at",
	).
		From("homework_assignments").
		Where(sq.Eq{"client_id": clientID}).
		Where(sq.Or{
			sq.Expr("due_date = ?::date", dateStr),
			sq.And{
				sq.Eq{"status": string(domain.StatusHomeworkAssigned)},
				sq.Expr("due_date <= ?::date", dateStr),
			},
		}).
		OrderBy("due_date ASC", "created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list due homework query: %w", err)
	}

	return r.queryHomeworks(ctx, query, args)
}

func (r *HomeworkRepository) scanHomework(row sq.RowScanner) (*domain.HomeworkAssignment, error) {
	var hw domain.HomeworkAssignment
	var assignedLessonID uuid.NullUUID
	var desc, reviewNotes sql.NullString
	var statusStr string
	var dueDate sql.NullTime

	err := row.Scan(
		&hw.ID,
		&hw.ClientID,
		&hw.TeacherID,
		&assignedLessonID,
		&hw.Title,
		&desc,
		&dueDate,
		&statusStr,
		&reviewNotes,
		&hw.CreatedAt,
		&hw.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if assignedLessonID.Valid {
		id := assignedLessonID.UUID
		hw.AssignedLessonID = &id
	}
	if desc.Valid {
		hw.Description = desc.String
	}
	if reviewNotes.Valid {
		hw.ReviewNotes = reviewNotes.String
	}
	if dueDate.Valid {
		d := dueDate.Time
		hw.DueDate = &d
	}
	hw.Status = domain.HomeworkStatus(statusStr)

	return &hw, nil
}

func (r *HomeworkRepository) queryHomeworks(ctx context.Context, query string, args []any) ([]*domain.HomeworkAssignment, error) {
	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query homeworks: %w", err)
	}
	defer rows.Close()

	result := make([]*domain.HomeworkAssignment, 0)
	for rows.Next() {
		hw, err := r.scanHomework(rows)
		if err != nil {
			return nil, fmt.Errorf("scan homework row: %w", err)
		}
		result = append(result, hw)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate homework rows: %w", err)
	}

	return result, nil
}
