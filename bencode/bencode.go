package bencode

import (
	"bytes"
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

type Decoder struct {
	reader    io.Reader
	bytesRead int
}

func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{reader, 0}
}

func (decoder *Decoder) readBytes(numBytes int) ([]byte, error) {
	bytes := make([]byte, numBytes)

	n, err := decoder.reader.Read(bytes)

	decoder.bytesRead += n

	if err != nil {
		return nil, err
	}

	if n < numBytes {
		return nil, ErrEndOfReader
	}

	return bytes, nil
}

func isPointerToEmptyInterface(v any) bool {
	var sliceOfEmptyInterface []interface{}

	pointerType := reflect.TypeOf(v).Elem()
	return pointerType == reflect.TypeOf(sliceOfEmptyInterface).Elem()
}

func (decoder *Decoder) Decode(v any) error {
	if reflect.TypeOf(v).Kind() != reflect.Pointer {
		errMsg := fmt.Sprintf("%v not passed pointer!", v)
		return errors.New(errMsg)
	}

	basic, err := decoder.decodeToBasicTypes()

	if err != nil {
		return err
	}

	if isPointerToEmptyInterface(v) {
		basicValue := reflect.ValueOf(basic)
		reflect.ValueOf(v).Elem().Set(basicValue)

		return nil
	}

	return assignBasicToComplex(basic, v)
}

func Unmarshal(data []byte, v any) error {
	reader := bytes.NewReader(data)
	decoder := NewDecoder(reader)

	return decoder.Decode(v)
}

func (decoder *Decoder) decodeDict() (map[string]any, error) {
	dict := make(map[string]any)

	for {
		bytes, err := decoder.readBytes(1)

		if err != nil {
			return nil, err
		}

		delimiter := bytes[0]

		if delimiter == 'e' {
			break
		}

		key, err := decoder.decodeByteString(delimiter)

		if err != nil {
			return nil, ErrExpectedByteString
		}

		bytes, err = decoder.readBytes(1)

		if err != nil {
			return nil, err
		}

		delimiter = bytes[0]
		element, err := decoder.decodeDelimiter(delimiter)

		if err != nil {
			return nil, err
		}

		dict[key] = element
	}

	return dict, nil
}

func (decoder *Decoder) decodeNumber() (int, error) {
	var bytesArray []byte

	for {
		bytes, err := decoder.readBytes(1)

		if err != nil {
			return 0, err
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

func (decoder *Decoder) decodeList() ([]any, error) {
	list := make([]any, 0)

	for {
		bytes, err := decoder.readBytes(1)

		if err != nil {
			return nil, err
		}

		delimiter := bytes[0]

		if delimiter == 'e' {
			break
		}

		element, err := decoder.decodeDelimiter(delimiter)

		if err != nil {
			return nil, err
		}

		list = append(list, element)
	}

	return list, nil
}

func (decoder *Decoder) decodeByteString(firstChar byte) (string, error) {
	lengthSlice := []byte{firstChar}

	for {
		bytes, err := decoder.readBytes(1)

		if err != nil {
			return "", err
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

	byteString, err := decoder.readBytes(length)

	if err != nil {
		return "", err
	}

	return string(byteString), nil
}

func (decoder *Decoder) decodeDelimiter(delimiter byte) (any, error) {
	switch delimiter {
	case 'd':
		return decoder.decodeDict()
	case 'l':
		return decoder.decodeList()
	case 'i':
		return decoder.decodeNumber()
	default:
		byteString, err := decoder.decodeByteString(delimiter)

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

func (decoder *Decoder) decodeToBasicTypes() (any, error) {
	bytes, err := decoder.readBytes(1)

	if err != nil {
		return nil, err
	}

	delimiter := bytes[0]
	content, err := decoder.decodeDelimiter(delimiter)

	if err != nil {
		return nil, err
	}

	return content, nil
}

func getStructFields(v any) map[string]reflect.StructField {
	typeOfStruct := reflect.TypeOf(v).Elem()
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

		fieldValue := reflect.ValueOf(v).Elem().FieldByName(field.Name)
		target := fieldValue.Addr().Interface()

		if fieldValue.Kind() == reflect.Pointer {
			isNullable = true
		}

		sourceValue, exists := anyMap[encodedName]
		if !isNullable && !exists {
			return errors.New(fmt.Sprintf("%v key does not exist and %v is not nullable.", encodedName, field.Name))
		}

		if isNullable && !exists {
			continue
		}

		assignBasicToComplex(sourceValue, target)
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
		assignBasicToComplex(slice[i], element)
	}

	reflect.ValueOf(v).Elem().Set(targetSlice)

	return nil
}

func assignBasicToComplex(source any, v any) error {
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
	case reflect.Pointer:
		// Handle pointer to pointer
		pointingType := reflect.TypeOf(v).Elem()
		target := reflect.New(pointingType.Elem())
		// fmt.Printf("%v %v %v", pointingType, pointingType.Elem(), target)
		err = assignBasicToComplex(source, target.Interface())

		if err != nil {
			return err
		}

		reflect.ValueOf(v).Elem().Set(target)
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
