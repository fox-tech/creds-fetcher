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
	"log"
	"reflect"
	"regexp"
	"strings"
)

const reflectTag = "ini"

var (
	ErrInvalidRegex   = errors.New("invalid regular expression")
	ErrInvalidContent = errors.New("content does not match expected")
)

func Unmarshal(data []byte, v interface{}) error {
	nameRe, err := regexp.Compile("\\[[[:graph:]]+\\]")
	if err != nil {
		log.Fatalf("error while compiling regex: %v", err)
		return ErrInvalidRegex
	}

	vt := reflect.TypeOf(v).Elem()
	kt := reflect.TypeOf(v).Key()

	sectionIndexes := nameRe.FindAllIndex(data, -1)

	split := createSplit(sectionIndexes, len(data))
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Split(split)

	for _, p := range sectionIndexes {
		sc.Scan()
		name := reflect.New(kt)
		parseName(data[p[0]:p[1]], name)

		c := reflect.New(vt)
		parseAttributes(sc.Bytes(), c.Elem())
		reflect.ValueOf(v).SetMapIndex(name.Elem(), c.Elem())
	}

	return nil
}

func Marshal(v interface{}) ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})

	vv := reflect.ValueOf(v)
	keys := vv.MapKeys()
	for _, k := range keys {
		val := vv.MapIndex(k)
		writeData(buffer, k, val)

	}

	return buffer.Bytes(), nil
}

func parseName(data []byte, v reflect.Value) error {
	// remove starting and trailing [ ]

	// It needs to have ONE identifier between []
	l := len(data)
	if l < 3 {
		return ErrInvalidContent
	}
	s := string(data[1 : l-1])
	v.Elem().SetString(s)
	return nil
}

func parseAttributes(data []byte, v reflect.Value) error {
	fields := readFields(data)

	lf := v.Type().NumField()
	t := v.Type()

	for i := 0; i < lf; i++ {
		f := t.Field(i)
		tag := f.Tag.Get(reflectTag)

		if fv, ok := fields[tag]; ok {
			v.Field(i).SetString(fv)
		}
	}

	return nil
}

func readFields(data []byte) map[string]string {
	flds := map[string]string{}

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Split(bufio.ScanLines)

	for sc.Scan() {
		parts := strings.Split(sc.Text(), "=")
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			flds[k] = v
		}
	}

	return flds

}

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

func writeData(w io.Writer, k reflect.Value, v reflect.Value) error {
	ks := fmt.Sprintf("[%s]\n", k.String())

	_, err := w.Write([]byte(ks))
	if err != nil {
		log.Fatalf("error marshaling: %v", err)
	}

	lf := v.Type().NumField()
	t := v.Type()

	for i := 0; i < lf; i++ {
		f := t.Field(i)
		tag := f.Tag.Get("ini")

		vs := fmt.Sprintf("%s = %s\n", tag, v.Field(i).String())
		_, err := w.Write([]byte(vs))
		if err != nil {
			log.Fatalf("error marshaling: %v", err)
		}
	}

	_, err = w.Write([]byte("\n"))
	if err != nil {
		log.Fatalf("error marshaling: %v", err)
	}

	return nil
}
