package ini

import (
	"bufio"
	"bytes"
	"errors"
	"reflect"
	"testing"
)

type testStruct struct {
	Fox string `ini:"fox"`
	Dog string `ini:"dog"`
	Cat string `ini:"cat"`
}

type noTagStruct struct {
	Fish string
}

const testDataFull = "myFirstPet"
const testDataEmptyField = "mySecondPet"

var testDataStruct = map[string]testStruct{
	testDataFull: {
		Fox: "the lazy fox",
		Dog: "woof woof!",
		Cat: "miaw",
	},
	testDataEmptyField: {
		Fox: "not lazy fox",
		Dog: "guau",
	},
}

var testDataBytes = map[string][]byte{
	testDataFull:       []byte("[myFirstPet]\nfox = the lazy fox\ndog = woof woof!\ncat = miaw\n\n"),
	testDataEmptyField: []byte("[mySecondPet]\nfox = not lazy fox\ndog = guau\ncat = \n\n"),
}

func TestMarshall(t *testing.T) {
	type expect struct {
		data [][]byte
		err  error
	}

	tests := []struct {
		name string
		arg  interface{}
		expect
	}{
		{
			name: "send map: returns bytes with data",
			arg:  testDataStruct,
			expect: expect{
				data: [][]byte{
					testDataBytes[testDataFull],
					testDataBytes[testDataEmptyField],
				},
				err: nil,
			},
		},
		{
			name: "send not supported type: returns error",
			arg:  "try",
			expect: expect{
				data: nil,
				err:  ErrUnsupportedType,
			},
		},
		{
			name: "send not supported map type: returns error",
			arg: map[string]string{
				"not": "work",
			},
			expect: expect{
				data: nil,
				err:  ErrUnsupportedType,
			},
		},
		{
			name: "send struct without tag: returns error",
			arg: map[string]noTagStruct{
				"fish1": {Fish: "swims"},
			},
			expect: expect{
				data: nil,
				err:  ErrMissingTag,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Marshal(tt.arg)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("Marshall() expected error: %s, got: %s", tt.expect.err, err)
			}

			// since the marshal handles the elements in the map in a unspecified order
			// verify that the data is in there.
			for _, b := range tt.expect.data {
				if !bytes.Contains(data, b) {
					t.Errorf("Marshall expected data to contain: %v, got: %v", b, data)
				}
			}
		})
	}
}

func TestWriteData(t *testing.T) {
	type expect struct {
		data []byte
		err  error
	}

	type args struct {
		w *bytes.Buffer
		k reflect.Value
		v reflect.Value
	}

	tests := []struct {
		name string
		args
		expect
	}{
		{
			name: "send data: should write it to the byte buffer",
			args: args{
				w: bytes.NewBuffer([]byte{}),
				k: reflect.ValueOf(testDataFull),
				v: reflect.ValueOf(testDataStruct[testDataFull]),
			},
			expect: expect{
				data: testDataBytes[testDataFull],
				err:  nil,
			},
		},
		{
			name: "send data with invalid key: should return error",
			args: args{
				w: bytes.NewBuffer([]byte{}),
				k: reflect.ValueOf(23),
				v: reflect.ValueOf(testDataStruct[testDataFull]),
			},
			expect: expect{
				data: []byte{},
				err:  ErrUnsupportedType,
			},
		},
		{
			name: "send data with invalid value: should return error",
			args: args{
				w: bytes.NewBuffer([]byte{}),
				k: reflect.ValueOf(testDataFull),
				v: reflect.ValueOf("this is not a struct"),
			},
			expect: expect{
				data: []byte{},
				err:  ErrUnsupportedType,
			},
		},
		{
			name: "send data with field without tag: should return error",
			args: args{
				w: bytes.NewBuffer([]byte{}),
				k: reflect.ValueOf("myPet"),
				v: reflect.ValueOf(noTagStruct{
					Fish: "swims",
				}),
			},
			expect: expect{
				data: []byte("[myPet]\n"),
				err:  ErrMissingTag,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := write(tt.args.w, tt.args.k, tt.args.v)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("write() expected error: %s, got: %s", tt.expect.err, err)
			}

			data := tt.args.w.Bytes()
			if !bytes.Equal(data, tt.expect.data) {
				t.Errorf("write() expected data: %v, got: %v", tt.expect.data, data)
			}
		})
	}
}

