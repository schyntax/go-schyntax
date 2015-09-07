package schyntax

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type testsData struct {
	Dates       expressionTests
	DaysOfMonth expressionTests
	DaysOfWeek  expressionTests
	Hours       expressionTests
	Minutes     expressionTests
	Seconds     expressionTests
}

type expressionTests struct {
	Checks []prevNextCheck
}

type prevNextCheck struct {
	Format string
	Date   time.Time
	Prev   time.Time
	Next   time.Time
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
	for _, check := range tests.Dates.Checks {
		runTest(t, &check)
	}
}

func TestDaysOfMonth(t *testing.T) {
	for _, check := range tests.DaysOfMonth.Checks {
		runTest(t, &check)
	}
}

func TestDaysOfWeek(t *testing.T) {
	for _, check := range tests.DaysOfWeek.Checks {
		runTest(t, &check)
	}
}

func TestHours(t *testing.T) {
	for _, check := range tests.Hours.Checks {
		runTest(t, &check)
	}
}

func TestMinutes(t *testing.T) {
	for _, check := range tests.Minutes.Checks {
		runTest(t, &check)
	}
}

func TestSeconds(t *testing.T) {
	for _, check := range tests.Seconds.Checks {
		runTest(t, &check)
	}
}

func runTest(t *testing.T, check *prevNextCheck) {
	t.Log("Testing " + check.Format)

	sch, err := New(check.Format)
	if err != nil {
		t.Error(err)
		return
	}

	prev, pErr := sch.TryPreviousAtOrBefore(check.Date)
	next, nErr := sch.TryNextAfter(check.Date)

	if pErr != nil {
		t.Error(pErr)
	} else if !prev.Equal(check.Prev) {
		t.Error("Expected: " + check.Prev.String() + ", Actual: " + prev.String())
	} else {
		t.Log("Prev ✓")
	}

	if nErr != nil {
		t.Error(nErr)
	} else if !next.Equal(check.Next) {
		t.Error("Expected: " + check.Next.String() + ", Actual: " + next.String())
	} else {
		t.Log("Next ✓")
	}

}
