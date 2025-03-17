package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User roles
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Don't include password in JSON responses
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UserModel handles database operations for users
type UserModel struct {
	DB *sql.DB
}

// NewUserModel creates a new user model
func NewUserModel(db *sql.DB) *UserModel {
	return &UserModel{DB: db}
}

// Create inserts a new user into the database
func (m *UserModel) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (
			id, username, password, email, role, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	// Create a transaction if one isn't provided
	var err error
	if tx == nil {
		tx, err = m.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()
	}

	_, err = tx.ExecContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Password,
		user.Email,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByUsername retrieves a user by username
func (m *UserModel) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, password, email, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &User{}
	err := m.DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (m *UserModel) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, username, password, email, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// AuthenticateUser checks if the provided username and password match a user
func (m *UserModel) AuthenticateUser(ctx context.Context, username, password string) (*User, error) {
	user, err := m.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// UpdatePassword updates a user's password
func (m *UserModel) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		UPDATE users
		SET password = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = m.DB.ExecContext(
		ctx,
		query,
		string(hashedPassword),
		time.Now().UTC(),
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}

// GetAll retrieves all users
func (m *UserModel) GetAll(ctx context.Context) ([]*User, error) {
	query := `
		SELECT id, username, '', email, role, created_at, updated_at
		FROM users
		ORDER BY username
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password, // Empty string as we don't return passwords
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}
