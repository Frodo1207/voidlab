package repository

import (
	"database/sql"
	"strings"

	"voidlabai/apps/api/internal/domain"
)

type UserCredential struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	IsActive     bool
}

type UserCreateParams struct {
	Username     string
	PasswordHash string
	Role         string
}

func (r *UserRepository) FindCredentialByUsername(username string) (UserCredential, error) {
	var user UserCredential

	err := r.db.QueryRow(
		`SELECT id, username, password_hash, role, is_active FROM users WHERE username = ? LIMIT 1`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.IsActive)

	if err != nil {
		return UserCredential{}, err
	}

	return user, nil
}

func (r *UserRepository) FindByID(id int64) (domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(
		`SELECT id, username, role, is_active FROM users WHERE id = ? LIMIT 1`,
		id,
	).Scan(&user.ID, &user.Username, &user.Role, &user.IsActive)

	if err != nil {
		return domain.User{}, err
	}

	user.DisplayName = buildDisplayName(user.Role)
	return user, nil
}

func (r *UserRepository) FindDefaultAdmin() (domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(
		`SELECT id, username, role, is_active FROM users WHERE is_active = 1 ORDER BY id ASC LIMIT 1`,
	).Scan(&user.ID, &user.Username, &user.Role, &user.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, err
		}
		return domain.User{}, err
	}

	user.DisplayName = buildDisplayName(user.Role)
	return user, nil
}

func (r *UserRepository) List() ([]domain.ManagedUser, error) {
	rows, err := r.db.Query(`
		SELECT id, username, role, is_active,
		       strftime('%Y-%m-%d %H:%M', created_at),
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM users
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.ManagedUser, 0)
	for rows.Next() {
		var user domain.ManagedUser
		if scanErr := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		); scanErr != nil {
			return nil, scanErr
		}

		user.DisplayName = buildDisplayName(user.Role)
		records = append(records, user)
	}

	return records, rows.Err()
}

func (r *UserRepository) GetManagedByID(id int64) (domain.ManagedUser, error) {
	row := r.db.QueryRow(`
		SELECT id, username, role, is_active,
		       strftime('%Y-%m-%d %H:%M', created_at),
		       strftime('%Y-%m-%d %H:%M', updated_at)
		FROM users
		WHERE id = ?
		LIMIT 1
	`, id)

	var user domain.ManagedUser
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return domain.ManagedUser{}, err
	}

	user.DisplayName = buildDisplayName(user.Role)
	return user, nil
}

func (r *UserRepository) Create(params UserCreateParams) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO users (username, password_hash, role, is_active, updated_at)
		VALUES (?, ?, ?, 1, CURRENT_TIMESTAMP)
	`, params.Username, params.PasswordHash, params.Role)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *UserRepository) UpdateRole(id int64, role string) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET role = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, role, id)
	return err
}

func (r *UserRepository) UpdateStatus(id int64, isActive bool) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, isActive, id)
	return err
}

func (r *UserRepository) UpdatePassword(id int64, passwordHash string) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, passwordHash, id)
	return err
}

func buildDisplayName(role string) string {
	switch strings.TrimSpace(role) {
	case "admin":
		return "VOIDLAB Admin"
	case "editor":
		return "VOIDLAB Editor"
	case "ops":
		return "VOIDLAB Ops"
	default:
		return "VOIDLAB Operator"
	}
}
