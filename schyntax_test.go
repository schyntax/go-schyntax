package schyntax

import (
	"encoding/json"
	"github.com/schyntax/go-schyntax/internals"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type testsData struct {
	TestsVersion int
	Hash         string
	Suites       suites
}

type suites struct {
	Dates          []check
	DaysOfMonth    []check
	DaysOfWeek     []check
	Hours          []check
	Minutes        []check
	Seconds        []check
	SyntaxErrors   []check
	ArgumentErrors []check
	Commas         []check
}

type check struct {
	Format          string
	Date            time.Time
	Prev            *time.Time
	Next            *time.Time
	ParseErrorIndex *int
}

var tests testsData

func TestMain(m *testing.M) {
	file, err := ioutil.ReadFile("tests.json")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(file, &tests); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestDates(t *testing.T) {
	for _, check := range tests.Suites.Dates {
		runTest(t, &check)
	}
}

func TestDaysOfMonth(t *testing.T) {
	for _, check := range tests.Suites.DaysOfMonth {
		runTest(t, &check)
	}
}

func TestDaysOfWeek(t *testing.T) {
	for _, check := range tests.Suites.DaysOfWeek {
		runTest(t, &check)
	}
}

func TestHours(t *testing.T) {
	for _, check := range tests.Suites.Hours {
		runTest(t, &check)
	}
}

func TestMinutes(t *testing.T) {
	for _, check := range tests.Suites.Minutes {
		runTest(t, &check)
	}
}

func TestSeconds(t *testing.T) {
	for _, check := range tests.Suites.Seconds {
		runTest(t, &check)
	}
}

func TestSyntaxErrors(t *testing.T) {
	for _, check := range tests.Suites.SyntaxErrors {
		runTest(t, &check)
	}
}

func TestArgumentErrors(t *testing.T) {
	for _, check := range tests.Suites.ArgumentErrors {
		runTest(t, &check)
	}
}

func TestCommas(t *testing.T) {
	for _, check := range tests.Suites.Commas {
		runTest(t, &check)
	}
}

func runTest(t *testing.T, check *check) {
	t.Log(`Testing "` + check.Format + `" - Start ` + check.Date.String())

	sch, err := New(check.Format)
	if err != nil {
		if check.ParseErrorIndex != nil {
			// we were expecting an error
			if parseError, ok := err.(*internals.ParseError); ok {
				if parseError.Index() == *check.ParseErrorIndex {
					t.Log("Expected Parse Error ✓")
				} else {
					t.Errorf("Wrong parse error index. Expected: %d. Actual: %d.\n", *check.ParseErrorIndex, parseError.Error())
				}

				return
			}
		}

		t.Error(err)
		return
	}

	if check.ParseErrorIndex != nil {
		t.Errorf("Expected a parse error at index %d, but no error was thrown.", *check.ParseErrorIndex)
		return
	}

	prev, pErr := sch.PreviousAtOrBefore(check.Date)

	if pErr != nil {
		if _, ok := pErr.(*ValidTimeNotFoundError); ok && check.Prev == nil {
			t.Log("Prev ✓ (ValidTimeNotFoundError)")
		} else {
			t.Error(pErr)
		}
	} else if check.Prev == nil {
		t.Error("Expected a ValidTimeNotFoundError. Date returned from previous: " + prev.String())
	} else if !prev.Equal(*check.Prev) {
		t.Error("Expected: " + check.Prev.String() + ", Actual: " + prev.String())
	} else {
		t.Log("Prev ✓")
	}

	next, nErr := sch.NextAfter(check.Date)

	if nErr != nil {
		if _, ok := nErr.(*ValidTimeNotFoundError); ok && check.Next == nil {
			t.Log("Prev ✓ (ValidTimeNotFoundError)")
		} else {
			t.Error(nErr)
		}
	} else if check.Next == nil {
		t.Error("Expected a ValidTimeNotFoundError. Date returned from next: " + next.String())
	} else if !next.Equal(*check.Next) {
		t.Error("Expected: " + check.Next.String() + ", Actual: " + next.String())
	} else {
		t.Log("Next ✓")
	}

}
