package args

import (
	"time"
	"unicode"

	"github.com/Renal37/goph-keeper/internal/client/storage"
)

func validateCardOwner(owner string) (string, bool) {
	findSpace := false

	for _, r := range owner {
		if unicode.IsLetter(r) {
			continue
		}

		if unicode.IsSpace(r) {
			if findSpace {
				return "", false
			}

			findSpace = true
			continue
		}

		return "", false
	}

	if !findSpace {
		return "", false
	}

	return owner, true
}

func validateExpDate(date string) (time.Time, bool) {
	exp, err := time.Parse(storage.ExpirationFormat, date)
	if err != nil {
		return time.Time{}, false
	}

	return exp, true
}

func validateCvvCode(cvvCode string) (string, bool) {
	if len(cvvCode) == 0 {
		return "Need set card's cvv", false
	}

	if len(cvvCode) != 3 {
		return "Cvv must be a string with 3 digits", false
	}

	for _, r := range cvvCode {
		if !unicode.IsDigit(r) {
			return "Cvv must contain only digits", false
		}
	}

	return cvvCode, true
}

func validateCardNumber(number string) (string, bool) {
	// number validation
	validatedNumber := make([]byte, 0, 16)
	for _, r := range number {
		if unicode.IsDigit(r) {
			validatedNumber = append(validatedNumber, byte(r))
		} else if unicode.IsSpace(r) {
			continue
		} else {
			return "", false
		}
	}

	if len(validatedNumber) != 16 {
		return "", false
	}

	validatedNumberStr := string(validatedNumber)

	return validatedNumberStr, true
}
