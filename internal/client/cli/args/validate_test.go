package args

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateCardNumber(t *testing.T) {
	validated, ok := validateCardNumber("1234 1234 1234 1234")
	require.True(t, ok)
	require.Equal(t, "1234123412341234", validated)

	validated, ok = validateCardNumber("12341234 12341234")
	require.True(t, ok)
	require.Equal(t, "1234123412341234", validated)

	_, ok = validateCardNumber("12341234 12341234 a")
	require.False(t, ok)
}

func TestValidateCvv(t *testing.T) {
	validated, ok := validateCvvCode("123")
	require.True(t, ok)
	require.Equal(t, "123", validated)

	_, ok = validateCvvCode("abc")
	require.False(t, ok)

	_, ok = validateCvvCode("12")
	require.False(t, ok)

	_, ok = validateCvvCode("1245")
	require.False(t, ok)
}

func TestValidateOwnerName(t *testing.T) {
	_, ok := validateCardOwner("Ivan Petrov")
	require.True(t, ok)

	_, ok = validateCardOwner("Ivan")
	require.False(t, ok)

	_, ok = validateCardOwner("IvanPetrov")
	require.False(t, ok)
}

func TestValidateExpDate(t *testing.T) {
	_, ok1 := validateExpDate("2003-12-31")
	require.True(t, ok1)

	_, ok2 := validateExpDate("31-12-2003")
	require.False(t, ok2)
}
