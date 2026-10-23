package model

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// UnknownUsersError — в запрос переданы id пользователей, которых нет в "user".
type UnknownUsersError struct {
	IDs []uuid.UUID
}

func (e *UnknownUsersError) Error() string {
	parts := make([]string, 0, len(e.IDs))
	for _, id := range e.IDs {
		parts = append(parts, id.String())
	}
	return fmt.Sprintf("unknown user ids: %s", strings.Join(parts, ", "))
}
