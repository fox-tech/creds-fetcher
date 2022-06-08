package env

import (
	"os"
	"testing"
)

type testStruct struct {
	A string `env:"a"`
	B string `env:"b"`
	C string `env:"c"`
}

func TestUnmarshal(t *testing.T) {
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
			env: map[string]string{
				"a": "foo",
				"b": "bar",
				"c": "baz",
			},
			args: args{
				value: testStruct{},
			},
			wantErr: true,
		},
		{
			env: map[string]string{
				"a": "foo",
				"b": "bar",
				"c": "baz",
			},
			args: args{
				value: &testStruct{},
			},
			want: func(got interface{}) (want interface{}, ok bool) {
				gotAsserted := got.(testStruct)
				want = testStruct{
					A: "foo",
					B: "bar",
					C: "baz",
				}

				ok = gotAsserted == want
				return
			},
		},
		{
			env: map[string]string{
				"a": "foo",
				"b": "bar",
				"c": "baz",
			},
			args: args{
				value: makeDoublePointer(),
			},
			want: func(got interface{}) (want interface{}, ok bool) {
				gotAsserted := got.(testStruct)
				want = testStruct{
					A: "foo",
					B: "bar",
					C: "baz",
				}

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

			if err := Unmarshal(tt.args.value); (err != nil) != tt.wantErr {
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
