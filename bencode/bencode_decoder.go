package bencode

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
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

func (decoder *Decoder) Decode(v any) (ret error) {
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("Recovered: %v", r)
			ret = errors.New(errMsg)
		}
	}()

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

		return byteString, nil
	}
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
	typeOfStruct := reflect.TypeOf(v)

	if reflect.TypeOf(v).Kind() == reflect.Pointer {
		typeOfStruct = reflect.TypeOf(v).Elem()
	}

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
		} else {
			continue
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

		err := assignBasicToComplex(sourceValue, target)
		if err != nil {
			return err
		}
	}

	return nil
}

func assignSlice(source any, v any) error {
	slice := reflect.ValueOf(source)

	targetType := reflect.TypeOf(v).Elem()
	targetSlice := reflect.MakeSlice(targetType, slice.Len(), slice.Cap())

	for i := 0; i < slice.Len(); i++ {
		element := targetSlice.Index(i).Addr().Interface()

		if slice.Index(i).Type().AssignableTo(targetType) {
			errMsg := fmt.Sprintf("Target type %v cannot be assigned from source %d %v", targetType, i, slice.Index(i).Type())
			return errors.New(errMsg)
		}

		assignBasicToComplex(slice.Index(i).Interface(), element)
	}

	reflect.ValueOf(v).Elem().Set(targetSlice)

	return nil
}

func assignDictionary(source any, v any) error {
	sourceDict, ok := source.(map[string]any)

	if !ok {
		return errors.New("Source not map[string]any")
	}

	newDict := make(map[string]any)

	for key := range sourceDict {
		valueType := reflect.TypeOf(sourceDict[key])
		value := reflect.New(valueType)

		assignBasicToComplex(sourceDict[key], value.Interface())
		newDict[key] = value.Elem().Interface()
	}

	dictPtr, ok := v.(*map[string]any)
	if !ok {
		return errors.New("v is not map[string]any")
	}

	*dictPtr = newDict

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
	case reflect.Map:
		err = assignDictionary(source, v)
		break
	case reflect.Pointer:
		// Handle pointer to pointer
		pointingType := reflect.TypeOf(v).Elem()
		target := reflect.New(pointingType.Elem())

		err = assignBasicToComplex(source, target.Interface())

		if err != nil {
			return err
		}

		reflect.ValueOf(v).Elem().Set(target)
	default:
		sourceValue := reflect.ValueOf(source)
		targetValue := reflect.ValueOf(v)

		targetValue.Elem().Set(sourceValue)
		// reflect.ValueOf(v).Elem().Set(source)
	}

	return err
}
