package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user with the given details
func (m *UserModel) CreateUser(ctx context.Context, username, email, password, role string) error {
	// Validate role
	if role != RoleAdmin && role != RoleUser {
		return errors.New("invalid role")
	}

	// Check if username already exists
	existingUser, err := m.GetByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("error checking existing username: %w", err)
	}
	if existingUser != nil {
		return errors.New("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	now := time.Now().UTC()
	user := &User{
		ID:        uuid.New().String(),
		Username:  username,
		Password:  string(hashedPassword),
		Email:     email,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save user to database
	err = m.Create(ctx, nil, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Authenticate verifies a user's credentials and returns the user if valid
func (m *UserModel) Authenticate(ctx context.Context, username, password string) (*User, error) {
	// Get user by username
	user, err := m.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	// No user found with that username
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	// Compare password with stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errors.New("invalid username or password")
		}
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	return user, nil
}

// UpdateUser updates a user's details
func (m *UserModel) UpdateUser(ctx context.Context, id, username, email, role string) error {
	// Get the user first
	user, err := m.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if username is being changed and if so, check it's not already taken
	if username != user.Username {
		existingUser, err := m.GetByUsername(ctx, username)
		if err != nil {
			return fmt.Errorf("error checking existing username: %w", err)
		}
		if existingUser != nil {
			return errors.New("username already exists")
		}
	}

	// Update user fields
	user.Username = username
	user.Email = email
	user.Role = role
	user.UpdatedAt = time.Now().UTC()

	// Save changes
	err = m.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdatePassword updates a user's password
func (m *UserModel) UpdatePassword(ctx context.Context, id, password string) error {
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update the password in the database
	query := `
		UPDATE users
		SET password = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = m.DB.ExecContext(ctx, query, string(hashedPassword), time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}