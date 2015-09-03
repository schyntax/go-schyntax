package internals

import (
	"fmt"
	"strconv"
	"strings"
)

type Validator struct {
	Input   string
	Program *ProgramNode
}

func (v *Validator) AssertValid() {
	v.assertProgram(v.Program)
}

func (v *Validator) assertProgram(program *ProgramNode) {
	if len(program.Expressions) == 0 {
		// no free-floating expressions, so we need to make sure there is at least one group with an expression
		hasExpressions := false
		for _, group := range program.Groups {
			if len(group.Expressions) > 0 {
				hasExpressions = true
				break
			}
		}

		if !hasExpressions {
			panic("Schedule must contain at least one expression.")
		}
	}

	for _, group := range program.Groups {
		v.assertGroup(group)
	}

	v.assertExpressionList(program.Expressions)
}

func (v *Validator) assertGroup(group *GroupNode) {
	v.assertExpressionList(group.Expressions)
}

func (v *Validator) assertExpressionList(expressions []*ExpressionNode) {
	for _, expression := range expressions {
		v.assertExpression(expression)
	}
}

func (v *Validator) assertExpression(expression *ExpressionNode) {
	if len(expression.Arguments) == 0 {
		panic("Expression has no arguments.\n\n" + getStringSnippet(v.Input, expression.Index()))
	}

	for _, arg := range expression.Arguments {
		if arg.HasInterval() && arg.IntervalValue() == 0 {
			panic(`"%0" is not a valid interval. If your intention was to include all ` +
				expressionTypeToHumanString(expression.ExpressionType) + ` use the wildcard operator "*" instead of an interval` +
				"\n\n" + getStringSnippet(v.Input, arg.IntervalTokenIndex()))
		}

		validator := v.getValidator(expression.ExpressionType)

		if arg.IsWildcard {
			if arg.IsExclusion && !arg.HasInterval() {
				panic("Wildcards can't be excluded with the ! operator, except when part of an interval (using %).\n\n" + getStringSnippet(v.Input, arg.Index()))
			}
		} else {
			if arg.Range == nil || arg.Range.Start == nil {
				panic("Expected a value or range.\n\n" + getStringSnippet(v.Input, arg.Index()))
			}

			v.assertRange(expression.ExpressionType, arg.Range, validator)
		}

		if arg.HasInterval() {
			validator(ExpressionTypeIntervalValue, arg.Interval)
		}
	}
}

func (v *Validator) getValidator(expType ExpressionType) ValueValidator {
	switch expType {
	case ExpressionTypeSeconds, ExpressionTypeMinutes:
		return v.secondOrMinute
	case ExpressionTypeHours:
		return v.hour
	case ExpressionTypeDaysOfWeek:
		return v.dayOfWeek
	case ExpressionTypeDaysOfMonth:
		return v.dayOfMonth
	case ExpressionTypeDates:
		return v.date
	default:
		panic("ExpressionType " + expType.Name() + " has not been implemented by the validator." + PLEASE_REPORT_BUG_MSG)
	}
}

type ValueValidator func(expType ExpressionType, value ValueNode)

func (v *Validator) assertRange(expType ExpressionType, rangeNode *RangeNode, validator ValueValidator) {
	validator(expType, rangeNode.Start)
	if rangeNode.End != nil {
		validator(expType, rangeNode.End)

		if rangeNode.IsHalfOpen && valuesAreEqual(expType, rangeNode.Start, rangeNode.End) {
			panic("Start and end values of a half-open range cannot be equal.\n\n" + getStringSnippet(v.Input, rangeNode.Start.Index()))
		}
	}

	if expType == ExpressionTypeDates && rangeNode.End != nil {
		// special validation to make the date range is sane
		start := rangeNode.Start.(*DateValueNode)
		end := rangeNode.End.(*DateValueNode)

		if start.HasYear || end.HasYear {
			if !start.HasYear || !end.HasYear {
				panic("Cannot mix full and partial dates in a date range.\n\n" + getStringSnippet(v.Input, start.Index()))
			}

			if !v.isStartBeforeEnd(start, end) {
				panic("End date of range is before the start date.\n\n" + getStringSnippet(v.Input, start.Index()))
			}
		}
	}
}

func (v *Validator) secondOrMinute(expType ExpressionType, value ValueNode) {
	v.integerValue(expType, value, 0, 59)
}

func (v *Validator) hour(expType ExpressionType, value ValueNode) {
	v.integerValue(expType, value, 0, 23)
}

func (v *Validator) dayOfWeek(expType ExpressionType, value ValueNode) {
	v.integerValue(expType, value, 1, 7)
}

func (v *Validator) dayOfMonth(expType ExpressionType, value ValueNode) {
	ival := v.integerValue(expType, value, -31, 31)
	if ival == 0 {
		panic("Day of month cannot be zero.\n\n" + getStringSnippet(v.Input, value.Index()))
	}
}

func (v *Validator) date(expType ExpressionType, value ValueNode) {
	date := value.(*DateValueNode)

	if date.HasYear {
		if date.Year < 1900 || date.Year > 2200 {
			panic("Year " + strconv.Itoa(date.Year) + " is not a valid year. Must be between 1900 and 2200.\n\n" + getStringSnippet(v.Input, date.Index()))
		}
	}

	if date.Month < 1 || date.Month > 12 {
		panic("Month " + strconv.Itoa(date.Month) + " is not a valid month. Must be between 1 and 12.\n\n" + getStringSnippet(v.Input, date.Index()))
	}

	var effectiveYear int
	if date.HasYear {
		effectiveYear = date.Year
	} else {
		effectiveYear = 0
	}
	days := daysInMonth(effectiveYear, date.Month)
	if date.Day < 1 || date.Day > days {
		panic(strconv.Itoa(date.Day) + " is not a valid day for the month specified. Must be between 1 and " + strconv.Itoa(days) + getStringSnippet(v.Input, date.Index()))
	}
}

func (v *Validator) integerValue(expType ExpressionType, value ValueNode, min, max int) int {
	ival := value.(*IntegerValueNode).Value
	if ival < min || ival > max {
		msg := fmt.Sprintf("%v cannot be %v. Value must be between %v and %v.\n\n%v",
			expressionTypeToHumanString(expType), ival, min, max, getStringSnippet(v.Input, value.Index()))
		panic(msg)
	}

	return ival
}

func valuesAreEqual(expType ExpressionType, a, b ValueNode) bool {
	if expType == ExpressionTypeDates {
		ad := a.(*DateValueNode)
		bd := b.(*DateValueNode)

		if ad.Day != bd.Day || ad.Month != bd.Month {
			return false
		}

		if ad.HasYear && ad.Year != bd.Year {
			return false
		}

		return true
	}

	// integer values
	ai := a.(*IntegerValueNode).Value
	bi := b.(*IntegerValueNode).Value

	return ai == bi
}

func (v *Validator) isStartBeforeEnd(start, end *DateValueNode) bool {
	if start.Year < end.Year {
		return true
	}

	if start.Year > end.Year {
		return false
	}

	// must be the same start and end year if we get here

	if start.Month < end.Month {
		return true
	}

	if start.Month > end.Month {
		return false
	}

	// must be the same month

	return start.Day <= end.Day
}

func expressionTypeToHumanString(expType ExpressionType) string {
	switch expType {
	case ExpressionTypeDaysOfMonth:
		return "days of the month"
	case ExpressionTypeDaysOfWeek:
		return "days of the week"
	case ExpressionTypeIntervalValue:
		return "interval"
	default:
		return strings.ToLower(expType.Name())
	}
}
