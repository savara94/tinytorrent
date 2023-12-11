package bencode

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"reflect"
	"sort"
	"strconv"
	"unicode/utf8"
)

var ErrEndOfReader = errors.New("end of reader reached")
var ErrDelimiterExpected = errors.New("expected delimiter")
var ErrEndExpected = errors.New("expected eof")
var ErrZeroByteString = errors.New("bytestring has length 0")
var ErrNoColonFound = errors.New("no colon found")
var ErrExpectedNumber = errors.New("expected number")
var ErrExpectedByteString = errors.New("expected key string")
var ErrNotSupportedType = errors.New("not supported type")

func decodeDict(reader io.Reader) (map[string]any, error) {
	bytes := make([]byte, 1)
	dict := make(map[string]any)

	for {
		n, err := reader.Read(bytes)

		if err != nil {
			return nil, err
		}

		if n == 0 {
			return nil, ErrEndOfReader
		}

		delimiter := bytes[0]

		if delimiter == 'e' {
			break
		}

		key, err := decodeByteString(delimiter, reader)

		if err != nil {
			return nil, ErrExpectedByteString
		}

		n, err = reader.Read(bytes)

		if err != nil {
			return nil, err
		}

		if n == 0 {
			return nil, ErrEndOfReader
		}

		delimiter = bytes[0]
		element, err := decodeDelimiter(delimiter, reader)

		if err != nil {
			return nil, err
		}

		dict[key] = element
	}

	return dict, nil
}

func decodeNumber(reader io.Reader) (int, error) {
	var bytesArray []byte
	bytes := make([]byte, 1)

	for {
		n, err := reader.Read(bytes)

		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, ErrEndOfReader
		}

		if bytes[0] == 'e' {
			break
		} else {
			bytesArray = append(bytesArray, bytes...)
		}
	}

	intString := string(bytesArray)

	number, err := strconv.Atoi(intString)

	if err != nil {
		return 0, ErrExpectedNumber
	}

	return number, nil
}

func decodeList(reader io.Reader) ([]any, error) {
	bytes := make([]byte, 1)
	list := make([]any, 0)

	for {
		n, err := reader.Read(bytes)

		if err != nil {
			return nil, err
		}

		if n == 0 {
			return nil, ErrEndOfReader
		}

		delimiter := bytes[0]

		if delimiter == 'e' {
			break
		}

		element, err := decodeDelimiter(delimiter, reader)

		if err != nil {
			return nil, err
		}

		list = append(list, element)
	}

	return list, nil
}

func decodeByteString(firstDigit byte, reader io.Reader) (string, error) {
	bytes := make([]byte, 1)
	lengthSlice := []byte{firstDigit}

	for {
		n, err := reader.Read(bytes)

		if err != nil {
			return "", err
		}

		if n == 0 {
			return "", ErrEndOfReader
		}

		if bytes[0] == ':' {
			break
		} else {
			lengthSlice = append(lengthSlice, bytes...)
		}
	}

	lengthString := string(lengthSlice)
	length, err := strconv.Atoi(lengthString)

	if err != nil {
		return "", err
	}

	if length == 0 {
		return "", nil
	}

	byteString := make([]byte, length)
	n, err := reader.Read(byteString)

	if err != nil {
		return "", err
	}

	if n != length {
		return "", ErrEndOfReader
	}

	return string(byteString), nil
}

func decodeDelimiter(delimiter byte, reader io.Reader) (any, error) {
	switch delimiter {
	case 'd':
		return decodeDict(reader)
	case 'l':
		return decodeList(reader)
	case 'i':
		return decodeNumber(reader)
	default:
		byteString, err := decodeByteString(delimiter, reader)

		if err != nil {
			return nil, err
		}

		if utf8.ValidString(byteString) {
			return byteString, nil
		}

		return []byte(byteString), nil
	}
}

func encodeNumber(number int) (string, error) {
	encoded := "i" + strconv.Itoa(number) + "e"
	return encoded, nil
}

func encodeString(str string) (string, error) {
	length := len(str)
	encoded := strconv.Itoa(length) + ":" + str
	return encoded, nil
}

func encodeList(list []any) (string, error) {
	encoded := ""

	for i := range list {
		encodedElement, err := Encode(list[i])

		if err != nil {
			return "", err
		}

		encoded += encodedElement
	}

	encoded = "l" + encoded + "e"
	return encoded, nil
}

