// Generated by: main.exe
// TypeWriter: stringer
// Directive: +gen on ContextMode

package internals

import (
	"fmt"
)

const _ContextMode_name = "ContextModeProgramContextModeGroupContextModeExpression"

var _ContextMode_index = [...]uint8{0, 18, 34, 55}

func (i ContextMode) String() string {
	if i < 0 || i+1 >= ContextMode(len(_ContextMode_index)) {
		return fmt.Sprintf("ContextMode(%d)", i)
	}
	return _ContextMode_name[_ContextMode_index[i]:_ContextMode_index[i+1]]
}
