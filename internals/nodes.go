package internals

// +gen stringer
type ExpressionType int

const (
	ExpressionTypeIntervalValue ExpressionType = iota + 1 // used internally by the parser (not a real expression type)
	ExpressionTypeSeconds
	ExpressionTypeMinutes
	ExpressionTypeHours
	ExpressionTypeDaysOfWeek
	ExpressionTypeDaysOfMonth
	ExpressionTypeDates
)

var s_expressionTypeLen int = len("ExpressionType")

func (e ExpressionType) Name() string {
	str := e.String()
	if len(str) > s_expressionTypeLen {
		str = str[s_expressionTypeLen:]
	}

	return str
}
