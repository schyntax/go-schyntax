package schyntax

import (
	"github.com/schyntax/go-schyntax/internals"
	"math"
	"time"
)

type Schedule interface {
	OriginalText() string
	NextOrPanic() time.Time
	Next() (time.Time, error)
	NextAfterOrPanic(after time.Time) time.Time
	NextAfter(after time.Time) (time.Time, error)
	PreviousOrPanic() time.Time
	Previous() (time.Time, error)
	PreviousAtOrBeforeOrPanic(atOrBefore time.Time) time.Time
	PreviousAtOrBefore(atOrBefore time.Time) (time.Time, error)
}

var _ Schedule = &scheduleImpl{}

type scheduleImpl struct {
	originalText string
	ir           *internals.IrProgram
}

func New(schedule string) (sch Schedule, err error) {
	defer func() {
		if e := recover(); e != nil {
			sch = nil
			switch e.(type) {
			case string:
				err = newInternalError(e.(string), schedule)
			default:
				if er, ok := e.(error); ok {
					err = er
				} else {
					panic(e) // no idea what was passed to panic, just pass it along
				}
			}
		}
	}()

	parser := internals.NewParser(schedule)
	ast := parser.Parse()

	validator := internals.Validator{schedule, ast}
	validator.AssertValid()

	ir := internals.CompileAst(ast)

	sch = &scheduleImpl{schedule, ir}
	return
}

func (s *scheduleImpl) OriginalText() string {
	return s.originalText
}

func (s *scheduleImpl) NextOrPanic() time.Time {
	t, err := s.Next()
	if err != nil {
		panic(err)
	}

	return t
}

func (s *scheduleImpl) Next() (time.Time, error) {
	return s.getEvent(time.Now(), searchModeAfter)
}

func (s *scheduleImpl) NextAfterOrPanic(after time.Time) time.Time {
	t, err := s.NextAfter(after)
	if err != nil {
		panic(err)
	}

	return t
}

func (s *scheduleImpl) NextAfter(after time.Time) (time.Time, error) {
	return s.getEvent(after, searchModeAfter)
}

func (s *scheduleImpl) PreviousOrPanic() time.Time {
	t, err := s.Previous()
	if err != nil {
		panic(err)
	}

	return t
}

func (s *scheduleImpl) Previous() (time.Time, error) {
	return s.getEvent(time.Now(), searchModeAtOrBefore)
}

func (s *scheduleImpl) PreviousAtOrBeforeOrPanic(atOrBefore time.Time) time.Time {
	t, err := s.PreviousAtOrBefore(atOrBefore)
	if err != nil {
		panic(err)
	}

	return t
}

func (s *scheduleImpl) PreviousAtOrBefore(atOrBefore time.Time) (time.Time, error) {
	return s.getEvent(atOrBefore, searchModeAtOrBefore)
}

type searchMode int8

const (
	searchModeAtOrBefore searchMode = iota
	searchModeAfter
)

func (s *scheduleImpl) getEvent(start time.Time, mode searchMode) (result time.Time, err error) {
	start = start.UTC()
	found := false

	for _, group := range s.ir.Groups {
		if e, good := tryGetGroupEvent(group, start, mode); good {
			if !found || (mode == searchModeAfter && e.Before(result)) || (mode == searchModeAtOrBefore && e.After(result)) {
				result = e
				found = true
			}
		}
	}

	if !found {
		err = &ValidTimeNotFoundError{s.originalText}
	}

	return
}

