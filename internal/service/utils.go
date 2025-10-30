package service

import (
	"errors"

	"github.com/google/uuid"
)

func parseUUID(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.New("invalid UUID format")
	}
	return uid, nil
}
