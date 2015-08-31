package internals

type Token struct {
	Type           TokenType
	RawValue       string
	Value          string
	Index          int
	LeadingTrivia  string
	TrailingTrivia string
	ExpressionType ExpressionType
}

// +gen stringer
type TokenType int

const (
	// meta
	TokenTypeNone TokenType = iota
	TokenTypeEndOfInput

	// operators
	TokenTypeRangeInclusive
	TokenTypeRangeHalfOpen
	TokenTypeInterval
	TokenTypeNot
	TokenTypeOpenParen
	TokenTypeCloseParen
	TokenTypeOpenCurly
	TokenTypeCloseCurly
	TokenTypeForwardSlash
	TokenTypeComma
	TokenTypeWildcard

	// alpha-numeric
	TokenTypePositiveInteger
	TokenTypeNegativeInteger
	TokenTypeExpressionName
	TokenTypeDayLiteral
)

var s_tokenTypeLen int = len("TokenType")

func (e TokenType) Name() string {
	str := e.String()
	if len(str) > s_tokenTypeLen {
		str = str[s_tokenTypeLen:]
	}

	return str
}

type TokenQueue struct {
	count int
	head  *TokenQueueNode
	tail  *TokenQueueNode
}

type TokenQueueNode struct {
	token *Token
	next  *TokenQueueNode
}

func (q *TokenQueue) Enqueue(token *Token) {
	node := &TokenQueueNode{token, nil}
	if q.tail == nil {
		q.head = node
		q.tail = node
	} else {
		q.tail.next = node
		q.tail = node
	}

	q.count++
}

func (q *TokenQueue) Dequeue() *Token {
	if q.head == nil {
		panic("Dequeue called on empty queue.")
	}

	token := q.head.token
	q.head = q.head.next
	if q.head == nil {
		q.tail = nil
	}
	q.count--
	return token
}

func (q *TokenQueue) Peek() *Token {
	return q.head.token
}

func (q *TokenQueue) IsEmpty() bool {
	return q.count == 0
}

func (q *TokenQueue) Count() int {
	return q.count
}
