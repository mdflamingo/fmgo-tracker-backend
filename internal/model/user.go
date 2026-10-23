package model

import (
	"github.com/google/uuid"
)

type UserDB struct {
	Id       uuid.UUID
	Username string
	Email    string
}

// Identity — атрибуты владельца токена, достаточные для провижининга в "user".
type Identity struct {
	ID       uuid.UUID
	Username string
	Email    *string
}
