package common

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestToPgUUID(t *testing.T) {
	// Test valid UUID
	t.Run("Valid UUID", func(t *testing.T) {
		// Arrange: Create a valid UUID
		validUUID := uuid.New()

		// Act: Call ToPgUUID to convert the valid UUID
		result, err := ToPgUUID(validUUID)

		// Assert: Ensure there is no error and that the result is valid
		require.NoError(t, err)
		require.True(t, result.Valid)
		require.Equal(t, validUUID[:], result.Bytes[:])
	})

	// Test nil UUID (invalid UUID)
	t.Run("Invalid UUID (Nil UUID)", func(t *testing.T) {
		// Arrange: Create a nil UUID
		nilUUID := uuid.Nil

		// Act: Call ToPgUUID to convert the nil UUID
		result, err := ToPgUUID(nilUUID)

		// Assert: Ensure there is an error and that the result is not valid
		require.Error(t, err)
		require.Equal(t, "invalid UUID: cannot be nil", err.Error())
		require.False(t, result.Valid)
	})
}

func TestPtr(t *testing.T) {
	// Test with an int value
	t.Run("Int", func(t *testing.T) {
		// Arrange: Create an int value
		value := 42

		// Act: Call Ptr to get a pointer to the int value
		ptr := Ptr(value)

		// Assert: Ensure the pointer is not nil and points to the correct value
		require.NotNil(t, ptr)
		require.Equal(t, value, *ptr)
	})

	// Test with a string value
	t.Run("String", func(t *testing.T) {
		// Arrange: Create a string value
		value := "Hello, world!"

		// Act: Call Ptr to get a pointer to the string value
		ptr := Ptr(value)

		// Assert: Ensure the pointer is not nil and points to the correct value
		require.NotNil(t, ptr)
		require.Equal(t, value, *ptr)
	})

	// Test with a boolean value
	t.Run("Boolean", func(t *testing.T) {
		// Arrange: Create a boolean value
		value := true

		// Act: Call Ptr to get a pointer to the boolean value
		ptr := Ptr(value)

		// Assert: Ensure the pointer is not nil and points to the correct value
		require.NotNil(t, ptr)
		require.Equal(t, value, *ptr)
	})
}

func TestToPgText(t *testing.T) {
	// Test case where string pointer is not nil
	t.Run("Valid String Pointer", func(t *testing.T) {
		// Arrange: Create a string pointer
		str := "Test String"

		// Act: Convert the string pointer to pgtype.Text
		result := ToPgText(&str)

		// Assert: Verify that the result is valid and contains the correct string
		require.True(t, result.Valid)
		require.Equal(t, str, result.String)
	})

	// Test case where string pointer is nil
	t.Run("Nil String Pointer", func(t *testing.T) {
		// Arrange: Create a nil string pointer
		var str *string

		// Act: Convert the nil string pointer to pgtype.Text
		result := ToPgText(str)

		// Assert: Verify that the result is invalid (Valid should be false)
		require.False(t, result.Valid)
	})
}

func TestToPgTimestamptz(t *testing.T) {
	// Test case where time pointer is not nil
	t.Run("Valid Time Pointer", func(t *testing.T) {
		// Arrange: Create a time pointer
		currentTime := time.Now()

		// Act: Convert the time pointer to pgtype.Timestamptz
		result := ToPgTimestamptz(&currentTime)

		// Assert: Verify that the result is valid and contains the correct time
		require.True(t, result.Valid)
		require.Equal(t, currentTime, result.Time)
	})

	// Test case where time pointer is nil
	t.Run("Nil Time Pointer", func(t *testing.T) {
		// Arrange: Create a nil time pointer
		var m *time.Time

		// Act: Convert the nil time pointer to pgtype.Timestamptz
		result := ToPgTimestamptz(m)

		// Assert: Verify that the result is invalid (Valid should be false)
		require.False(t, result.Valid)
	})
}

func TestToPgInt4(t *testing.T) {
	// Test case where int32 is passed
	t.Run("Valid Int32", func(t *testing.T) {
		// Arrange: Create a valid int32 value
		input := int32(42)

		// Act: Convert the int32 to pgtype.Int4
		result := ToPgInt4(input)

		// Assert: Verify that the result is valid and contains the correct int32 value
		require.True(t, result.Valid)
		require.Equal(t, input, result.Int32)
	})
}

func TestFromPgUUID(t *testing.T) {
	// Test case when pgtype.UUID is valid
	t.Run("Valid UUID", func(t *testing.T) {
		// Arrange: Create a valid pgtype.UUID
		validUUID := uuid.New()
		pgUUID := pgtype.UUID{
			Bytes: validUUID,
			Valid: true,
		}

		// Act: Convert pgtype.UUID to uuid.UUID
		result, err := FromPgUUID(pgUUID)

		// Assert: Verify the result
		require.NoError(t, err)
		require.Equal(t, validUUID, result)
	})

	// Test case when pgtype.UUID is invalid
	t.Run("Invalid UUID", func(t *testing.T) {
		// Arrange: Create an invalid pgtype.UUID
		invalidUUID := pgtype.UUID{Valid: false}

		// Act: Convert pgtype.UUID to uuid.UUID
		result, err := FromPgUUID(invalidUUID)

		// Assert: Verify the result
		require.NoError(t, err)
		require.Equal(t, uuid.Nil, result) // Expect uuid.Nil for invalid UUID
	})
}

