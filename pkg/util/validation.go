package util

import (
	"fmt"
)

func ValidateThreshold(w, c int64) error {
	if w > c {
		return fmt.Errorf(
			"can not set critical threshold less than warn threshold (warn:%d crit:%d)", w, c,
		)
	}

	return nil
}
