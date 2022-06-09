package configuration

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func Test_getReader(t *testing.T) {
	var toClose []*os.File
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		prep    func() error
		wantR   io.ReadCloser
		wantErr bool
	}{
		{
			name: "stdin",
			args: args{
				src: "stdin",
			},
			prep: func() (err error) {
				var tmp *os.File
				if tmp, err = createTestTempFile(`{"a":"foo"}`); err != nil {
					return
				}

				os.Stdin = tmp
				toClose = append(toClose, tmp)
				return
			},
		},
		{
			name: "foo.json",
			args: args{
				src: "./foo.json",
			},
			prep: func() (err error) {
				var tmp *os.File
				if tmp, err = createTestFile("./foo.json", `{"a":"foo"}`); err != nil {
					return
				}

				toClose = append(toClose, tmp)
				return
			},
		},
	}

	type testStruct struct {
		A string `json:"a"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.prep(); err != nil {
				t.Fatal(err)
			}

			gotR, err := getReader(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getReader() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			var ts testStruct
			if err = json.NewDecoder(gotR).Decode(&ts); err != nil {
				t.Fatal(err)
			}
		})
	}

	closeAndRemove(toClose)
}

func Test_getStdinReader(t *testing.T) {
	var toClose []*os.File
	tests := []struct {
		name         string
		prep         func() error
		wantContents string
		wantErr      bool
	}{
		{
			name: "with contents",
			prep: func() (err error) {
				var tmp *os.File
				if tmp, err = createTestTempFile(`{"a":"foo"}`); err != nil {
					return
				}

				os.Stdin = tmp
				toClose = append(toClose, tmp)
				return
			},
			wantContents: `{"a":"foo"}`,
		},
		{
			name: "without contents",
			prep: func() (err error) {
				return
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.prep(); err != nil {
				t.Fatal(err)
			}

			gotR, err := getStdinReader()
			if (err != nil) != tt.wantErr {
				t.Fatalf("getStdinReader() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			gotBS, err := ioutil.ReadAll(gotR)
			if err != nil {
				t.Fatal(err)
			}

			gotString := string(gotBS)
			if gotString != tt.wantContents {
				t.Fatalf("getStdinReader() received = <%s>, wantContents = <%s>", gotString, tt.wantContents)
			}
		})
	}

	closeAndRemove(toClose)
}

func closeAndRemove(fs []*os.File) {
	for _, f := range fs {
		name := f.Name()
		os.Remove(name)
		f.Close()
	}

}

func createTestTempFile(str string) (tmp *os.File, err error) {
	if tmp, err = os.CreateTemp("", ""); err != nil {
		return
	}

	err = writeAndResetTestFile(tmp, str)
	return
}

func createTestFile(destination, str string) (tmp *os.File, err error) {
	if tmp, err = os.Create(destination); err != nil {
		return
	}

	err = writeAndResetTestFile(tmp, str)
	return
}

func writeAndResetTestFile(f *os.File, str string) (err error) {
	if _, err = f.WriteString(str); err != nil {
		return
	}

	_, err = f.Seek(0, 0)
	return
}
