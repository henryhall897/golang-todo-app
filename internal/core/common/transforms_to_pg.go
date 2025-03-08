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
		// Handler is supposed for this. This is a safety net if a layer is skipped. this should never happen.
		return pgtype.UUID{}, ErrInvalidUUID
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
func ToPgTimestamp(t *time.Time) pgtype.Timestamp {
	if t != nil {
		return pgtype.Timestamp{
			Time:  t.UTC(),
			Valid: true,
		}
	}
	return pgtype.Timestamp{
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