func TestParseName(t *testing.T) {
	type args struct {
		data []byte
		v    reflect.Value
	}

	type expect struct {
		v   reflect.Value
		err error
	}

	name := "lorem ipsum"
	empty := ""
	number := 3

	tests := []struct {
		name string
		args
		expect
	}{
		{
			name: "send correct data: data is read into value",
			args: args{
				data: []byte("[" + name + "]"),
				v:    reflect.ValueOf(&empty),
			},
			expect: expect{
				v:   reflect.ValueOf(&name),
				err: nil,
			},
		},
		{
			name: "send invalid data: error is returned",
			args: args{
				data: []byte(name),
				v:    reflect.ValueOf(&empty),
			},
			expect: expect{
				v:   reflect.ValueOf(&empty),
				err: ErrInvalidContent,
			},
		},
		{
			name: "send empty data: error is returned",
			args: args{
				data: []byte(empty),
				v:    reflect.ValueOf(&empty),
			},
			expect: expect{
				v:   reflect.ValueOf(&empty),
				err: ErrInvalidContent,
			},
		},
		{
			name: "send nil data: error is returned",
			args: args{
				data: nil,
				v:    reflect.ValueOf(&empty),
			},
			expect: expect{
				v:   reflect.ValueOf(&empty),
				err: ErrInvalidContent,
			},
		},
		{
			name: "send unsupported type: error is returned",
			args: args{
				data: nil,
				v:    reflect.ValueOf(&number),
			},
			expect: expect{
				v:   reflect.ValueOf(&number),
				err: ErrUnsupportedType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseName(tt.args.data, tt.args.v)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("parseName() expected error: %s, got: %s", tt.expect.err, err)
			}

			if tt.args.v.String() != tt.expect.v.String() {
				t.Errorf("parseName() expected data: %v, got: %v", tt.expect.v.Elem(), tt.args.v.Elem())
			}
		})
	}
}

func TestReadFields(t *testing.T) {
	type expect struct {
		fields map[string]string
		err    error
	}

	tests := []struct {
		name string
		arg  []byte
		expect
	}{
		{
			name: "send valid data: fields are read correctly",
			arg:  []byte("key1=value1\n key2 = value2\n emptyKey = \n"),
			expect: expect{
				fields: map[string]string{
					"key1":     "value1",
					"key2":     "value2",
					"emptyKey": "",
				},
				err: nil,
			},
		},
		{
			name: "send invalid data: error is returned",
			arg:  []byte("key1=value1\n key2 = value2\n emptyKey\n"),
			expect: expect{
				fields: nil,
				err:    ErrInvalidContent,
			},
		},
		{
			name: "send empty data: empty map is returned, no error is returned",
			arg:  []byte{},
			expect: expect{
				fields: map[string]string{},
				err:    nil,
			},
		},
		{
			name: "send nil data: empty map is returned, no error is returned",
			arg:  []byte{},
			expect: expect{
				fields: map[string]string{},
				err:    nil,
			},
		},
		{
			name: "send valid data ending in equals: fields are read correctly",
			arg:  []byte("key1=value1\n key2 = value2==\n emptyKey = \n"),
			expect: expect{
				fields: map[string]string{
					"key1":     "value1",
					"key2":     "value2==",
					"emptyKey": "",
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := readFields(tt.arg)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("readFields() expected error: %s, got: %s", tt.expect.err, err)
			}

			if !reflect.DeepEqual(fields, tt.expect.fields) {
				t.Errorf("readFields() expected data: %v, got: %v", tt.expect.fields, fields)
			}
		})
	}

}

