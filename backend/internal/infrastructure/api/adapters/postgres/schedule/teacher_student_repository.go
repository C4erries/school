package schedule

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/C4erries/school/backend/internal/domain"
	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/postgres"
)

// TeacherStudentRepository реализует работу с привязками учеников к преподавателям в PostgreSQL.
type TeacherStudentRepository struct {
	db *sql.DB
	sb sq.StatementBuilderType
}

func NewTeacherStudentRepository(db *sql.DB) *TeacherStudentRepository {
	return &TeacherStudentRepository{
		db: db,
		sb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *TeacherStudentRepository) getDBTX(ctx context.Context) postgres.DBTX {
	if tx := postgres.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Create сохраняет привязку ученика к преподавателю.
func (r *TeacherStudentRepository) Create(ctx context.Context, ts *domain.TeacherStudent) error {
	query, args, err := r.sb.Insert("teacher_students").
		Columns("id", "teacher_id", "student_id", "created_at").
		Values(ts.ID, ts.TeacherID, ts.StudentID, ts.CreatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert teacher_student query: %w", err)
	}

	_, err = r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		if postgres.IsUniqueViolation(err) {
			return domain.ErrTeacherStudentAlreadyExists
		}
		return fmt.Errorf("exec insert teacher_student: %w", err)
	}

	return nil
}

// Delete удаляет привязку.
func (r *TeacherStudentRepository) Delete(ctx context.Context, teacherID, studentID uuid.UUID) error {
	query, args, err := r.sb.Delete("teacher_students").
		Where(sq.Eq{"teacher_id": teacherID, "student_id": studentID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete teacher_student query: %w", err)
	}

	res, err := r.getDBTX(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec delete teacher_student: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if rows == 0 {
		return domain.ErrTeacherStudentNotFound
	}

	return nil
}

// IsAssigned проверяет, прикреплен ли ученик к преподавателю.
func (r *TeacherStudentRepository) IsAssigned(ctx context.Context, teacherID, studentID uuid.UUID) (bool, error) {
	query, args, err := r.sb.Select("1").
		From("teacher_students").
		Where(sq.Eq{"teacher_id": teacherID, "student_id": studentID}).
		Limit(1).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("build is assigned query: %w", err)
	}

	var dummy int
	err = r.getDBTX(ctx).QueryRowContext(ctx, query, args...).Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("query is assigned: %w", err)
	}

	return true, nil
}

// ListStudentsByTeacher возвращает список пользователей-учеников, закрепленных за данным преподавателем.
func (r *TeacherStudentRepository) ListStudentsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*domain.User, error) {
	query, args, err := r.sb.Select(
		"u.id", "u.email", "u.password_hash", "u.full_name", "u.phone", "u.role", "u.created_at", "u.updated_at",
	).
		From("teacher_students ts").
		Join("users u ON u.id = ts.student_id").
		Where(sq.Eq{"ts.teacher_id": teacherID}).
		OrderBy("u.full_name ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list students by teacher query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list students by teacher: %w", err)
	}
	defer rows.Close()

	var students []*domain.User
	for rows.Next() {
		var u domain.User
		var role string
		if err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.PasswordHash,
			&u.FullName,
			&u.Phone,
			&role,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan student user: %w", err)
		}
		u.Role = domain.Role(role)
		students = append(students, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return students, nil
}

// ListTeachersByStudent возвращает список пользователей-преподавателей, закрепленных за данным учеником.
func (r *TeacherStudentRepository) ListTeachersByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.User, error) {
	query, args, err := r.sb.Select(
		"u.id", "u.email", "u.password_hash", "u.full_name", "u.phone", "u.role", "u.created_at", "u.updated_at",
	).
		From("teacher_students ts").
		Join("users u ON u.id = ts.teacher_id").
		Where(sq.Eq{"ts.student_id": studentID}).
		OrderBy("u.full_name ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list teachers by student query: %w", err)
	}

	rows, err := r.getDBTX(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list teachers by student: %w", err)
	}
	defer rows.Close()

	var teachers []*domain.User
	for rows.Next() {
		var u domain.User
		var role string
		if err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.PasswordHash,
			&u.FullName,
			&u.Phone,
			&role,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan teacher user: %w", err)
		}
		u.Role = domain.Role(role)
		teachers = append(teachers, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return teachers, nil
}
