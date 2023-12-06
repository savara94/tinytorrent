package torrent

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

type testCase struct {
	reader    io.Reader
	want      interface{}
	wantedErr error
}

type testCmpFn func(got, want interface{}) bool

func simpleCmp(got, want interface{}) bool {
	return got == want
}

func sliceCmp(got, want interface{}) bool {
	return reflect.DeepEqual(got, want)
}

func TestParseBencode(t *testing.T) {
	generalTestCases := []testCase{
		// No valid delimiter
		{strings.NewReader("psa"), nil, io.EOF},
	}

	numberTestCases := []testCase{
		// Good weather cases
		{strings.NewReader("i45e"), 45, nil},
		{strings.NewReader("i567e"), 567, nil},
		{strings.NewReader("i-45e"), -45, nil},
		// Missing number
		{strings.NewReader("ie"), nil, ErrExpectedNumber},
		// Missing 'e'
		{strings.NewReader("i456"), nil, io.EOF},
	}

	stringParsingTestCases := []testCase{
		// Good weather case
		{strings.NewReader("3:ben"), "ben", nil},
		// Less characters than specified
		{strings.NewReader("4:ben"), nil, ErrEndOfReader},
		// More characters than specified
		{strings.NewReader("2:ben"), nil, ErrEndExpected},
		// 0 characters specified
		{strings.NewReader("0:"), "", nil},
		// 1 character specified
		{strings.NewReader("1:a"), "a", nil},
	}

	listParsingTestCases := []testCase{
		// Good weather cases
		{strings.NewReader("l3:ben2:goe"), []interface{}{"ben", "go"}, nil},
		{strings.NewReader("l3:beni56ee"), []interface{}{"ben", 56}, nil},
		// Empty list
		{strings.NewReader("le"), []interface{}{}, nil},
		// Missing end of list
		{strings.NewReader("l"), nil, io.EOF},
	}

	dictParsingTestCases := []testCase{
		// Good weather cases
		{strings.NewReader("d3:ben2:goe"), map[string]interface{}{"ben": "go"}, nil},
		{strings.NewReader("d3:beni56ee"), map[string]interface{}{"ben": 56}, nil},
		// Empty dict
		{strings.NewReader("de"), map[string]interface{}{}, nil},
		// Missing end of dict
		{strings.NewReader("d"), nil, io.EOF},
	}

	runTestCases("general", generalTestCases, t, simpleCmp)
	runTestCases("numbers", numberTestCases, t, simpleCmp)
	runTestCases("strings", stringParsingTestCases, t, simpleCmp)
	runTestCases("lists", listParsingTestCases, t, sliceCmp)
	runTestCases("dicts", dictParsingTestCases, t, sliceCmp)

	complexTestCases := []testCase{
		// Nested lists
		{strings.NewReader("llelleei0ee"), []interface{}{[]interface{}{}, []interface{}{[]interface{}{}}, 0}, nil},
		// Nested dicts
		{strings.NewReader("d3:bend3:ben3:kenee"), map[string]interface{}{"ben": map[string]interface{}{"ben": "ken"}}, nil},
	}

	runTestCases("complex", complexTestCases, t, sliceCmp)
}

func runTestCases(category string, testCases []testCase, t *testing.T, cmpFn testCmpFn) {
	for i := range testCases {
		testCase := testCases[i]

		reader := testCase.reader
		want := testCase.want
		wantedErr := testCase.wantedErr

		got, err := ParseBencode(reader)

		assertError(category, i, wantedErr, err, t)

		if err == nil {
			if !cmpFn(got, want) {
				t.Errorf("category %s case %d: got %v, want %v", category, i, got, want)
			}
		}
	}
}

func assertError(category string, i int, wantedError, gottenError error, t *testing.T) {
	t.Helper()

	if wantedError != nil && gottenError != wantedError {
		t.Errorf("category %s case %d: got %v, wanted %v", category, i, gottenError, wantedError)
	}

	if wantedError == nil && gottenError != nil {
		t.Errorf("category %s case %d: got %v wanted nil", category, i, gottenError)
	}
}