func encodeDict(dict map[string]any) (string, error) {
	encoded := ""

	// Keys should be encoded in sorted order.
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for i := range keys {
		encodedKey, err := Encode(keys[i])
		if err != nil {
			return "", err
		}

		value := dict[keys[i]]
		encodedValue, err := Encode(value)
		if err != nil {
			return "", err
		}

		encoded += encodedKey + encodedValue
	}

	encoded = "d" + encoded + "e"
	return encoded, nil
}

func Decode(reader io.Reader) (any, error) {
	bytes := make([]byte, 1)

	n, err := reader.Read(bytes)

	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrEndOfReader
	}

	content, err := decodeDelimiter(bytes[0], reader)

	if err != nil {
		return nil, err
	}

	n, err = reader.Read(bytes)

	if err != nil {
		if err != io.EOF {
			return nil, ErrEndExpected
		}
	}

	if n > 0 {
		return nil, ErrEndExpected
	}

	return content, nil
}

func getStructFields(v any) map[string]reflect.StructField {
	typeOfStruct := reflect.TypeOf(v)
	fields := make([]string, typeOfStruct.NumField())

	fieldMap := make(map[string]reflect.StructField)

	for i := range fields {
		field := typeOfStruct.Field(i)

		if !field.IsExported() {
			continue
		}

		if field.Anonymous {
			continue
		}

		encodeName := field.Name

		if encName, exists := field.Tag.Lookup("bencode"); exists {
			encodeName = encName
		}

		fieldMap[encodeName] = field
	}

	return fieldMap
}

func assignStruct(source any, v any) error {
	anyMap, ok := source.(map[string]any)

	if !ok {
		return errors.New("Source is not map")
	}

	fieldMap := getStructFields(v)

	for encodedName := range fieldMap {
		field := fieldMap[encodedName]
		isNullable := false

		fieldValue := reflect.ValueOf(v).FieldByName(field.Name)

		if fieldValue.Kind() == reflect.Pointer {
			isNullable = true
		}

		sourceValue, exists := anyMap[encodedName]
		if !exists && !isNullable {
			return errors.New(fmt.Sprintf("%v key does not exist and %v is not nullable.", encodedName, field.Name))
		}

		structV := fieldValue.Interface()
		Unmarshal(sourceValue, &structV)
	}

	return nil
}

func assignInt(source any, v any) error {
	number, ok := source.(int)

	if !ok {
		return errors.New(fmt.Sprintf("Source %v %T not int", source, source))
	}

	target := reflect.ValueOf(v).Interface().(*int)
	*target = number

	return nil
}

func assignString(source any, v any) error {
	str, ok := source.(string)

	if !ok {
		return errors.New("Source not string")
	}

	target := reflect.ValueOf(v).Interface().(*string)
	*target = str

	return nil
}

func assignSlice(source any, v any) error {
	slice, ok := source.([]any)

	if !ok {
		return errors.New("Source not slice of any")
	}

	targetType := reflect.TypeOf(v).Elem()
	targetSlice := reflect.MakeSlice(targetType, len(slice), len(slice))

	for i := range slice {
		element := targetSlice.Index(i).Addr().Interface()
		Unmarshal(slice[i], element)
	}

	reflect.ValueOf(v).Elem().Set(targetSlice)

	return nil
}

func Unmarshal(source any, v any) error {
	if reflect.TypeOf(v).Kind() != reflect.Pointer {
		return errors.New("v is not a pointer")
	}

	var err error

	switch kind := reflect.ValueOf(v).Elem().Kind(); kind {
	case reflect.Struct:
		err = assignStruct(source, v)
		break
	case reflect.Slice:
		err = assignSlice(source, v)
		break
	case reflect.Int:
		err = assignInt(source, v)
		break
	case reflect.String:
		err = assignString(source, v)
		break
	default:
		return errors.New(fmt.Sprintf("%v not supported", kind))
	}

	return err
}

func Encode(object any) (string, error) {
	switch casted := object.(type) {
	case int:
		return encodeNumber(casted)
	case float64:
		// Check if no data loss occurred
		return encodeNumber(int(math.Round(casted)))
	case string:
		return encodeString(casted)
	case []any:
		return encodeList(casted)
	case map[string]any:
		return encodeDict(casted)
	case nil:
		return "", nil
	default:
		log.Printf("Not supported type %T", object)
		return "", ErrNotSupportedType
	}
}