package env

import (
	"os"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	baseEnv := map[string]string{
		"a":           "foo",
		"b":           "bar",
		"c":           "baz",
		"ignoreField": "shouldn't exist",
	}

	baseStruct := testStruct{
		A: "foo",
		B: "bar",
		C: "baz",
	}

	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		env     map[string]string
		args    args
		want    func(got interface{}) (want interface{}, ok bool)
		wantErr bool
	}{
		{
			env: baseEnv,
			args: args{
				value: testStruct{},
			},
			wantErr: true,
		},
		{
			env: baseEnv,
			args: args{
				value: &invalidTestStruct{},
			},
			wantErr: true,
		},
		{
			env: baseEnv,
			args: args{
				value: &testStruct{},
			},
			want: func(got interface{}) (want interface{}, ok bool) {
				gotAsserted := got.(testStruct)
				want = baseStruct
				ok = gotAsserted == want
				return
			},
		},
		{
			env: baseEnv,
			args: args{
				value: makeDoublePointer(),
			},
			want: func(got interface{}) (want interface{}, ok bool) {
				gotAsserted := got.(testStruct)
				want = baseStruct
				ok = gotAsserted == want
				return
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			err := Unmarshal(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func makeDoublePointer() **testStruct {
	var t testStruct
	tp := &t
	return &tp
}

func makeGetter() testGetter {
	var t testStruct
	return &t
}

type testStruct struct {
	A string `env:"a"`
	B string `env:"b"`
	C string `env:"c"`

	IgnoreField string
}

func (t *testStruct) Get(key string) (value string) {
	switch key {
	case "a":
		return t.A
	case "b":
		return t.B
	case "c":
		return t.C
	default:
		return
	}
}

type invalidTestStruct struct {
	A string `env:"a"`
	B int    `env:"b"`
	C string `env:"c"`

	IgnoreField string
}

type testGetter interface {
	Get(key string) (value string)
}
