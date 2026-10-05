package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	valkeylib "github.com/valkey-io/valkey-go"

	"github.com/C4erries/school/backend/internal/domain"
)

// TagRepository реализует хранение тегов репетитора и привязок к клиентам в PostgreSQL.
type TagRepository struct {
	db           *sql.DB
	sb           sq.StatementBuilderType
	valkeyClient valkeylib.Client
}

func NewTagRepository(db *sql.DB, valkeyClient ...valkeylib.Client) *TagRepository {
	repo := &TagRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
	if len(valkeyClient) > 0 {
		repo.valkeyClient = valkeyClient[0]
	}
	return repo
}

func (r *TagRepository) getDBTX(ctx context.Context) DBTX {
	if tx := ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create создает новый тег.
func (r *TagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	query, args, err := r.sb.Insert("tags").
		Columns("id", "teacher_id", "name", "school_percent", "color", "created_at").
		Values(tag.ID, tag.TeacherID, tag.Name, tag.SchoolPercent, tag.Color, tag.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert tag query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec insert tag: %w", err)
	}

	if r.valkeyClient != nil {
		_ = r.valkeyClient.Do(ctx, r.valkeyClient.B().Del().Key(fmt.Sprintf("cache:tags:teacher:%s", tag.TeacherID)).Build())
	}

	return nil
}

// GetByID находит тег по идентификатору.
func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	cacheKey := fmt.Sprintf("cache:tags:%s", id)
	if r.valkeyClient != nil {
		cmd := r.valkeyClient.B().Get().Key(cacheKey).Cache()
		res := r.valkeyClient.DoCache(ctx, cmd, 10*time.Minute)
		if str, err := res.ToString(); err == nil && str != "" {
			var cached domain.Tag
			if err := json.Unmarshal([]byte(str), &cached); err == nil {
				return &cached, nil
			}
		}
	}

	query, args, err := r.sb.Select("id", "teacher_id", "name", "school_percent", "color", "created_at").
		From("tags").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select tag by id query: %w", err)
	}

	var t domain.Tag
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(
		&t.ID,
		&t.TeacherID,
		&t.Name,
		&t.SchoolPercent,
		&t.Color,
		&t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTagNotFound
		}
		return nil, fmt.Errorf("query tag by id: %w", err)
	}

	if r.valkeyClient != nil {
		if data, err := json.Marshal(t); err == nil {
			_ = r.valkeyClient.Do(ctx, r.valkeyClient.B().Set().Key(cacheKey).Value(string(data)).Ex(10*time.Minute).Build())
		}
	}

	return &t, nil
}

// ListByTeacherID возвращает все теги преподавателя.
func (r *TagRepository) ListByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]*domain.Tag, error) {
	cacheKey := fmt.Sprintf("cache:tags:teacher:%s", teacherID)
	if r.valkeyClient != nil {
		cmd := r.valkeyClient.B().Get().Key(cacheKey).Cache()
		res := r.valkeyClient.DoCache(ctx, cmd, 10*time.Minute)
		if str, err := res.ToString(); err == nil && str != "" {
			var cached []*domain.Tag
			if err := json.Unmarshal([]byte(str), &cached); err == nil {
				return cached, nil
			}
		}
	}

	query, args, err := r.sb.Select("id", "teacher_id", "name", "school_percent", "color", "created_at").
		From("tags").
		Where(sq.Eq{"teacher_id": teacherID}).
		OrderBy("name ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list tags query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(
			&t.ID,
			&t.TeacherID,
			&t.Name,
			&t.SchoolPercent,
			&t.Color,
			&t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	if r.valkeyClient != nil {
		if data, err := json.Marshal(tags); err == nil {
			_ = r.valkeyClient.Do(ctx, r.valkeyClient.B().Set().Key(cacheKey).Value(string(data)).Ex(10*time.Minute).Build())
		}
	}

	return tags, nil
}

// Delete удаляет тег по идентификатору.
func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var teacherID uuid.UUID
	_ = r.getDBTX(ctx).QueryRowContext(ctx, "SELECT teacher_id FROM tags WHERE id = $1", id).Scan(&teacherID)

	query, args, err := r.sb.Delete("tags").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete tag query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec delete tag: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrTagNotFound
	}

	if r.valkeyClient != nil {
		delKeys := []string{fmt.Sprintf("cache:tags:%s", id)}
		if teacherID != uuid.Nil {
			delKeys = append(delKeys, fmt.Sprintf("cache:tags:teacher:%s", teacherID))
		}
		_ = r.valkeyClient.Do(ctx, r.valkeyClient.B().Del().Key(delKeys...).Build())
	}

	return nil
}

// AssignToClient привязывает тег к клиенту.
func (r *TagRepository) AssignToClient(ctx context.Context, clientID, tagID uuid.UUID) error {
	query := `INSERT INTO client_tags (client_id, tag_id) VALUES ($1, $2) ON CONFLICT (client_id, tag_id) DO NOTHING`
	_, err := r.getDBTX(ctx).ExecContext(ctx, query, clientID, tagID)
	if err != nil {
		return fmt.Errorf("assign tag to client: %w", err)
	}
	return nil
}

// RemoveFromClient отвязывает тег от клиента.
func (r *TagRepository) RemoveFromClient(ctx context.Context, clientID, tagID uuid.UUID) error {
	query := `DELETE FROM client_tags WHERE client_id = $1 AND tag_id = $2`
	_, err := r.getDBTX(ctx).ExecContext(ctx, query, clientID, tagID)
	if err != nil {
		return fmt.Errorf("remove tag from client: %w", err)
	}
	return nil
}

// GetByClientID возвращает все теги, привязанные к клиенту.
func (r *TagRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]domain.Tag, error) {
	query := `SELECT t.id, t.teacher_id, t.name, t.school_percent, t.color, t.created_at
		FROM tags t
		JOIN client_tags ct ON ct.tag_id = t.id
		WHERE ct.client_id = $1
		ORDER BY t.name ASC`

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("query tags by client id: %w", err)
	}
	defer rows.Close()

	tags := make([]domain.Tag, 0)
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(
			&t.ID,
			&t.TeacherID,
			&t.Name,
			&t.SchoolPercent,
			&t.Color,
			&t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan client tag: %w", err)
		}
		tags = append(tags, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return tags, nil
}

// SetClientTags заменяет список тегов клиента на переданные tagIDs.
func (r *TagRepository) SetClientTags(ctx context.Context, clientID uuid.UUID, tagIDs []uuid.UUID) error {
	tx := r.getDBTX(ctx)

	// Удаляем старые теги клиента
	_, err := tx.ExecContext(ctx, "DELETE FROM client_tags WHERE client_id = $1", clientID)
	if err != nil {
		return fmt.Errorf("delete old client tags: %w", err)
	}

	// Вставляем новые теги
	for _, tagID := range tagIDs {
		_, err := tx.ExecContext(ctx, "INSERT INTO client_tags (client_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", clientID, tagID)
		if err != nil {
			return fmt.Errorf("insert client tag: %w", err)
		}
	}

	return nil
}
