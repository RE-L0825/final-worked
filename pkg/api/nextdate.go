package api

import (
	"errors"
	"strconv"
	"strings"
	"time"

)

const dateLayout = "20060102"

func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
    if repeat == "" {
        return "", errors.New("empty repeat rule")
    }

    date, err := time.Parse("20060102", dateStr)
    if err != nil {
        return "", errors.New("invalid date format")
    }

    parts := strings.Fields(repeat)
    if len(parts) == 0 {
        return "", errors.New("invalid repeat format")
    }

    switch parts[0] {
    case "d":
        if len(parts) != 2 {
            return "", errors.New("invalid 'd' format")
        }
        days, err := strconv.Atoi(parts[1])
        if err != nil || days < 1 || days > 400 {
            return "", errors.New("invalid days value")
        }
        return calculateDaily(date, now, days), nil

    case "y":
        return calculateYearly(date, now), nil

    default:
        return "", errors.New("unsupported repeat rule")
    }
}

func calculateDaily(start, now time.Time, days int) string {
	date := start
	for {
		date = date.AddDate(0, 0, days)
		if date.After(now) {
			break
		}
	}
	return date.Format(dateLayout)
}

func calculateYearly(start, now time.Time) string {
	date := start
	for {
		date = date.AddDate(1, 0, 0)
		if date.After(now) {
			break
		}
	}
	return date.Format(dateLayout)
}
