package internals

func CompileAst(program *ProgramNode) *IrProgram {
	ir := NewIrProgram()

	// free-floating expressions are placed in an implicit group
	irGroup := compileGroup(program.Expressions)
	if irGroup != nil {
		ir.Groups = append(ir.Groups, irGroup)
	}

	// compile all groups
	for _, groupNode := range program.Groups {
		irGroup = compileGroup(groupNode.Expressions)
		if irGroup != nil {
			ir.Groups = append(ir.Groups, irGroup)
		}
	}

	return ir
}

func compileGroup(expressions []*ExpressionNode) *IrGroup {
	if len(expressions) == 0 {
		return nil
	}

	irGroup := NewIrGroup()

	for _, expression := range expressions {
		compileExpression(irGroup, expression)
	}

	// setup implied rules
	if !irGroup.HasSeconds() && !irGroup.HasSecondsExcluded() { // don't need to setup any defaults if seconds are defined
		if irGroup.HasMinutes() || irGroup.HasMinutesExcluded() {
			irGroup.Seconds = append(irGroup.Seconds, getZeroInteger())
		} else if irGroup.HasHours() || irGroup.HasHoursExcluded() {
			irGroup.Seconds = append(irGroup.Seconds, getZeroInteger())
			irGroup.Minutes = append(irGroup.Minutes, getZeroInteger())
		} else { // only a date level expression was set
			irGroup.Seconds = append(irGroup.Seconds, getZeroInteger())
			irGroup.Minutes = append(irGroup.Minutes, getZeroInteger())
			irGroup.Hours = append(irGroup.Hours, getZeroInteger())
		}
	}

	return irGroup
}

func compileExpression(irGroup *IrGroup, expression *ExpressionNode) {
	for _, arg := range expression.Arguments {
		switch expression.ExpressionType {
		case ExpressionTypeSeconds:
			compileSecondsArgument(irGroup, arg)
		case ExpressionTypeMinutes:
			compileMinutesArgument(irGroup, arg)
		case ExpressionTypeHours:
			compileHoursArgument(irGroup, arg)
		case ExpressionTypeDaysOfWeek:
			compileDaysOfWeekArgument(irGroup, arg)
		case ExpressionTypeDaysOfMonth:
			compileDaysOfMonthArgument(irGroup, arg)
		case ExpressionTypeDates:
			compileDateArgument(irGroup, arg)
		default:
			panic("Expression type " + expression.ExpressionType.Name() + " not supported by the schyntax compiler." + PLEASE_REPORT_BUG_MSG)
		}
	}
}

func compileDateArgument(irGroup *IrGroup, arg *ArgumentNode) {
	var irStart *IrDate = nil
	var irEnd *IrDate = nil
	isSplit := false

	if arg.IsWildcard {
		irStart = NewIrDate(0, 1, 1, false)
		irEnd = NewIrDate(0, 12, 31, false)
	} else {
		start := arg.Range.Start.(*DateValueNode)
		irStart = NewIrDate(start.Year, start.Month, start.Day, start.HasYear)

		if arg.Range.End != nil {
			end := arg.Range.End.(*DateValueNode)
			irEnd = NewIrDate(end.Year, end.Month, end.Day, end.HasYear)
		} else if arg.HasInterval() {
			// if there is an interval, but no end value specified, then the end value is implied
			irEnd = NewIrDate(0, 12, 31, false)
		}

		// check for split range (spans January 1) - not applicable for dates with explicit years
		if irEnd != nil && start.HasYear {
			if irStart.Month >= irEnd.Month && (irStart.Month > irEnd.Month || irStart.Day > irEnd.Day) {
				isSplit = true
			}
		}
	}

	interval := 0
	if arg.HasInterval() {
		interval = arg.IntervalValue()
	}

	isHalfOpen := arg.Range != nil && arg.Range.IsHalfOpen

	irArg := NewIrDateRange(irStart, irEnd, interval, isSplit, isHalfOpen)
	if arg.IsExclusion {
		irGroup.DatesExcluded = append(irGroup.DatesExcluded, irArg)
	} else {
		irGroup.Dates = append(irGroup.Dates, irArg)
	}
}

func compileSecondsArgument(irGroup *IrGroup, arg *ArgumentNode) {
	irArg := compileIntegerArgument(arg, 0, 59)
	if arg.IsExclusion {
		irGroup.SecondsExcluded = append(irGroup.SecondsExcluded, irArg)
	} else {
		irGroup.Seconds = append(irGroup.Seconds, irArg)
	}
}

func compileMinutesArgument(irGroup *IrGroup, arg *ArgumentNode) {
	irArg := compileIntegerArgument(arg, 0, 59)
	if arg.IsExclusion {
		irGroup.MinutesExcluded = append(irGroup.MinutesExcluded, irArg)
	} else {
		irGroup.Minutes = append(irGroup.Minutes, irArg)
	}
}

func compileHoursArgument(irGroup *IrGroup, arg *ArgumentNode) {
	irArg := compileIntegerArgument(arg, 0, 23)
	if arg.IsExclusion {
		irGroup.HoursExcluded = append(irGroup.HoursExcluded, irArg)
	} else {
		irGroup.Hours = append(irGroup.Hours, irArg)
	}
}

func compileDaysOfWeekArgument(irGroup *IrGroup, arg *ArgumentNode) {
	irArg := compileIntegerArgument(arg, 1, 7)
	if arg.IsExclusion {
		irGroup.DaysOfWeekExcluded = append(irGroup.DaysOfWeekExcluded, irArg)
	} else {
		irGroup.DaysOfWeek = append(irGroup.DaysOfWeek, irArg)
	}
}

func compileDaysOfMonthArgument(irGroup *IrGroup, arg *ArgumentNode) {
	irArg := compileIntegerArgument(arg, 1, 31)
	if arg.IsExclusion {
		irGroup.DaysOfMonthExcluded = append(irGroup.DaysOfMonthExcluded, irArg)
	} else {
		irGroup.DaysOfMonth = append(irGroup.DaysOfMonth, irArg)
	}
}

func compileIntegerArgument(arg *ArgumentNode, wildStart, wildEnd int) *IrIntegerRange {
	start := 0
	end := 0
	hasEnd := false
	isSplit := false

	if arg.IsWildcard {
		start = wildStart
		end = wildEnd
		hasEnd = true
	} else {
		start = arg.Range.Start.(*IntegerValueNode).Value
		if arg.Range.End != nil {
			end = arg.Range.End.(*IntegerValueNode).Value
			hasEnd = true
		} else if arg.HasInterval() { // if there is an interval, but no end value specified, then the end value is implied
			end = wildEnd
			hasEnd = true
		}

		// check for a split range
		if hasEnd && end < start {
			// Start is greater than end, so it's probably a split range, but there's one exception.
			// If this is a month expression, and end is a negative number (counting back from the end of the month)
			// then it might not actually be a split range
			if start < 0 || end > 0 {
				// check says that either start is negative or end is positive
				// (means we're probably not in the weird day of month scenario)
				// todo: implement a better check which looks for possible overlap between a positive start and negative end
				isSplit = true
			}
		}
	}

	interval := 0
	if arg.HasInterval() {
		interval = arg.IntervalValue()
	}

	isHalfOpen := arg.Range != nil && arg.Range.IsHalfOpen

	return NewIrIntegerRange(start, end, hasEnd, interval, isSplit, isHalfOpen)
}

func getZeroInteger() *IrIntegerRange {
	return NewIrIntegerRange(0, 0, false, 0, false, false)
}
