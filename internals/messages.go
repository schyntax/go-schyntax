package internals

import "strings"

const PLEASE_REPORT_BUG_MSG string = " This indicates a bug in Schyntax. Please open an issue on github."

func getStringSnippet(input string, index int) string {
	before := []rune(input[0:index])
	after := []rune(input[index:])

	beforeLen := len(before)
	afterLen := len(after)

	if beforeLen > 20 {
		before = before[beforeLen-20:]
		beforeLen = 20
	}

	if afterLen > 50 {
		after = after[0:50]
		afterLen = 50
	}

	return string(before) + string(after) + "\n" + strings.Repeat(" ", beforeLen) + "^"
}
