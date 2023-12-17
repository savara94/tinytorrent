package bencode

import (
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strconv"
)

type Encoder struct {
	writer io.Writer
}

func NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{writer}
}

func Marshal(v any) ([]byte, error) {
	switch kind := reflect.TypeOf(v).Kind(); kind {
	case reflect.Int:
		num := v.(int)
		return encodeNumber(num), nil
	case reflect.String:
		byteString := v.(string)
		return encodeString(byteString), nil
	case reflect.Slice:
		return encodeList(v)
	case reflect.Map:
		dict := v.(map[string]any)
		return encodeDict(dict)
	case reflect.Struct:
		return encodeStruct(v)
	case reflect.Pointer:
		pointerValue := reflect.ValueOf(v)
		isNil := pointerValue.IsNil()
		if isNil {
			return []byte{}, nil
		}

		return Marshal(pointerValue.Elem().Interface())
	default:
		log.Printf("Not supported kind %#v %#v", kind, v)
		return nil, ErrNotSupportedType
	}
}

func (encoder *Encoder) Encode(v any) error {
	bytes, err := Marshal(v)
	if err != nil {
		return err
	}

	n, err := encoder.writer.Write(bytes)

	if err != nil {
		return err
	}

	if n < len(bytes) {
		errMsg := fmt.Sprintf("Error writing to %v", encoder.writer)
		return errors.New(errMsg)
	}

	return nil
}

func encodeNumber(number int) []byte {
	encoded := "i" + strconv.Itoa(number) + "e"

	return []byte(encoded)
}

func encodeString(str string) []byte {
	length := len(str)
	encoded := strconv.Itoa(length) + ":" + str

	return []byte(encoded)
}

func encodeList(list any) ([]byte, error) {
	encoded := ""

	listValue := reflect.ValueOf(list)

	for i := 0; i < listValue.Len(); i++ {
		elementValue := listValue.Index(i).Interface()
		encodedElement, err := Marshal(elementValue)

		if err != nil {
			return nil, err
		}

		encoded += string(encodedElement)
	}

	encoded = "l" + encoded + "e"

	return []byte(encoded), nil
}

func encodeDict(dict map[string]any) ([]byte, error) {
	encoded := ""

	// Keys should be encoded in sorted order.
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for i := range keys {
		encodedKey, err := Marshal(keys[i])
		if err != nil {
			return nil, err
		}

		value := dict[keys[i]]
		encodedValue, err := Marshal(value)
		if err != nil {
			return nil, err
		}

		encoded += string(encodedKey) + string(encodedValue)
	}

	encoded = "d" + encoded + "e"

	return []byte(encoded), nil
}

func encodeStruct(object any) ([]byte, error) {
	encoded := ""

	fieldMap := getStructFields(object)

	keys := make([]string, 0, len(fieldMap))
	for key := range fieldMap {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for i := range keys {
		encodedName := keys[i]
		field := fieldMap[encodedName]

		fieldValue := reflect.ValueOf(object).FieldByName(field.Name)

		if fieldValue.Kind() == reflect.Pointer && fieldValue.IsNil() {
			continue
		}

		encodedValue, err := Marshal(fieldValue.Interface())
		if err != nil {
			return nil, err
		}

		encodedKey := encodeString(encodedName)

		encoded += string(encodedKey) + string(encodedValue)

	}

	encoded = "d" + encoded + "e"

	return []byte(encoded), nil
}
