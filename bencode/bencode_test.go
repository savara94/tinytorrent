package bencode

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

type testCase struct {
	reader    io.Reader
	want      any
	wantedErr error
}

type testCmpFn func(got, want any) bool

func simpleCmp(got, want any) bool {
	return got == want
}

func sliceCmp(got, want any) bool {
	return reflect.DeepEqual(got, want)
}

func TestDecode(t *testing.T) {
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
		{strings.NewReader("2:ben"), "be", nil},
		// 0 characters specified
		{strings.NewReader("0:"), "", nil},
		// 1 character specified
		{strings.NewReader("1:a"), "a", nil},
	}

	listParsingTestCases := []testCase{
		// Good weather cases
		{strings.NewReader("l3:ben2:goe"), []any{"ben", "go"}, nil},
		{strings.NewReader("l3:beni56ee"), []any{"ben", 56}, nil},
		// Empty list
		{strings.NewReader("le"), []any{}, nil},
		// Missing end of list
		{strings.NewReader("l"), nil, io.EOF},
	}

	dictParsingTestCases := []testCase{
		// Good weather cases
		{strings.NewReader("d3:ben2:goe"), map[string]any{"ben": "go"}, nil},
		{strings.NewReader("d3:beni56ee"), map[string]any{"ben": 56}, nil},
		// Empty dict
		{strings.NewReader("de"), map[string]any{}, nil},
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
		{strings.NewReader("llelleei0ee"), []any{[]any{}, []any{[]any{}}, 0}, nil},
		// Nested dicts
		{strings.NewReader("d3:bend3:ben3:kenee"), map[string]any{"ben": map[string]any{"ben": "ken"}}, nil},
	}

	runTestCases("complex", complexTestCases, t, sliceCmp)
}

// func TestUnmarshal(t *testing.T) {
// 	intTestCases := []struct {
// 		input   any
// 		want    int
// 		wantErr bool
// 	}{
// 		{123, 123, false},
// 		{-123, -123, false},
// 		{"pwd", 0, true},
// 		{make([]any, 1), 0, true},
// 		{make(map[string]any), 0, true},
// 	}

// 	for i := range intTestCases {
// 		input := intTestCases[i].input
// 		wantErr := intTestCases[i].wantErr

// 		var a int

// 		err := Unmarshal(input, &a)

// 		if err == nil && !wantErr {
// 			if input != a {
// 				t.Errorf("wanted %v, got %v", input, a)
// 			}
// 		}

// 		if err != nil && !wantErr {
// 			t.Errorf("did not expect error %v", err)
// 		}
// 	}

// 	stringTestCases := []struct {
// 		input   any
// 		wantErr bool
// 	}{
// 		{"123", false},
// 		{-123, true},
// 		{"pwd", false},
// 		{make([]any, 1), true},
// 		{make(map[string]any), true},
// 	}

// 	for i := range stringTestCases {
// 		input := stringTestCases[i].input
// 		want := stringTestCases[i].input
// 		wantErr := stringTestCases[i].wantErr

// 		var a string

// 		err := Unmarshal(input, &a)

// 		if err == nil && !wantErr {
// 			if want != a {
// 				t.Errorf("wanted %v, got %v", want, a)
// 			}
// 		}

// 		if err != nil && !wantErr {
// 			t.Errorf("did not expect error %v", err)
// 		}
// 	}

// 	listInt := []int{1, 2, 3}
// 	var targetListInt []int

// 	err := Unmarshal(toSliceOfAny(listInt), &targetListInt)
// 	if err != nil {
// 		t.Errorf("Not expected error %v", err)
// 	} else {
// 		if !reflect.DeepEqual(listInt, targetListInt) {
// 			t.Errorf("Not equal %v and %v", listInt, targetListInt)
// 		}
// 	}

// 	listString := []string{"a", "b", "c"}
// 	var targetListString []string

// 	err = Unmarshal(toSliceOfAny(listString), &targetListString)
// 	if err != nil {
// 		t.Errorf("Not expected error %v", err)
// 	} else {
// 		if !reflect.DeepEqual(listString, targetListString) {
// 			t.Errorf("Not equal %v and %v", listString, targetListString)
// 		}
// 	}

// 	dict := map[string]any{
// 		"ben":    "ken",
// 		"number": 3,
// 		"list":   []any{1, 2, 3},
// 		"nested": map[string]any{"key": "value"},
// 	}

// 	goodStruct := struct {
// 		Ben      string `bencode:"ben"`
// 		Number   int    `bencode:"number"`
// 		List     []int  `bencode:"list"`
// 		Nullable *int   `bencode:"nullable"`
// 	}{}

// 	err = Unmarshal(dict, &goodStruct)
// 	if err != nil {
// 		t.Errorf("Not expected error %v", err)
// 	}

// 	if goodStruct.Ben != dict["ben"] || goodStruct.Number != dict["number"] || !reflect.DeepEqual(goodStruct.List, []int{1, 2, 3}) {
// 		t.Errorf("Not equal %v", goodStruct)
// 	}

// 	dict["nullable"] = 5
// 	err = Unmarshal(dict, &goodStruct)
// 	if err != nil {
// 		t.Errorf("Not expected error %v", err)
// 	}

// 	if *goodStruct.Nullable != dict["nullable"] {
// 		t.Errorf("Not equal %v", goodStruct)
// 	}
// }

// func toSliceOfAny[T any](s []T) []any {
// 	result := make([]any, len(s))
// 	for i, v := range s {
// 		result[i] = v
// 	}
// 	return result
// }

// func TestEncode(t *testing.T) {
// 	testCases := []struct {
// 		input       any
// 		want        string
// 		wantedError error
// 	}{
// 		{123, "i123e", nil},
// 		{0, "i0e", nil},
// 		{-123, "i-123e", nil},
// 		{"ben", "3:ben", nil},
// 		{"", "0:", nil},
// 		{[]any{0, "ben"}, "li0e3:bene", nil},
// 		{map[string]any{"ben": 123, "ken": []any{}}, "d3:beni123e3:kenlee", nil},
// 	}

// 	for i, testCase := range testCases {
// 		input := testCase.input
// 		want := testCase.want
// 		wantedError := testCase.wantedError

// 		got, gottenErr := Encode(input)

// 		if wantedError != nil && gottenErr == nil {
// 			t.Errorf("wanted %v got nil", wantedError)
// 		}

// 		if wantedError == nil && gottenErr != nil {
// 			t.Errorf("%d not expected error but got %v", i, gottenErr)
// 		}

// 		if wantedError != nil && gottenErr != nil {
// 			if wantedError != gottenErr {
// 				t.Errorf("got %v expected %v", gottenErr, wantedError)
// 			}
// 		}

// 		if got != want {
// 			t.Errorf("wanted %v got %v", want, got)
// 		}
// 	}
// }

func runTestCases(category string, testCases []testCase, t *testing.T, cmpFn testCmpFn) {
	for i := range testCases {
		testCase := testCases[i]

		reader := testCase.reader
		want := testCase.want
		wantedErr := testCase.wantedErr

		decoder := NewDecoder(reader)

		var got any

		err := decoder.Decode(&got)

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
