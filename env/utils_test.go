package env

import (
	"reflect"
	"testing"
)

func Test_getTarget(t *testing.T) {
	type args struct {
		getValue func() interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantTarget interface{}
		wantOk     bool
	}{
		{
			name: "basic",
			args: args{
				getValue: func() interface{} {
					var t testStruct
					t.A = "foo"
					t.B = "bar"
					t.C = "baz"
					return &t
				},
			},
			wantTarget: testStruct{
				A: "foo",
				B: "bar",
				C: "baz",
			},
			wantOk: true,
		},
		{
			name: "interface",
			args: args{
				getValue: func() interface{} {
					return makeGetter()
				},
			},
			wantTarget: testStruct{},
			wantOk:     true,
		},
		{
			name: "map error",
			args: args{
				getValue: func() interface{} {
					return map[string]string{"a": "foo", "b": "bar", "c": "baz"}
				},
			},
			wantOk: false,
		},
		{
			name: "slice error",
			args: args{
				getValue: func() interface{} {
					return []string{"a", "b", "c"}
				},
			},
			wantOk: false,
		},
		{
			name: "int error",
			args: args{
				getValue: func() interface{} {
					return 1
				},
			},
			wantOk: false,
		},
		{
			name: "float error",
			args: args{
				getValue: func() interface{} {
					return 3.14
				},
			},
			wantOk: false,
		},
		{
			name: "bool error",
			args: args{
				getValue: func() interface{} {
					return true
				},
			},
			wantOk: false,
		},
		{
			name: "string error",
			args: args{
				getValue: func() interface{} {
					return "hello world"
				},
			},
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTarget, gotOk := getTarget(tt.args.getValue())
			if gotOk != tt.wantOk {
				t.Errorf("getTarget() gotOk = %v, want %v", gotOk, tt.wantOk)
			}

			if !gotOk {
				return
			}

			if !reflect.DeepEqual(gotTarget.Interface(), tt.wantTarget) {
				t.Errorf("getTarget() gotTarget = %T, want %T", gotTarget.Interface(), tt.wantTarget)
			}

		})
	}
}
