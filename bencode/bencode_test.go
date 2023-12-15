package bencode

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"
)

type testCase struct {
	name      string
	input     string
	want      any
	wantedErr error
}

type decoderAsserterFn func(decoder *Decoder, testCase *testCase, t *testing.T) (any, error)

func noDecoderTypeCheckingAssert(decoder *Decoder, testCase *testCase, t *testing.T) (any, error) {
	t.Helper()

	var got any
	want := testCase.want
	name := testCase.name

	err := decoder.Decode(&got)

	if err == nil {
		if !reflect.DeepEqual(got, want) {
			t.Errorf("name %s: got %v, want %v", name, got, want)
		}
	}

	return got, err
}

func TestDecode(t *testing.T) {
	testGeneral(t)
	testNumbers(t)
	testStrings(t)
	testLists(t)
	testDicts(t)
}

func testGeneral(t *testing.T) {
	generalTestCases := []testCase{
		{"No valid delimiter", "psa", nil, io.EOF},
	}

	runTestCases(generalTestCases, noDecoderTypeCheckingAssert, t)
}

func testNumbers(t *testing.T) {
	numberTestCases := []testCase{
		{"1st number", "i45e", 45, nil},
		{"2nd number", "i567e", 567, nil},
		{"3rd number", "i-45e", -45, nil},
		{"Missing number case", "ie", nil, ErrExpectedNumber},
		{"Missing number terminator", "i456", nil, io.EOF},
	}

	runTestCases(numberTestCases, noDecoderTypeCheckingAssert, t)

	intTypeCheckingAssert := func(decoder *Decoder, testCase *testCase, t *testing.T) (any, error) {
		var got int

		err := decoder.Decode(&got)

		if err == nil {
			if got != testCase.want {
				t.Errorf("%s got %d want %d", testCase.name, got, testCase.want)
			}
		}

		return got, err
	}

	runTestCases(numberTestCases, intTypeCheckingAssert, t)
}

func testStrings(t *testing.T) {
	stringTestCases := []testCase{
		{"1st string", "3:ben", "ben", nil},
		// Less characters than specified
		{"2nd string", "4:ben", nil, ErrEndOfReader},
		// More characters than specified
		{"3rd string", "2:ben", "be", nil},
		// 0 characters specified
		{"4th string", "0:", "", nil},
		{"5th string", "1:a", "a", nil},
	}

	runTestCases(stringTestCases, noDecoderTypeCheckingAssert, t)

	stringTypeCheck := func(decoder *Decoder, testCase *testCase, t *testing.T) (any, error) {
		var got string

		err := decoder.Decode(&got)

		if err == nil {
			if got != testCase.want {
				t.Errorf("%s got %s want %s", testCase.name, got, testCase.want)
			}
		}

		return got, err
	}

	runTestCases(stringTestCases, stringTypeCheck, t)
}

func testLists(t *testing.T) {
	listTestCases := []testCase{
		{"1st list", "l3:ben2:goe", []any{"ben", "go"}, nil},
		{"2nd list", "l3:beni56ee", []any{"ben", 56}, nil},
		{"Empty list", "le", []any{}, nil},
		{"Nested lists", "llelleei0ee", []any{[]any{}, []any{[]any{}}, 0}, nil},
		{"Missing end of list", "l", nil, io.EOF},
	}

	runTestCases(listTestCases, noDecoderTypeCheckingAssert, t)

	intListTypeCheckingCases := []testCase{
		{"int list", "li1ei2ei3ee", []int{1, 2, 3}, nil},
	}

	intListTypeCheckingAssert := func(decoder *Decoder, testCase *testCase, t *testing.T) (any, error) {
		var got []int

		err := decoder.Decode(&got)

		if err == nil {
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("got %v wanted %v", got, testCase.want)
			}
		}

		return got, err
	}

	runTestCases(intListTypeCheckingCases, intListTypeCheckingAssert, t)

	stringListTypeCheckingCases := []testCase{
		{"string list", "l3:ben3:ken4:gwene", []string{"ben", "ken", "gwen"}, nil},
	}

	stringListTypeCheckingAssert := func(decoder *Decoder, testCase *testCase, t *testing.T) (any, error) {
		var got []string

		err := decoder.Decode(&got)

		if err == nil {
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("got %v wanted %v", got, testCase.want)
			}
		}

		return got, err
	}

	runTestCases(stringListTypeCheckingCases, stringListTypeCheckingAssert, t)
}

func testDicts(t *testing.T) {
	dictTestCases := []testCase{
		{"1st dict", "d3:ben2:goe", map[string]any{"ben": "go"}, nil},
		{"2nd dict", "d3:beni56ee", map[string]any{"ben": 56}, nil},
		{"Empty dict", "de", map[string]any{}, nil},
		{"Nested dicts", "d3:bend3:ben3:kenee", map[string]any{"ben": map[string]any{"ben": "ken"}}, nil},
		{"Missing dict terminator", "d", nil, io.EOF},
	}

	runTestCases(dictTestCases, noDecoderTypeCheckingAssert, t)

	type MatchingStructExample struct {
		Ben      string `bencode:"ben"`
		Number   int    `bencode:"number"`
		List     []int  `bencode:"list"`
		Nullable *int   `bencode:"nullable"`
	}

	structAssignmentCases := []testCase{
		{"1st struct", "d3:ben3:ken6:numberi3e4:listli1ei2ei3ee6:nestedd3:key5:valueee", MatchingStructExample{"ken", 3, []int{1, 2, 3}, nil}, nil},
		{"2nd struct", "d3:ben3:ken6:numberi3e4:listli1ei2ei3ee6:nestedd3:key5:valuee8:nullablei5ee", MatchingStructExample{"ken", 3, []int{1, 2, 3}, new(int)}, nil},
	}

	runTestCases(structAssignmentCases, func(decoder *Decoder, testCase *testCase, t *testing.T) (any, error) {
		var got MatchingStructExample

		err := decoder.Decode(&got)

		if err == nil {
			gotJson, _ := json.Marshal(got)
			wantJson, _ := json.Marshal(testCase.want)

			if !reflect.DeepEqual(gotJson, wantJson) {
				t.Errorf("%s got %v want %v", testCase.name, string(gotJson), string(wantJson))
			}
		}

		return got, err
	}, t)
}

// func TestUnmarshal(t *testing.T) {

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

func runTestCases(testCases []testCase, assertDecoder decoderAsserterFn, t *testing.T) {
	for i := range testCases {
		testCase := testCases[i]

		reader := strings.NewReader(testCase.input)
		wantedErr := testCase.wantedErr
		name := testCase.name

		decoder := NewDecoder(reader)

		_, err := assertDecoder(decoder, &testCase, t)
		assertError(name, wantedErr, err, t)
	}
}

func assertError(name string, wantedError, gottenError error, t *testing.T) {
	t.Helper()

	if wantedError != nil && gottenError != wantedError {
		t.Errorf("name %s: got %v, wanted %v", name, gottenError, wantedError)
	}

	if wantedError == nil && gottenError != nil {
		t.Errorf("name %s: got %v wanted nil", name, gottenError)
	}
}
