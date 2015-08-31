package internals

import (
	"regexp"
)

// literal terminals
var TermsRangeInclusive *Terminal = &Terminal{TokenTypeRangeInclusive, "..", nil, 0}
var TermsRangeHalfOpen *Terminal = &Terminal{TokenTypeRangeHalfOpen, "..<", nil, 0}
var TermsInterval *Terminal = &Terminal{TokenTypeInterval, "%", nil, 0}
var TermsNot *Terminal = &Terminal{TokenTypeNot, "!", nil, 0}
var TermsOpenParen *Terminal = &Terminal{TokenTypeOpenParen, "(", nil, 0}
var TermsCloseParen *Terminal = &Terminal{TokenTypeCloseParen, ")", nil, 0}
var TermsOpenCurly *Terminal = &Terminal{TokenTypeOpenCurly, "{", nil, 0}
var TermsCloseCurly *Terminal = &Terminal{TokenTypeCloseCurly, "}", nil, 0}
var TermsForwardSlash *Terminal = &Terminal{TokenTypeForwardSlash, "/", nil, 0}
var TermsComma *Terminal = &Terminal{TokenTypeComma, ",", nil, 0}
var TermsWildcard *Terminal = &Terminal{TokenTypeWildcard, "*", nil, 0}

// regex terminals
var TermsPositiveInteger *Terminal = &Terminal{TokenTypePositiveInteger, "", regexp.MustCompile(`^[0-9]+`), 0}
var TermsNegativeInteger *Terminal = &Terminal{TokenTypeNegativeInteger, "", regexp.MustCompile(`^-[0-9]+`), 0}

var TermsSunday *Terminal = &Terminal{TokenTypeDayLiteral, "SUNDAY", regexp.MustCompile(`(?i)^(su|sun|sunday)(?:\b)`), 0}
var TermsMonday *Terminal = &Terminal{TokenTypeDayLiteral, "MONDAY", regexp.MustCompile(`(?i)^(mo|mon|monday)(?:\b)`), 0}
var TermsTuesday *Terminal = &Terminal{TokenTypeDayLiteral, "TUESDAY", regexp.MustCompile(`(?i)^(tu|tue|tuesday|tues)(?:\b)`), 0}
var TermsWednesday *Terminal = &Terminal{TokenTypeDayLiteral, "WEDNESDAY", regexp.MustCompile(`(?i)^(we|wed|wednesday)(?:\b)`), 0}
var TermsThursday *Terminal = &Terminal{TokenTypeDayLiteral, "THURSDAY", regexp.MustCompile(`(?i)^(th|thu|thursday|thur|thurs)(?:\b)`), 0}
var TermsFriday *Terminal = &Terminal{TokenTypeDayLiteral, "FRIDAY", regexp.MustCompile(`(?i)^(fr|fri|friday)(?:\b)`), 0}
var TermsSaturday *Terminal = &Terminal{TokenTypeDayLiteral, "SATURDAY", regexp.MustCompile(`(?i)^(sa|sat|saturday)(?:\b)`), 0}

var TermsSeconds *Terminal = &Terminal{TokenTypeExpressionName, "", regexp.MustCompile(`(?i)^(s|sec|second|seconds|secondofminute|secondsofminute)(?:\b)`), ExpressionTypeSeconds}
var TermsMinutes *Terminal = &Terminal{TokenTypeExpressionName, "", regexp.MustCompile(`(?i)^(m|min|minute|minutes|minuteofhour|minutesofhour)(?:\b)`), ExpressionTypeMinutes}
var TermsHours *Terminal = &Terminal{TokenTypeExpressionName, "", regexp.MustCompile(`(?i)^(h|hour|hours|hourofday|hoursofday)(?:\b)`), ExpressionTypeHours}
var TermsDaysOfWeek *Terminal = &Terminal{TokenTypeExpressionName, "", regexp.MustCompile(`(?i)^(day|days|dow|dayofweek|daysofweek)(?:\b)`), ExpressionTypeDaysOfWeek}
var TermsDaysOfMonth *Terminal = &Terminal{TokenTypeExpressionName, "", regexp.MustCompile(`(?i)^(dom|dayofmonth|daysofmonth)(?:\b)`), ExpressionTypeDaysOfMonth}
var TermsDates *Terminal = &Terminal{TokenTypeExpressionName, "", regexp.MustCompile(`(?i)^(date|dates)(?:\b)`), ExpressionTypeDates}

type Terminal struct {
	TokenType      TokenType
	Value          string
	Regex          *regexp.Regexp
	ExpressionType ExpressionType
}

func (t *Terminal) GetToken(input string, index int) *Token {
	if t.Regex == nil {
		valLen := len(t.Value)
		if len(input)-index < valLen {
			return nil
		}

		for i := 0; i < valLen; i++ {
			if input[index+i] != t.Value[i] {
				return nil
			}
		}

		return t.createToken(index, t.Value, "")
	}

	match := t.Regex.FindString(input[index:])
	if match == "" {
		return nil
	}

	return t.createToken(index, match, match)
}

func (t *Terminal) createToken(index int, raw string, value string) *Token {
	tok := &Token{}
	tok.Type = t.TokenType
	tok.Index = index
	tok.RawValue = raw
	if t.Value != "" {
		tok.Value = t.Value
	} else {
		tok.Value = value
	}

	tok.ExpressionType = t.ExpressionType

	return tok
}
