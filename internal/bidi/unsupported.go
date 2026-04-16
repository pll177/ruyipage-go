package bidi

import "strings"

func isUnsupportedBiDiCommandError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unknown command") ||
		strings.Contains(message, "not supported") ||
		strings.Contains(message, "unknown method") ||
		strings.Contains(message, "invalid method")
}
