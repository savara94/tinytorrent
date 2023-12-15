package bencode

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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
	switch casted := v.(type) {
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
		return nil, nil
	default:
		log.Printf("Not supported type %T", v)
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

func encodeNumber(number int) ([]byte, error) {
	encoded := "i" + strconv.Itoa(number) + "e"

	return []byte(encoded), nil
}

func encodeString(str string) ([]byte, error) {
	length := len(str)
	encoded := strconv.Itoa(length) + ":" + str

	return []byte(encoded), nil
}

func encodeList(list []any) ([]byte, error) {
	encoded := ""

	for i := range list {
		encodedElement, err := Marshal(list[i])

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
