package util

import (
	"fmt"

	"github.com/makocchi-git/kubectl-free/pkg/constants"
)

func ValidateThreshold(w, c int64) error {
	if w > c {
		return fmt.Errorf(
			"can not set critical threshold less than warn threshold (warn:%d crit:%d)", w, c,
		)
	}

	return nil
}

func ValidateHeaderOpt(opt string) error {
	for _, v := range constants.ValidheaderOptions {
		if v == opt {
			return nil
		}
	}

	return fmt.Errorf("invalid header option: %s", opt)
}