func tryGetGroupEvent(group *internals.IrGroup, start time.Time, mode searchMode) (result time.Time, found bool) {
	after := mode == searchModeAfter
	inc := 1 // used for incrementing values up or down depending on the direction we're searching
	initHour := 0
	initMinute := 0
	initSecond := 0

	if !after {
		inc = -1
		initHour = 23
		initMinute = 59
		initSecond = 59
	}

	var hourCount, minuteCount, secondCount int

	// todo: make the length of the search configurable
	for d := 0; d < 367; d++ {
		var date time.Time
		var hour, minute, second int
		if d == 0 {
			// "after" events must be in the future
			if after {
				date = start.Add(time.Second)
			} else {
				date = start
			}

			hour = date.Hour()
			minute = date.Minute()
			second = date.Second()
		} else {
			date = start.AddDate(0, 0, d*inc)

			hour = initHour
			minute = initMinute
			second = initSecond
		}

		year := date.Year()
		month := int(date.Month())
		dayOfWeek := int(date.Weekday()) + 1 // Weekday is zero-indexed
		dayOfMonth := date.Day()

		// check if today is an applicable date
		if group.HasDates() {
			applicable := false
			for _, r := range group.Dates {
				if inDateRange(r, year, month, dayOfMonth) {
					applicable = true
					break
				}
			}

			if !applicable {
				goto CONTINUE_DATE_LOOP
			}
		}

		if group.HasDatesExcluded() {
			for _, r := range group.DatesExcluded {
				if inDateRange(r, year, month, dayOfMonth) {
					goto CONTINUE_DATE_LOOP
				}
			}
		}

		// check if date is an applicable day of month
		if group.HasDaysOfMonth() {
			applicable := false
			for _, r := range group.DaysOfMonth {
				if inDayOfMonthRange(r, year, month, dayOfMonth) {
					applicable = true
					break
				}
			}

			if !applicable {
				goto CONTINUE_DATE_LOOP
			}
		}

		if group.HasDaysOfMonthExcluded() {
			for _, r := range group.DaysOfMonthExcluded {
				if inDayOfMonthRange(r, year, month, dayOfMonth) {
					goto CONTINUE_DATE_LOOP
				}
			}
		}

		// check if date is an applicable day of week
		if group.HasDaysOfWeek() && !inRule(7, group.DaysOfWeek, dayOfWeek) {
			goto CONTINUE_DATE_LOOP
		}

		if group.HasDaysOfWeekExcluded() && inRule(7, group.DaysOfWeekExcluded, dayOfWeek) {
			goto CONTINUE_DATE_LOOP
		}

		// if we've gotten this far, then today is an applicable day, let's keep going with hour checks
		if after {
			hourCount = 24 - hour
		} else {
			hourCount = hour + 1
		}

		for hourCount > 0 {
			if group.HasHours() && !inRule(24, group.Hours, hour) {
				goto CONTINUE_HOUR_LOOP
			}

			if group.HasHoursExcluded() && inRule(24, group.HoursExcluded, hour) {
				goto CONTINUE_HOUR_LOOP
			}

			// if we've gotten here, the date and hour are valid. Let's check for minutes
			if after {
				minuteCount = 60 - minute
			} else {
				minuteCount = minute + 1
			}

			for minuteCount > 0 {
				if group.HasMinutes() && !inRule(60, group.Minutes, minute) {
					goto CONTINUE_MINUTE_LOOP
				}

				if group.HasMinutesExcluded() && inRule(60, group.MinutesExcluded, minute) {
					goto CONTINUE_MINUTE_LOOP
				}

				// check for valid seconds
				if after {
					secondCount = 60 - second
				} else {
					secondCount = second + 1
				}

				for secondCount > 0 {
					if group.HasSeconds() && !inRule(60, group.Seconds, second) {
						goto CONTINUE_SECOND_LOOP
					}

					if group.HasSecondsExcluded() && inRule(60, group.SecondsExcluded, second) {
						goto CONTINUE_SECOND_LOOP
					}

					// we've found our event
					result = time.Date(year, time.Month(month), dayOfMonth, hour, minute, second, 0, time.UTC)
					found = true
					return

				CONTINUE_SECOND_LOOP:
					secondCount--
					second += inc
				}

			CONTINUE_MINUTE_LOOP:
				minuteCount--
				minute += inc
				second = initSecond
			}

		CONTINUE_HOUR_LOOP:
			hourCount--
			hour += inc
			minute = initMinute
			second = initSecond
		}

	CONTINUE_DATE_LOOP:
	}

	// we didn't find an applicable date
	return
}

func inRule(lengthOfUnit int, ranges []*internals.IrIntegerRange, value int) bool {
	for _, r := range ranges {
		if inIntegerRange(r, value, lengthOfUnit) {
			return true
		}
	}

	return false
}

