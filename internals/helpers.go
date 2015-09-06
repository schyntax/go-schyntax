package internals

import (
	"strconv"
	"time"
)

func DaysInMonth(year, month int) int {
	if month == 2 { // February is weird
		if year == 0 {
			return 29 // default to a leap year, if no year is specified
		}

		if time.Date(year, 2, 29, 0, 0, 0, 0, time.UTC).Day() == 29 {
			return 29
		}

		return 28
	}

	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	}

	panic("Invalid month " + strconv.Itoa(month))
}
