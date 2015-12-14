package internals

import (
	"strconv"
)

func DaysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		// February is weird - default to a leap year, if no year is specified
		if year == 0 || IsLeapYear(year) {
			return 29
		}

		return 28
	}

	panic("Invalid month " + strconv.Itoa(month))
}

func DaysInYear(year int) int {
	if IsLeapYear(year) {
		return 366
	}

	return 365
}

func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