func TestParseAttributes(t *testing.T) {
	type args struct {
		data []byte
		v    reflect.Value
	}

	type expect struct {
		v   reflect.Value
		err error
	}

	vtype := reflect.TypeOf(testDataStruct).Elem()
	ktype := reflect.TypeOf(testDataStruct).Key()

	nts := noTagStruct{
		Fish: "swims",
	}
	ntstype := reflect.TypeOf(nts)

	tests := []struct {
		name string
		args
		expect
	}{
		{
			name: "send data: attributes are loaded into value",
			args: args{
				data: []byte("\nfox = the lazy fox\ndog = woof woof!\ncat = miaw\n\n"),
				v:    reflect.New(vtype).Elem(),
			},
			expect: expect{
				v:   reflect.ValueOf(testDataStruct[testDataFull]),
				err: nil,
			},
		},
		{
			name: "send empty data: attributes are loaded empty",
			args: args{
				data: []byte(""),
				v:    reflect.New(vtype).Elem(),
			},
			expect: expect{
				v:   reflect.New(vtype).Elem(),
				err: nil,
			},
		},
		{
			name: "send unsupported type: error is returned",
			args: args{
				data: []byte("\nfox = the lazy fox\ndog = woof woof!\ncat = miaw\n\n"),
				v:    reflect.New(ktype).Elem(),
			},
			expect: expect{
				v:   reflect.New(ktype).Elem(),
				err: ErrUnsupportedType,
			},
		},
		{
			name: "send invalid data: error is returned",
			args: args{
				data: []byte("\nfox = the lazy fox\ndog = woof woof!\ncat miaw\n\n"),
				v:    reflect.New(vtype).Elem(),
			},
			expect: expect{
				v:   reflect.New(vtype).Elem(),
				err: ErrInvalidContent,
			},
		},
		{
			name: "send struct without tag: error is returned",
			args: args{
				data: []byte("fish = swims"),
				v:    reflect.New(ntstype).Elem(),
			},
			expect: expect{
				v:   reflect.New(ntstype).Elem(),
				err: ErrMissingTag,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAttributes(tt.args.data, tt.args.v)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("parseAttributes() expected error: %s, got: %s", tt.expect.err, err)
			}

			exp := tt.expect.v.Interface()
			got := tt.args.v.Interface()

			if !reflect.DeepEqual(exp, got) {
				t.Errorf("parseAttributes() expected data: %v, got: %v", exp, got)
			}
		})
	}
}

func TestUnmarshall(t *testing.T) {
	type args struct {
		data []byte
		v    interface{}
	}

	type expect struct {
		data interface{}
		err  error
	}

	tests := []struct {
		name string
		args
		expect
	}{
		{
			name: "send data: data is loaded in v",
			args: args{
				data: append(testDataBytes[testDataFull], testDataBytes[testDataEmptyField]...),
				v:    map[string]testStruct{},
			},
			expect: expect{
				data: testDataStruct,
				err:  nil,
			},
		},
		{
			name: "send data with emtpy fields: data is left empty in v",
			args: args{
				data: []byte("[myFirstPet]\n"),
				v:    map[string]testStruct{},
			},
			expect: expect{
				data: map[string]testStruct{"myFirstPet": {}},
				err:  nil,
			},
		},
		{
			name: "send unsupported data: error is returned",
			args: args{
				data: append(testDataBytes[testDataFull], testDataBytes[testDataEmptyField]...),
				v:    testStruct{},
			},
			expect: expect{
				data: testStruct{},
				err:  ErrUnsupportedType,
			},
		},
		{
			name: "send unsupported data: error is returned",
			args: args{
				data: testDataBytes[testDataFull],
				v:    map[int]testStruct{},
			},
			expect: expect{
				data: map[int]testStruct{},
				err:  ErrUnsupportedType,
			},
		},
		{
			name: "send invalid data: error is returned",
			args: args{
				data: []byte("[myFirstPet]\nfox = the lazy fox\ncat\n\n"),
				v:    map[string]testStruct{},
			},
			expect: expect{
				data: map[string]testStruct{},
				err:  ErrInvalidContent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal(tt.args.data, tt.args.v)

			if !errors.Is(err, tt.expect.err) {
				t.Errorf("Unmarshal() expected error: %s, got: %s", tt.expect.err, err)
			}

			exp := tt.expect.data
			got := tt.args.v

			if !reflect.DeepEqual(exp, got) {
				t.Errorf("Unmarshal() expected data: %v, got: %v", exp, got)
			}
		})
	}
}

func TestCreateSplit(t *testing.T) {
	data := []byte("[myFirstPet]\nfox = the lazy fox\n\n[mySecondPet]\nfox = the not lazy fox\n\n")

	idx := [][]int{
		{0, 11},
		{33, 45},
	}
	totalBytes := len(data)
	split := createSplit(idx, totalBytes)

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Split(split)

	buf := make([]byte, 2)
	sc.Buffer(buf, bufio.MaxScanTokenSize)

	// assert function works as expected, numers are calculated from the byte slice
	tests := []struct {
		name   string
		arg    bool
		expect []byte
	}{
		{
			name:   "read first token",
			expect: []byte("\nfox = the lazy fox\n\n"),
		},
		{
			name:   "read second token",
			expect: []byte("\nfox = the not lazy fox\n\n"),
		},
		{
			name:   "reached EOF, stop reading",
			arg:    true,
			expect: nil,
		},
	}

	for _, tt := range tests {
		sc.Scan()

		tkn := sc.Bytes()
		if !bytes.Equal(tkn, tt.expect) {
			t.Errorf("createSplit() expected token: %s, got: %s", tt.expect, tkn)
		}
	}
}
