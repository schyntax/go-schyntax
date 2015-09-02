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

/**********************************************************************************************
 * Base
**********************************************************************************************/

type NodeBase struct {
	Tokens []*Token
}

func (b *NodeBase) Index() int {
	return b.Tokens[0].Index
}

func (b *NodeBase) AddToken(token *Token) {
	b.Tokens = append(b.Tokens, token)
}

/**********************************************************************************************
 * Node
**********************************************************************************************/

type ProgramNode struct {
	NodeBase
	Groups []*GroupNode
	Expressions []*ExpressionNode
}

func (n *ProgramNode) AddGroup(group *GroupNode) {
	n.Groups = append(n.Groups, group)
}

func (n *ProgramNode) AddExpression(exp *ExpressionNode) {
	n.Expressions = append(n.Expressions, exp)
}

/**********************************************************************************************
 * Group
**********************************************************************************************/

type GroupNode struct {
	NodeBase
	Expressions []*ExpressionNode
}

func (n *GroupNode) AddExpression(exp *ExpressionNode) {
	n.Expressions = append(n.Expressions, exp)
}

/**********************************************************************************************
 * Expression
**********************************************************************************************/

type ExpressionNode struct {
	NodeBase
	ExpressionType ExpressionType
	Arguments []*ArgumentNode
}

func (n *ExpressionNode) AddArgument(arg *ArgumentNode) {
	n.Arguments = append(n.Arguments, arg)
}

/**********************************************************************************************
 * Argument
**********************************************************************************************/

type ArgumentNode struct {
	NodeBase
	IsExclusion bool
	Interval *IntegerValueNode
	IsWildcard bool
	Range *RangeNode
}

func (n *ArgumentNode) HasInterval() bool {
	return n.Interval != nil
}

func (n *ArgumentNode) IntervalValue() int {
	return n.Interval.Value
}

func (n *ArgumentNode) IsRange() bool {
	return n.Range != nil && n.Range.End != nil
}

func (n *ArgumentNode) Value() ValueNode {
	if n.Range == nil {
		return nil
	}

	return n.Range.Start
}

func (n *ArgumentNode) IntervalTokenIndex() int {
	for _, tok := range n.Tokens {
		if tok.Type == TokenTypeInterval {
			return tok.Index
		}
	}

	panic(`IntervalTokenIndex called, but no there are no Interval tokens on this node.`)
}

/**********************************************************************************************
 * Range
**********************************************************************************************/

type RangeNode struct {
	NodeBase
	Start ValueNode
	End ValueNode
	IsHalfOpen bool
}

/**********************************************************************************************
 * Value
**********************************************************************************************/

type ValueNodeType int8
const(
	IntegerValueType ValueNodeType = iota
	DateValueType
)

type ValueNode interface {
	ValueNodeType() ValueNodeType
}

/**********************************************************************************************
 * IntegerValue
**********************************************************************************************/

var _ ValueNode = &IntegerValueNode{}

type IntegerValueNode struct {
	NodeBase
	Value int
}

func (n *IntegerValueNode) ValueNodeType() ValueNodeType {
	return IntegerValueType
}

/**********************************************************************************************
 * DateValue
**********************************************************************************************/

var _ ValueNode = &DateValueNode{}

type DateValueNode struct {
	NodeBase
	HasYear bool
	Year int
	Month int
	Day int
}

func (n *DateValueNode) ValueNodeType() ValueNodeType {
	return DateValueType
}
