package bencode

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

type testCase struct {
	name      string
	input     any
	want      any
	wantedErr error
	extras    any
}

type testCaseAsserterFn func(testCase *testCase, t *testing.T) (any, error)

func decoderAssert[T any](testCase *testCase, t *testing.T) (any, error) {
	t.Helper()

	input := testCase.input.(string)
	reader := strings.NewReader(input)
	decoder := NewDecoder(reader)

	var got T
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

func encoderAssert[T any](testCase *testCase, t *testing.T) (any, error) {
	t.Helper()

	var b bytes.Buffer
	input := testCase.input.(T)
	writer := io.Writer(&b)

	encoder := NewEncoder(writer)
	err := encoder.Encode(input)

	if err == nil {
		got := string(b.Bytes())
		want := testCase.want.(string)

		if got != want {
			t.Errorf("name %s: got %v want %v", testCase.name, got, want)
		}
	}

	return nil, err
}

func makePointerTo[T any](value T) *T {
	ptr := new(T)
	*ptr = value

	return ptr
}

func TestDecode(t *testing.T) {
	testGeneralDecode(t)
	testNumbersDecode(t)
	testStringsDecode(t)
	testListsDecode(t)
	testDictsDecode(t)
}

func testGeneralDecode(t *testing.T) {
	generalTestCases := []testCase{
		{"No valid delimiter", "psa", nil, io.EOF, nil},
	}

	runTestCases(generalTestCases, decoderAssert[any], t)
}

func testNumbersDecode(t *testing.T) {
	numberTestCases := []testCase{
		{"1st number", "i45e", 45, nil, nil},
		{"2nd number", "i567e", 567, nil, nil},
		{"3rd number", "i-45e", -45, nil, nil},
		{"Missing number case", "ie", nil, ErrExpectedNumber, nil},
		{"Missing number terminator", "i456", nil, io.EOF, nil},
	}

	runTestCases(numberTestCases, decoderAssert[int], t)
}

func testStringsDecode(t *testing.T) {
	stringTestCases := []testCase{
		{"1st string", "3:ben", "ben", nil, nil},
		// Less characters than specified
		{"2nd string", "4:ben", nil, ErrEndOfReader, nil},
		// More characters than specified
		{"3rd string", "2:ben", "be", nil, nil},
		// 0 characters specified
		{"4th string", "0:", "", nil, nil},
		{"5th string", "1:a", "a", nil, nil},
	}

	runTestCases(stringTestCases, decoderAssert[string], t)
}

func testListsDecode(t *testing.T) {
	listTestCases := []testCase{
		{"1st list", "l3:ben2:goe", []any{"ben", "go"}, nil, nil},
		{"2nd list", "l3:beni56ee", []any{"ben", 56}, nil, nil},
		{"Empty list", "le", []any{}, nil, nil},
		{"Nested lists", "llelleei0ee", []any{[]any{}, []any{[]any{}}, 0}, nil, nil},
		{"Missing end of list", "l", nil, io.EOF, nil},
	}

	runTestCases(listTestCases, decoderAssert[any], t)

	intListTypeCheckingCases := []testCase{
		{"int list", "li1ei2ei3ee", []int{1, 2, 3}, nil, nil},
	}

	runTestCases(intListTypeCheckingCases, decoderAssert[[]int], t)

	stringListTypeCheckingCases := []testCase{
		{"string list", "l3:ben3:ken4:gwene", []string{"ben", "ken", "gwen"}, nil, nil},
	}

	runTestCases(stringListTypeCheckingCases, decoderAssert[[]string], t)
}

func testDictsDecode(t *testing.T) {
	dictTestCases := []testCase{
		{"1st dict", "d3:ben2:goe", map[string]any{"ben": "go"}, nil, nil},
		{"2nd dict", "d3:beni56ee", map[string]any{"ben": 56}, nil, nil},
		{"Empty dict", "de", map[string]any{}, nil, nil},
		{"Nested dicts", "d3:bend3:ben3:kenee", map[string]any{"ben": map[string]any{"ben": "ken"}}, nil, nil},
		{"Missing dict terminator", "d", nil, io.EOF, nil},
	}

	runTestCases(dictTestCases, decoderAssert[map[string]any], t)

	type InnerStructExample struct {
		X string `bencode:"x"`
		Y *int   `bencode:"y"`
	}
	type MatchingStructExample struct {
		Ben      string              `bencode:"ben"`
		Number   int                 `bencode:"number"`
		List     []int               `bencode:"list"`
		Nullable *int                `bencode:"nullable"`
		Inner    *InnerStructExample `bencode:"inner"`
	}

	structAssignmentCases := []testCase{
		{"1st struct assignment", "d3:ben3:ken6:numberi3e4:listli1ei2ei3eee", MatchingStructExample{"ken", 3, []int{1, 2, 3}, nil, nil}, nil, nil},
		{"2nd struct assignment", "d3:ben3:ken6:numberi3e4:listli1ei2ei3ee8:nullablei5ee", MatchingStructExample{"ken", 3, []int{1, 2, 3}, makePointerTo(5), nil}, nil, nil},
		{"3nd struct assignment", "d3:ben3:ken6:numberi3e4:listli1ei2ei3ee5:innerd1:x1:xee", MatchingStructExample{"ken", 3, []int{1, 2, 3}, nil, &InnerStructExample{"x", nil}}, nil, nil},
	}

	runTestCases(structAssignmentCases, decoderAssert[MatchingStructExample], t)
}

func TestEncode(t *testing.T) {
	testIntEncode(t)
	testStringEncode(t)
	testListEncode(t)
	testDictEncode(t)
}

func testIntEncode(t *testing.T) {
	testCases := []testCase{
		{"1st number encode", 123, "i123e", nil, nil},
		{"2nd number encode", 0, "i0e", nil, nil},
		{"3rd number encode", -123, "i-123e", nil, nil},
	}
	runTestCases(testCases, encoderAssert[int], t)
}

func testStringEncode(t *testing.T) {
	testCases := []testCase{
		{"1st string encode", "ben", "3:ben", nil, nil},
		{"2nd string encode", "", "0:", nil, nil},
	}
	runTestCases(testCases, encoderAssert[string], t)
}

func testListEncode(t *testing.T) {
	testCases := []testCase{
		{"1st list encode", []any{0, "ben"}, "li0e3:bene", nil, nil},
	}
	runTestCases(testCases, encoderAssert[[]any], t)
}

func testDictEncode(t *testing.T) {
	testCases := []testCase{
		{"1st dict encode", map[string]any{"ben": 123, "ken": []any{}}, "d3:beni123e3:kenlee", nil, nil},
	}
	runTestCases(testCases, encoderAssert[map[string]any], t)
}

func runTestCases(testCases []testCase, testCaseAsserter testCaseAsserterFn, t *testing.T) {
	t.Helper()

	for i := range testCases {
		testCase := testCases[i]

		_, err := testCaseAsserter(&testCase, t)
		assertError(err, &testCase, t)
	}
}

func assertError(err error, testCase *testCase, t *testing.T) {
	t.Helper()

	name := testCase.name
	wantedError := testCase.wantedErr

	if wantedError != nil && err != wantedError {
		t.Errorf("name %s: got %v, wanted %v", name, err, wantedError)
	}

	if wantedError == nil && err != nil {
		t.Errorf("name %s: got %v wanted nil", name, err)
	}
}
