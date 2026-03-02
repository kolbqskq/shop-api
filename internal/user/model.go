package user

import (
	"net/mail"
	"shop-api/internal/errs"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	id           uuid.UUID
	email        Email
	passwordHash PasswordHash
	role         Role

	isActive      bool
	emailVerified bool

	lastLoginAt *time.Time

	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
	version   int64
}

type PasswordHash struct {
	value string
}

type Email struct {
	value string
}

type Role string

const (
	Admin     Role = "admin"
	Default   Role = "default"
	Moderator Role = "moderator"
)

func NewUser(id uuid.UUID, email Email, passwordHash PasswordHash, role Role) (*User, error) {
	if role != Default && role != Admin {
		return nil, errs.ErrNoPermission
	}

	return &User{
		id:            id,
		email:         email,
		passwordHash:  passwordHash,
		role:          role,
		isActive:      true,
		emailVerified: false,
	}, nil
}

func NewPasswordHash(password string) (PasswordHash, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return PasswordHash{}, err
	}
	return PasswordHash{
		value: string(hashedPassword),
	}, nil
}

func (c PasswordHash) Compare(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(c.value), []byte(password)); err != nil {
		return errs.ErrInvalidPassword
	}
	return nil
}

func NewEmail(email string) (Email, error) {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return Email{}, errs.ErrInvalidEmail
	}
	email = strings.TrimSpace(strings.ToLower(email))
	return Email{
		value: email,
	}, nil
}

func NewRole(role string) (Role, error) {
	switch role {
	case "default", "admin", "moderator":
		return Role(role), nil
	default:
		return "", errs.ErrInvalidRole
	}

}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Role() Role {
	return u.role
}

func (u *User) IsActive() bool {
	return u.isActive
}

func (u *User) EmailVerified() bool {
	return u.emailVerified
}

func (u *User) PasswordHash() PasswordHash {
	return u.passwordHash
}

func (u *User) ChangePassword(password string) error {
	passwordHash, err := NewPasswordHash(password)
	if err != nil {
		return err
	}
	u.passwordHash = passwordHash
	u.updatedAt = time.Now()
	return nil
}

func (u *User) Login(password string) error {
	if !u.isActive {
		return errs.ErrUserInactive
	}
	if err := u.passwordHash.Compare(password); err != nil {
		return errs.ErrInvalidCredentials
	}

	now := time.Now()
	u.lastLoginAt = &now

	return nil
}

func (u *User) VerifyEmail() {
	u.updatedAt = time.Now()
	u.emailVerified = true
}

func (u *User) Deactivate() {
	u.updatedAt = time.Now()
	u.isActive = false
}

func (u *User) Delete() {
	now := time.Now()
	u.deletedAt = &now
}