func inDateRange(r *internals.IrDateRange, year, month, dayOfMonth int) bool {
	// first, check if this is actually a range
	if !r.IsRange {
		// not a range, so just do a straight comparison
		if r.Start.Month != month || r.Start.Day != dayOfMonth {
			return false
		}

		if r.DatesHaveYear && r.Start.Year != year {
			return false
		}

		return true
	}

	if r.IsHalfOpen {
		// check if this is the last date in a half-open range
		end := r.End
		if end.Day == dayOfMonth && end.Month == month && (!r.DatesHaveYear || end.Year == year) {
			return false
		}
	}

	// check if in-between start and end dates
	if r.DatesHaveYear {
		// when we have a year, the check is much simpler because the range can't be split
		if year < r.Start.Year || year > r.End.Year {
			return false
		}

		if year == r.Start.Year && compareMonthAndDay(month, dayOfMonth, r.Start.Month, r.Start.Day) == -1 {
			return false
		}

		if year == r.End.Year && compareMonthAndDay(month, dayOfMonth, r.End.Month, r.End.Day) == 1 {
			return false
		}

	} else if r.IsSplit { // split ranges aren't allowed to have years (it wouldn't make any sense)
		if month == r.Start.Month || month == r.End.Month {
			if month == r.Start.Month && dayOfMonth < r.Start.Day {
				return false
			}

			if month == r.End.Month && dayOfMonth > r.End.Day {
				return false
			}

		} else if !(month < r.End.Month || month > r.Start.Month) {
			return false
		}

	} else {
		// not a split range, and no year information - just month and day to go on
		if compareMonthAndDay(month, dayOfMonth, r.Start.Month, r.Start.Day) == -1 {
			return false
		}

		if compareMonthAndDay(month, dayOfMonth, r.End.Month, r.End.Day) == 1 {
			return false
		}
	}

	// If we get here, then we're definitely somewhere within the range.
	// If there's no interval, then there's nothing else we need to check
	if !r.HasInterval {
		return true
	}

	// figure out the actual date of the low date so we know whether we're on the desired interval
	var startYear int
	if r.DatesHaveYear {
		startYear = r.Start.Year
	} else if r.IsSplit && month <= r.End.Month {
		// start date is from the previous year
		startYear = year - 1
	} else {
		startYear = year
	}

	startDay := r.Start.Day

	// check if start date was actually supposed to be February 29th, but isn't because of non-leap-year.
	if r.Start.Month == 2 && r.Start.Day == 29 && internals.DaysInMonth(startYear, 2) != 29 {
		// bump the start day back to February 28th so that interval schemes work based on that imaginary date
		// but seriously, people should probably just expect weird results if they're doing something that stupid.
		startDay = 28
	}

	start := time.Date(startYear, time.Month(r.Start.Month), startDay, 0, 0, 0, 0, time.UTC)
	current := time.Date(year, time.Month(month), dayOfMonth, 0, 0, 0, 0, time.UTC)
	dayCount := roundPositiveFloat(current.Sub(start).Hours() / 24.0)

	return (dayCount % r.Interval) == 0
}

// returns 0 if A and B are equal, -1 if A is before B, or 1 if A is after B
func compareMonthAndDay(monthA, dayA, monthB, dayB int) int {
	if monthA == monthB {
		if dayA == dayB {
			return 0
		}

		if dayA > dayB {
			return 1
		}

		return -1
	}

	if monthA > monthB {
		return 1
	}

	return -1
}

func inDayOfMonthRange(r *internals.IrIntegerRange, year, month, dayOfMonth int) bool {
	if r.Start < 0 || (r.IsRange && r.End < 0) {
		// one of the range values is negative, so we need to convert it to a positive by counting back from the end of the month
		daysInMonth := internals.DaysInMonth(year, month)

		revisedStart := r.Start
		if revisedStart < 0 {
			revisedStart = daysInMonth + revisedStart + 1
		}

		revisedEnd := r.End
		if revisedEnd < 0 {
			revisedEnd = daysInMonth + revisedEnd + 1
		}

		r = r.CloneWithRevisedRange(revisedStart, revisedEnd)
	}

	return inIntegerRange(r, dayOfMonth, daysInPreviousMonth(year, month))
}

func inIntegerRange(r *internals.IrIntegerRange, value, lengthOfUnit int) bool {
	if !r.IsRange {
		return value == r.Start
	}

	if r.IsHalfOpen && value == r.End {
		return false
	}

	if r.IsSplit { // range spans across the max value and loops back around
		if value <= r.End || value >= r.Start {
			if r.HasInterval {
				if value >= r.Start {
					return (value-r.Start)%r.Interval == 0
				}

				return (value+lengthOfUnit-r.Start)%r.Interval == 0
			}

			return true
		}

	} else { // not a split range (easier case)
		if value >= r.Start && value <= r.End {
			if r.HasInterval {
				return (value-r.Start)%r.Interval == 0
			}

			return true
		}
	}

	return false
}

func daysInPreviousMonth(year, month int) int {
	month--
	if month == 0 {
		year--
		month = 12
	}

	return internals.DaysInMonth(year, month)
}

func roundPositiveFloat(f float64) int {
	return int(math.Floor(f + 0.5))
}
