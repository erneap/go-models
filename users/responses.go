package users

import (
	"time"
)

type AuthenticationResponse struct {
	Token       string    `json:"token"`
	UserID      string    `json:"userid,omitempty"`
	Name        UserName  `json:"name"`
	Permissions []string  `json:"permissions,omitempty"`
	PwdExpires  time.Time `json:"pwdExpires,omitempty"`
	Exception   string    `json:"exception"`
}

type UserResponse struct {
	User      User   `json:"user"`
	Exception string `json:"exception"`
}

type TokenRenewalResponse struct {
	Token     string `json:"token"`
	Exception string `json:"exception"`
}

type UsersResponse struct {
	Users     []User `json:"users"`
	Exception string `json:"exception"`
}

type ExceptionResponse struct {
	Exception string `json:"exception"`
}
