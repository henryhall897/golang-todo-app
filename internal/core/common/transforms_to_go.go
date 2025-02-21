package common

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// FromPgUUID converts a pgtype.UUID to a Go uuid.UUID.
func FromPgUUID(pgUUID pgtype.UUID) (uuid.UUID, error) {
	if !pgUUID.Valid {
		return uuid.Nil, nil
	}
	return uuid.FromBytes(pgUUID.Bytes[:])
}

// FromPgText converts a pgtype.Text to a Go string pointer.
func FromPgText(pgText pgtype.Text) *string {
	if !pgText.Valid {
		return nil
	}
	return &pgText.String
}

// FromPgTimestamptz converts a pgtype.Timestamptz to a Go time.Time pointer.
func FromPgTimestamp(pgTime pgtype.Timestamp) *time.Time {
	if !pgTime.Valid {
		return nil
	}
	t := pgTime.Time.UTC()
	return &t
}

// FromPgInt4 converts a pgtype.Int4 to a Go int32.
func FromPgInt4(pgInt pgtype.Int4) *int32 {
	if !pgInt.Valid {
		return nil
	}
	return &pgInt.Int32
}
