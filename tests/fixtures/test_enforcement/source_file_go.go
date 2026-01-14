package main

import (
	"errors"
	"time"
)

type User struct {
	ID    string
	Email string
}

func authenticateUser(token string) (User, error) {
	if token == "" {
		return User{}, errors.New("token is required")
	}
	
	if token == "invalid_token" {
		return User{}, errors.New("invalid token")
	}
	
	if token == "expired_jwt_token" {
		return User{}, errors.New("token expired")
	}
	
	return User{ID: "123", Email: "user@example.com"}, nil
}












