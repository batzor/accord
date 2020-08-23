package accord

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User contains user's information
type User struct {
	username        string
	hashedPassword  string
	channelsToRoles map[uint64]Role
}

// NewUser returns a new user
func NewUser(username string, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	user := &User{
		username:        username,
		hashedPassword:  string(hashedPassword),
		channelsToRoles: make(map[uint64]Role),
	}

	return user, nil
}

// IsCorrectPassword checks if the provided password is correct or not
func (user *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.hashedPassword), []byte(password))
	return err == nil
}

// Clone returns a clone of this user
func (user *User) Clone() *User {
	return &User{
		username:        user.username,
		hashedPassword:  user.hashedPassword,
		channelsToRoles: make(map[uint64]Role),
	}
}
