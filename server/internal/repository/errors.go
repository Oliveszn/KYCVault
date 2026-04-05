package repository

// detects unique constraint violations.
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()

	return contains(msg, "duplicate key") ||
		contains(msg, "UNIQUE constraint failed") ||
		contains(msg, "Duplicate entry")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