func TestFromPgText(t *testing.T) {
	// Test case when pgtype.Text is valid
	t.Run("Valid Text", func(t *testing.T) {
		// Arrange: Create a valid pgtype.Text
		validText := "Hello, world!"
		pgText := pgtype.Text{
			String: validText,
			Valid:  true,
		}

		// Act: Convert pgtype.Text to string pointer
		result := FromPgText(pgText)

		// Assert: Verify the result
		require.NotNil(t, result)
		require.Equal(t, validText, *result)
	})

	// Test case when pgtype.Text is invalid
	t.Run("Invalid Text", func(t *testing.T) {
		// Arrange: Create an invalid pgtype.Text
		invalidText := pgtype.Text{Valid: false}

		// Act: Convert pgtype.Text to string pointer
		result := FromPgText(invalidText)

		// Assert: Verify the result
		require.Nil(t, result)
	})
}

func TestFromPgTimestamptz(t *testing.T) {
	// Test case when pgtype.Timestamptz is valid
	t.Run("Valid Timestamptz", func(t *testing.T) {
		// Arrange: Create a valid pgtype.Timestamptz
		validTime := time.Now()
		pgTime := pgtype.Timestamptz{
			Time:  validTime,
			Valid: true,
		}

		// Act: Convert pgtype.Timestamptz to time pointer
		result := FromPgTimestamptz(pgTime)

		// Assert: Verify the result
		require.NotNil(t, result)
		require.Equal(t, validTime, *result)
	})

	// Test case when pgtype.Timestamptz is invalid
	t.Run("Invalid Timestamptz", func(t *testing.T) {
		// Arrange: Create an invalid pgtype.Timestamptz
		invalidTime := pgtype.Timestamptz{Valid: false}

		// Act: Convert pgtype.Timestamptz to time pointer
		result := FromPgTimestamptz(invalidTime)

		// Assert: Verify the result
		require.Nil(t, result)
	})
}

func TestFromPgInt4(t *testing.T) {
	// Test case when pgtype.Int4 is valid
	t.Run("Valid Int4", func(t *testing.T) {
		// Arrange: Create a valid pgtype.Int4
		validInt32 := int32(42)
		pgInt := pgtype.Int4{
			Int32: validInt32,
			Valid: true,
		}

		// Act: Convert pgtype.Int4 to int32 pointer
		result := FromPgInt4(pgInt)

		// Assert: Verify the result
		require.NotNil(t, result)
		require.Equal(t, validInt32, *result)
	})

	// Test case when pgtype.Int4 is invalid
	t.Run("Invalid Int4", func(t *testing.T) {
		// Arrange: Create an invalid pgtype.Int4
		invalidInt := pgtype.Int4{Valid: false}

		// Act: Convert pgtype.Int4 to int32 pointer
		result := FromPgInt4(invalidInt)

		// Assert: Verify the result
		require.Nil(t, result)
	})
}

func TestToPgUUIDArray(t *testing.T) {
	// Test valid UUIDs
	t.Run("Valid UUID Array", func(t *testing.T) {
		// Arrange: Create multiple valid UUIDs
		validUUIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

		// Act: Call ToPgUUIDArray to convert the valid UUIDs
		result, err := ToPgUUIDArray(validUUIDs)

		// Assert: Ensure there is no error and the array is correctly transformed
		require.NoError(t, err)
		require.Len(t, result, len(validUUIDs))

		// Check each transformed UUID
		for i, pgUUID := range result {
			require.True(t, pgUUID.Valid)
			require.Equal(t, validUUIDs[i][:], pgUUID.Bytes[:])
		}
	})

	// Test empty UUID array (should return nil)
	t.Run("Empty UUID Array", func(t *testing.T) {
		// Arrange: Create an empty UUID slice
		emptyUUIDs := []uuid.UUID{}

		// Act: Call ToPgUUIDArray with an empty slice
		result, err := ToPgUUIDArray(emptyUUIDs)

		// Assert: Ensure there is no error and the result is nil
		require.NoError(t, err)
		require.Nil(t, result)
	})

	// Test invalid UUID in the array (contains uuid.Nil)
	t.Run("Invalid UUID in Array", func(t *testing.T) {
		// Arrange: Create a slice with one valid UUID and one invalid (nil) UUID
		invalidUUIDs := []uuid.UUID{uuid.New(), uuid.Nil}

		// Act: Call ToPgUUIDArray with the invalid array
		result, err := ToPgUUIDArray(invalidUUIDs)

		// Assert: Ensure there is an error and result is nil
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, "invalid UUID in array: invalid UUID: cannot be nil", err.Error())
	})
}
