package utils

import (
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

func ErrIsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows)
}
