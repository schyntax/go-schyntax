package internals

/**********************************************************************************************
 * IrProgram
**********************************************************************************************/

type IrProgram struct {
	Groups []*IrGroup
}

func NewIrProgram() *IrProgram {
	return &IrProgram{}
}

/**********************************************************************************************
 * IrGroup
**********************************************************************************************/

type IrGroup struct {
	Seconds             []*IrIntegerRange
	SecondsExcluded     []*IrIntegerRange
	Minutes             []*IrIntegerRange
	MinutesExcluded     []*IrIntegerRange
	Hours               []*IrIntegerRange
	HoursExcluded       []*IrIntegerRange
	DaysOfWeek          []*IrIntegerRange
	DaysOfWeekExcluded  []*IrIntegerRange
	DaysOfMonth         []*IrIntegerRange
	DaysOfMonthExcluded []*IrIntegerRange
	Dates               []*IrDateRange
	DatesExcluded       []*IrDateRange
}

func NewIrGroup() *IrGroup {
	return &IrGroup{}
}

func (ir *IrGroup) HasSeconds() bool {
	return len(ir.Seconds) > 0
}

func (ir *IrGroup) HasSecondsExcluded() bool {
	return len(ir.SecondsExcluded) > 0
}

func (ir *IrGroup) HasMinutes() bool {
	return len(ir.Minutes) > 0
}

func (ir *IrGroup) HasMinutesExcluded() bool {
	return len(ir.MinutesExcluded) > 0
}

func (ir *IrGroup) HasHours() bool {
	return len(ir.Hours) > 0
}

func (ir *IrGroup) HasHoursExcluded() bool {
	return len(ir.HoursExcluded) > 0
}

func (ir *IrGroup) HasDaysOfWeek() bool {
	return len(ir.DaysOfWeek) > 0
}

func (ir *IrGroup) HasDaysOfWeekExcluded() bool {
	return len(ir.DaysOfWeekExcluded) > 0
}

func (ir *IrGroup) HasDaysOfMonth() bool {
	return len(ir.DaysOfMonth) > 0
}

func (ir *IrGroup) HasDaysOfMonthExcluded() bool {
	return len(ir.DaysOfMonthExcluded) > 0
}

func (ir *IrGroup) HasDates() bool {
	return len(ir.Dates) > 0
}

func (ir *IrGroup) HasDatesExcluded() bool {
	return len(ir.DatesExcluded) > 0
}

/**********************************************************************************************
 * IrIntegerRange
**********************************************************************************************/

type IrIntegerRange struct {
	IsRange     bool
	IsHalfOpen  bool
	IsSplit     bool
	Start       int
	End         int
	Interval    int
	HasInterval bool
}

func NewIrIntegerRange(start, end int, hasEnd bool, interval int, isSplit, isHalfOpen bool) *IrIntegerRange {
	ir := &IrIntegerRange{}

	ir.Start = start

	if hasEnd {
		ir.IsSplit = isSplit
		ir.IsHalfOpen = isHalfOpen
		ir.IsRange = true
		ir.End = end
	}

	ir.Interval = interval
	ir.HasInterval = interval != 0

	return ir
}

func (ir IrIntegerRange) CloneWithRevisedRange(start, end int) *IrIntegerRange {
	// receiver is intentionally by value so we start with a copy
	ir.Start = start
	if ir.IsRange {
		ir.End = end
	}

	return &ir
}

/**********************************************************************************************
 * IrDateRange
**********************************************************************************************/

type IrDateRange struct {
	IsRange       bool
	IsHalfOpen    bool
	IsSplit       bool
	Start         *IrDate
	End           *IrDate
	DatesHaveYear bool
	Interval      int
	HasInterval   bool
}

func NewIrDateRange(start *IrDate, end *IrDate, interval int, isSplit, isHalfOpen bool) *IrDateRange {
	ir := &IrDateRange{}

	ir.Start = start
	ir.DatesHaveYear = ir.Start.Year != 0

	if end != nil {
		ir.IsRange = true
		ir.IsSplit = isSplit
		ir.IsHalfOpen = isHalfOpen
		ir.End = end
	}

	ir.Interval = interval
	ir.HasInterval = interval != 0

	return ir
}

/**********************************************************************************************
 * IrDate
**********************************************************************************************/

type IrDate struct {
	Year  int
	Month int
	Day   int
}

func NewIrDate(year, month, day int, hasYear bool) *IrDate {
	ir := &IrDate{year, month, day}
	if !hasYear {
		ir.Year = 0
	}

	return ir
}
