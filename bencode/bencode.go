package bencode

import (
	"errors"
	"io"
	"strconv"
)

var ErrEndOfReader = errors.New("end of reader reached")
var ErrDelimiterExpected = errors.New("expected delimiter")
var ErrEndExpected = errors.New("expected eof")
var ErrZeroByteString = errors.New("bytestring has length 0")
var ErrNoColonFound = errors.New("no colon found")
var ErrExpectedNumber = errors.New("expected number")
var ErrExpectedByteString = errors.New("expected key string")

func decodeDict(reader io.Reader) (map[string]interface{}, error) {
	bytes := make([]byte, 1)
	dict := make(map[string]interface{})

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

func decodeList(reader io.Reader) ([]interface{}, error) {
	bytes := make([]byte, 1)
	list := make([]interface{}, 0)

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

func decodeDelimiter(delimiter byte, reader io.Reader) (interface{}, error) {
	switch delimiter {
	case 'd':
		return decodeDict(reader)
	case 'l':
		return decodeList(reader)
	case 'i':
		return decodeNumber(reader)
	default:
		return decodeByteString(delimiter, reader)
	}
}

func Decode(reader io.Reader) (interface{}, error) {
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
