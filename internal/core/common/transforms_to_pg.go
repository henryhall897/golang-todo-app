package common

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToPgUUID converts a Go uuid.UUID to a pgtype.UUID with validation.
func ToPgUUID(id uuid.UUID) (pgtype.UUID, error) {
	if id == uuid.Nil {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID: cannot be nil")
	}
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}, nil
}

// ToPgText converts a Go string pointer to a pgtype.Text.
func ToPgText(str *string) pgtype.Text {
	if str != nil {
		return pgtype.Text{
			String: *str,
			Valid:  true,
		}
	}
	return pgtype.Text{
		Valid: false,
	}
}

// ToPgTimestamptz converts a Go time.Time pointer to a pgtype.Timestamptz.
func ToPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t != nil {
		return pgtype.Timestamptz{
			Time:  *t,
			Valid: true,
		}
	}
	return pgtype.Timestamptz{
		Valid: false,
	}
}

// ToPgInt4 converts a Go int32 to a pgtype.Int4.
func ToPgInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: i,
		Valid: true,
	}
}

// ToPgUUIDArray converts a slice of uuid.UUID to a slice of pgtype.UUID.
func ToPgUUIDArray(ids []uuid.UUID) ([]pgtype.UUID, error) {
	if len(ids) == 0 {
		return nil, nil // Returning nil represents NULL for deleting all
	}

	dbIDs := make([]pgtype.UUID, len(ids))

	for i, id := range ids {
		pgUUID, err := ToPgUUID(id)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID in array: %w", err)
		}
		dbIDs[i] = pgUUID
	}

	return dbIDs, nil
}
