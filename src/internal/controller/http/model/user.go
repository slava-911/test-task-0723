package cmodel

import (
	"errors"

	dmodel "github.com/slava-911/test-task-0723/internal/domain/model"
)

type SignInUserDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserDTO struct {
	FirstName      string `json:"firstname" validate:"required,min=2"`
	LastName       string `json:"lastname" validate:"required,min=2"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	RepeatPassword string `json:"repeat_password" validate:"required,min=8"`
	Age            int    `json:"age" validate:"required,min=18"`
	IsMarried      bool   `json:"is_married"`
}

func (d *CreateUserDTO) ToUser() *dmodel.User {
	return &dmodel.User{
		FirstName: d.FirstName,
		LastName:  d.LastName,
		Email:     d.Email,
		Password:  d.Password,
		Age:       d.Age,
		IsMarried: d.IsMarried,
	}
}

type UpdateUserDTO struct {
	FirstName   *string `json:"firstname,omitempty"`
	LastName    *string `json:"lastname,omitempty"`
	Email       *string `json:"email,omitempty"`
	OldPassword *string `json:"old_password,omitempty"`
	NewPassword *string `json:"new_password,omitempty"`
	IsMarried   *bool   `json:"is_married,omitempty"`
}

func (d *CreateUserDTO) ValidatePassword() error {
	if d.Password != d.RepeatPassword {
		return errors.New("password does not match repeat password")
	}
	return nil
}
