# Go Schyntax

[![Build Status](https://travis-ci.org/schyntax/go-schyntax.svg?branch=master)](https://travis-ci.org/schyntax/go-schyntax)
<sup>Supported Schyntax Version: [1.0.1](https://github.com/schyntax/schyntax/tree/v1.0.1)</sup>

[Schyntax](https://github.com/schyntax/schyntax) is a domain-specific language for defining event schedules in a terse, but readable, format. For example, if you want something to run every five minutes during work hours on weekdays, you could write `days(mon..fri) hours(9..<17) min(*%5)`. This project is official implementation of Schyntax for Go.

## Usage

> This library is __NOT__ a scheduled task runner. Most likely, you'll want to use [Schtick](https://github.com/schyntax/go-schtick), which is a scheduled task runner built on top of Schyntax, instead of using this library directly.

To create a new `Schedule` interface:

```go
schedule, err := schyntax.New(`min(*%2)`);
```

You'll need the import statement `import "github.com/schyntax/go-schyntax"`.

Most errors returned will be of interface type `SchyntaxError` which gives you details including the index of any parse errors.

```go
type SchyntaxError interface {
	Error() string
	Input() string
	Index() int
}
```

### Schedule#Next

Returns a `time.Time` representing the next timestamp which matches the scheduling criteria, and an error. The date will always be greater than, never equal to the current time. If no timestamp could be found which matches the scheduling criteria, an error is returned of type `schyntax.ValidTimeNotFoundError`, which indicates there is no time within the next year which matches the schedule.

```go
nextEventTime, err := schedule.Next();
```

There is also a `NextAfter(after time.Time)` method which allows you to search for the next scheduled time relative to a specific time, rather than now.

### Schedule#Previous

Same as `Previous()` accept that its return value will be less than or equal to the current time. This means that if you want to find the last n-previous events, you should subtract at least a millisecond from the result before passing it back to the function.

```go
prevEventTime, err := schedule.Previous(); 
```

There is also a `PreviousAtOrBefore(atOrBefore time.Time)` method.
