package ini

/*
 Custom limited INI parser.
 /TODO Add docs
*/

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
)

const reflectTag = "ini"

var (
	ErrInvalidContent  = errors.New("invalid content")
	ErrUnsupportedType = errors.New("content is not supported")
	ErrMissingTag      = errors.New("missing tag for field")
	ErrUnexpected      = errors.New("unexpected behavior")
)

// Unmarshall reads a bytes slice and assigns them to the passed interface
// currently supports only interfaces in the form map[string]struct
func Unmarshal(data []byte, v interface{}) error {
	if reflect.ValueOf(v).Kind() != reflect.Map {
		return fmt.Errorf("%w: value to unmarshall is not map", ErrUnsupportedType)
	}

	nameRe := regexp.MustCompile("\\[[[:graph:]]+\\]")

	vt := reflect.TypeOf(v).Elem()
	kt := reflect.TypeOf(v).Key()

	sectionIndexes := nameRe.FindAllIndex(data, -1)

	split := createSplit(sectionIndexes, len(data))
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Split(split)

	for _, p := range sectionIndexes {
		sc.Scan()
		name := reflect.New(kt)
		if err := parseName(data[p[0]:p[1]], name); err != nil {
			return err
		}

		c := reflect.New(vt)
		if err := parseAttributes(sc.Bytes(), c.Elem()); err != nil {
			return err
		}

		reflect.ValueOf(v).SetMapIndex(name.Elem(), c.Elem())
	}

	return nil
}

// Marshal takes a value and writes it into a byte array.
// currently only supports map[string]struct
func Marshal(v interface{}) ([]byte, error) {
	if reflect.ValueOf(v).Kind() != reflect.Map {
		return nil, fmt.Errorf("%w: value to marshall is not map", ErrUnsupportedType)
	}

	buffer := bytes.NewBuffer([]byte{})

	vv := reflect.ValueOf(v)
	keys := vv.MapKeys()
	for _, k := range keys {
		val := vv.MapIndex(k)
		if err := write(buffer, k, val); err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}

// parseName transforms bytes into a string name, it expects name to be in
// the format [name]
func parseName(data []byte, v reflect.Value) error {
	if v.Elem().Kind() != reflect.String {
		return fmt.Errorf("%w: expected name to be string, but is %s", ErrUnsupportedType, v.Elem().Kind().String())
	}

	l := len(data)
	if l < 3 || string(data[0]) != "[" || string(data[l-1]) != "]" {
		return fmt.Errorf("%w: name should be enclosed in []", ErrInvalidContent)
	}
	s := string(data[1 : l-1])

	v.Elem().SetString(s)
	return nil
}

// parseAttributes takes a bytes array, reads the fields in them and assigns
// those to the passed value.
// currently in only supports assigning to struct
func parseAttributes(data []byte, v reflect.Value) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("%w: expected value to be struct, but is %s", ErrUnsupportedType, v.Kind().String())
	}

	fields, err := readFields(data)
	if err != nil {
		return err
	}

	lf := v.Type().NumField()
	t := v.Type()

	for i := 0; i < lf; i++ {
		f := t.Field(i)
		tag := f.Tag.Get(reflectTag)
		if tag == "" {
			return fmt.Errorf("%w: %s", ErrMissingTag, f.Name)
		}

		if fv, ok := fields[tag]; ok {
			v.Field(i).SetString(fv)
		}
	}

	return nil
}

// readFields takes data from byte array, splits it into lines and tokenizes
// these using " = ". Returns a map where the key is the content on the left
// side of = and the value is the content on the right side.
func readFields(data []byte) (map[string]string, error) {
	flds := map[string]string{}

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Split(bufio.ScanLines)

	for sc.Scan() {
		line := sc.Text()

		if len(line) == 0 {
			continue
		}

		parts := strings.Split(sc.Text(), "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: each field should be in the format key = value, invalid %s ", ErrInvalidContent, line)
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		flds[k] = v
	}

	return flds, nil

}

// createSplit creates a split function for the scanner that tokenizes
// the data using [name] as separator. Takes in a slice with the indexes of
// the start and end of each separator and the totalBytes that will be read.
// Uses the separator indexes to read the data between them.
func createSplit(idx [][]int, totalBytes int) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	i := 0
	readBytes := 0

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Return nothing if at end of file and no data passed
		if atEOF && i >= len(idx) {
			return 0, nil, nil
		}

		nameLen := idx[i][1] - idx[i][0]
		var end int
		if i+1 < len(idx) {
			end = idx[i+1][0]
		} else {
			end = totalBytes
		}

		advance = end - readBytes
		if advance > len(data) {
			return 0, nil, nil
		}

		token = data[nameLen+1 : advance]
		i += 1
		readBytes += advance
		return
	}

}

// write writes to provided writer the formatted data.
// k currently only supports string and will be written as: [k]
// v currently only supports struct and will be written as lines of:
// field_tag = value
func write(w io.Writer, k reflect.Value, v reflect.Value) error {
	if k.Kind() != reflect.String {
		return fmt.Errorf("%w: name is not a string", ErrUnsupportedType)
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("%w: value is not a struct", ErrUnsupportedType)
	}

	ks := fmt.Sprintf("[%s]\n", k.String())

	// ignoring returning error since it's always nil
	// https://pkg.go.dev/bytes#Buffer.Write
	w.Write([]byte(ks))

	lf := v.Type().NumField()
	t := v.Type()

	for i := 0; i < lf; i++ {
		f := t.Field(i)
		tag := f.Tag.Get("ini")
		if tag == "" {
			return fmt.Errorf("%w: %s", ErrMissingTag, f.Name)
		}

		vs := fmt.Sprintf("%s = %s\n", tag, v.Field(i).String())
		w.Write([]byte(vs))
	}

	w.Write([]byte("\n"))

	return nil
}
