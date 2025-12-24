package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestErrIsRecordNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "gorm.ErrRecordNotFound",
			err:  gorm.ErrRecordNotFound,
			want: true,
		},
		{
			name: "sql.ErrNoRows",
			err:  sql.ErrNoRows,
			want: true,
		},
		{
			name: "wrapped gorm.ErrRecordNotFound",
			err:  fmt.Errorf("wrapped error: %w", gorm.ErrRecordNotFound),
			want: true,
		},
		{
			name: "wrapped sql.ErrNoRows",
			err:  fmt.Errorf("wrapped error: %w", sql.ErrNoRows),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ErrIsRecordNotFound(tt.err)
			assert.Equal(t, tt.want, result)
		})
	}
}
