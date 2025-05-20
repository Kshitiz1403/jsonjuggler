package utils

import (
	"time"

	isodurationparser "github.com/sosodev/duration"
)

func ParseISODuration(duration string) (time.Duration, error) {
	d, err := isodurationparser.Parse(duration)
	if err != nil {
		return 0, err
	}

	return d.ToTimeDuration(), nil
}
