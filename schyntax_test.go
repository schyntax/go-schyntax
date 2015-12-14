package schyntax

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/schyntax/go-schyntax/internals"
)

type testsData struct {
	TestsVersion int
	Hash         string
	Suites       map[string][]*check
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
		log.Fatal(err)
	}

	if err = json.Unmarshal(file, &tests); err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestSuites(t *testing.T) {
	for name, checks := range tests.Suites {
		t.Logf("-----SUITE %s-----", name)
		for _, check := range checks {
			runTest(t, check)
		}
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
					logNonVerbose(t, "Expected Parse Error ✓")
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
			logNonVerbose(t, "Prev ✓ (ValidTimeNotFoundError)")
		} else {
			t.Error(pErr)
		}
	} else if check.Prev == nil {
		t.Error("Expected a ValidTimeNotFoundError. Date returned from previous: " + prev.String())
	} else if !prev.Equal(*check.Prev) {
		t.Error("Expected: " + check.Prev.String() + ", Actual: " + prev.String())
	} else {
		logNonVerbose(t, "Prev ✓")
	}

	next, nErr := sch.NextAfter(check.Date)

	if nErr != nil {
		if _, ok := nErr.(*ValidTimeNotFoundError); ok && check.Next == nil {
			logNonVerbose(t, "Prev ✓ (ValidTimeNotFoundError)")
		} else {
			t.Error(nErr)
		}
	} else if check.Next == nil {
		t.Error("Expected a ValidTimeNotFoundError. Date returned from next: " + next.String())
	} else if !next.Equal(*check.Next) {
		t.Error("Expected: " + check.Next.String() + ", Actual: " + next.String())
	} else {
		logNonVerbose(t, "Next ✓")
	}

}

func logNonVerbose(t *testing.T, msg string) {
	if testing.Verbose() {
		t.Log(msg)
	}
}
