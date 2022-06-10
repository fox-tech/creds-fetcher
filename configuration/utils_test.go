package configuration

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_getConfiguration(t *testing.T) {
	type args struct {
		overrideLocation string
	}

	tests := []struct {
		name    string
		args    args
		prep    func() (toRemove *os.File, err error)
		wantCfg *Configuration
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				overrideLocation: "",
			},
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestTempFile(exampleJSON)
				os.Stdin = toRemove
				return
			},
			wantCfg: exampleConfiguration,
		},
		{
			name: "success (with override location)",
			args: args{
				overrideLocation: "./Test_getConfiguration.override.json",
			},
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestFile("./Test_getConfiguration.override.json", exampleJSON)
				return
			},
			wantCfg: exampleConfiguration,
		},
		{
			name: "success (with override values)",
			args: args{},
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestTempFile(exampleJSON)
				os.Stdin = toRemove
				os.Setenv("aws_provider_arn", "1n")
				os.Setenv("aws_role_arn", "2n")
				return
			},
			wantCfg: &Configuration{
				AWSProviderARN: "1n",
				AWSRoleARN:     "2n",
				OktaClientID:   "3",
				OktaAppID:      "4",
				OktaURL:        "5",
			},
		},
		{
			name: "failure (closed file)",
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestClosedTempfile()
				os.Stdin = toRemove
				return
			},
			wantErr: true,
		},
		{
			name: "failure (decode error)",
			prep: func() (toRemove *os.File, err error) {
				toRemove, err = createTestTempFile("hello world")
				os.Stdin = toRemove
				return
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				toRemove *os.File
				err      error
			)

			os.Clearenv()

			if toRemove, err = tt.prep(); err != nil {
				t.Errorf("getConfiguration() error preparing test: %v", err)
				return
			}

			if toRemove != nil {
				defer os.Remove(toRemove.Name())
				defer toRemove.Close()
			}

			gotCfg, err := getConfiguration(tt.args.overrideLocation)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCfg, tt.wantCfg) {
				t.Errorf("getConfiguration() = %v, want %v", gotCfg, tt.wantCfg)
			}
		})
	}
}

func Test_getSource(t *testing.T) {
	type args struct {
		overrideLocation string
	}

	tests := []struct {
		name string
		args args
		prep func() (toRemove *os.File, err error)

		wantKey string
		wantErr bool
	}{
		{
			name: "success (stdin)",
			prep: func() (tmp *os.File, err error) {
				if tmp, err = createTestTempFile(`{"a":"foo"}`); err != nil {
					return
				}

				os.Stdin = tmp
				return
			},
			wantKey: "stdin",
		},
		{
			name: "success (config.json)",
			prep: func() (tmp *os.File, err error) {
				return createTestFile("./config.json", `{"a":"foo"}`)
			},
			wantKey: "./config.json",
		},
		{
			name: "success (config.toml)",
			prep: func() (tmp *os.File, err error) {
				return createTestFile("./config.toml", `{"a":"foo"}`)
			},
			wantKey: "./config.toml",
		},
		{
			name: "success (override location)",
			args: args{
				overrideLocation: "./Test_getSource.override.json",
			},
			prep: func() (tmp *os.File, err error) {
				return createTestFile("./Test_getSource.override.json", exampleJSON)
			},
			wantKey: "./Test_getSource.override.json",
		},
		{
			name: "failure (no config)",
			prep: func() (tmp *os.File, err error) {
				return nil, nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				toRemove *os.File
				err      error
			)

			if toRemove, err = tt.prep(); err != nil {
				t.Fatal(err)
			}

			if toRemove != nil {
				defer os.Remove(toRemove.Name())
				defer toRemove.Close()
			}

			_, gotKey, err := getSource(tt.args.overrideLocation)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotKey != tt.wantKey {
				t.Errorf("getSource() gotKey = %v, want %v", gotKey, tt.wantKey)
				return
			}
		})
	}
}

func Test_getReader(t *testing.T) {
	type args struct {
		src string
	}

	tests := []struct {
		name string
		args args
		prep func() (toRemove *os.File, err error)

		wantR   io.ReadCloser
		wantErr bool
	}{
		{
			name: "stdin",
			args: args{
				src: "stdin",
			},
			prep: func() (tmp *os.File, err error) {
				if tmp, err = createTestTempFile(`{"a":"foo"}`); err != nil {
					return
				}

				os.Stdin = tmp
				return
			},
		},
		{
			name: "stdin (closed)",
			args: args{
				src: "stdin",
			},
			prep: func() (tmp *os.File, err error) {
				if tmp, err = createTestTempFile(`{"a":"foo"}`); err != nil {
					return
				}

				os.Stdin = tmp
				if err = tmp.Close(); err != nil {
					return
				}

				return
			},
			wantErr: true,
		},
		{
			name: "foo.json",
			args: args{
				src: "./foo.json",
			},
			prep: func() (tmp *os.File, err error) {
				return createTestFile("./foo.json", `{"a":"foo"}`)
			},
		},
	}

	type testStruct struct {
		A string `json:"a"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toRemove, err := tt.prep()
			if err != nil {
				t.Fatal(err)
			}
			if toRemove != nil {
				defer os.Remove(toRemove.Name())
				defer toRemove.Close()
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
}

func Test_getStdinReader(t *testing.T) {
	tests := []struct {
		name string
		prep func() (toRemove *os.File, err error)

		wantContents string
		wantErr      bool
	}{
		{
			name: "with contents",
			prep: func() (tmp *os.File, err error) {
				if tmp, err = createTestTempFile(`{"a":"foo"}`); err != nil {
					return
				}

				os.Stdin = tmp
				return
			},
			wantContents: `{"a":"foo"}`,
		},
		{
			name: "without contents",
			prep: func() (tmp *os.File, err error) {
				return
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toRemove, err := tt.prep()
			if err != nil {
				t.Fatal(err)
			}
			if toRemove != nil {
				defer os.Remove(toRemove.Name())
				defer toRemove.Close()
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
}

func Test_getReaderLength(t *testing.T) {
	type args struct {
		getReader func() (io.ReadSeeker, error)
	}

	tests := []struct {
		name    string
		args    args
		wantN   int64
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					r = strings.NewReader("hello world")
					return
				},
			},
			wantN: 11,
		},
		{
			name: "failure (closed file)",
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					return createTestClosedTempfile()
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				r   io.ReadSeeker
				err error
			)

			if r, err = tt.args.getReader(); err != nil {
				t.Errorf("getReaderLength() error getting reader: %v", err)
				return
			}

			if f, ok := r.(*os.File); ok {
				defer os.Remove(f.Name())
				defer f.Close()
			}

			gotN, err := getReaderLength(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("getReaderLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotN != tt.wantN {
				t.Errorf("getReaderLength() = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}

func Test_parseReader(t *testing.T) {
	type args struct {
		getReader func() (io.ReadSeeker, error)
	}

	tests := []struct {
		name    string
		args    args
		wantCfg *Configuration
		wantErr bool
	}{
		{
			name: "success (json)",
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					r = strings.NewReader(exampleJSON)
					return
				},
			},
			wantCfg: exampleConfiguration,
		},
		{
			name: "success (toml)",
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					r = strings.NewReader(exampleTOML)
					return
				},
			},
			wantCfg: exampleConfiguration,
		},
		{
			name: "failure (plaintext)",
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					r = strings.NewReader("hello world")
					return
				},
			},
			wantErr: true,
		},
		{
			name: "failure (json array)",
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					r = strings.NewReader(exampleJSONArray)
					return
				},
			},
			wantErr: true,
		},
		{
			name: "failure (no contents)",
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					r = strings.NewReader("")
					return
				},
			},
			wantErr: true,
		},
		{
			name: "failure (closed file)",
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					return createTestClosedTempfile()
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				r   io.ReadSeeker
				err error
			)

			if r, err = tt.args.getReader(); err != nil {
				t.Errorf("getReaderLength() error getting reader: %v", err)
				return
			}

			gotCfg, err := parseReader(r)
			if f, ok := r.(*os.File); ok {
				defer os.Remove(f.Name())
				defer f.Close()
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("parseReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCfg, tt.wantCfg) {
				t.Errorf("parseReader() = %v, want %v", gotCfg, tt.wantCfg)
			}
		})
	}
}

func Test_decodeAsTOML(t *testing.T) {
	type args struct {
		r io.Reader
	}

	tests := []struct {
		name    string
		args    args
		wantCfg *Configuration
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				r: bytes.NewBufferString(exampleTOML),
			},
			wantCfg: exampleConfiguration,
		},
		{
			name: "failure (json array)",
			args: args{
				r: bytes.NewBufferString(exampleJSONArray),
			},
			wantErr: true,
		},
		{
			name: "failure (json object)",
			args: args{
				r: bytes.NewBufferString(exampleJSON),
			},
			wantErr: true,
		},
		{
			name: "failure (plaintext)",
			args: args{
				r: bytes.NewBufferString(`hello world`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCfg, err := decodeAsTOML(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeAsTOML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCfg, tt.wantCfg) {
				t.Errorf("decodeAsTOML() = %v, want %v", gotCfg, tt.wantCfg)
			}
		})
	}
}

func Test_decodeAsJSON(t *testing.T) {
	type args struct {
		r io.Reader
	}

	tests := []struct {
		name    string
		args    args
		wantCfg *Configuration
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				r: bytes.NewBufferString(exampleJSON),
			},
			wantCfg: exampleConfiguration,
		},
		{
			name: "failure (array)",
			args: args{
				r: bytes.NewBufferString(exampleJSONArray),
			},
			wantErr: true,
		},
		{
			name: "failure (toml)",
			args: args{
				r: bytes.NewBufferString(exampleTOML),
			},
			wantErr: true,
		},
		{
			name: "failure (plaintext)",
			args: args{
				r: bytes.NewBufferString(`hello world`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCfg, err := decodeAsJSON(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeAsJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCfg, tt.wantCfg) {
				t.Errorf("decodeAsJSON() = %v, want %v", gotCfg, tt.wantCfg)
			}
		})
	}
}

func createTestTempFile(str string) (tmp *os.File, err error) {
	if tmp, err = os.CreateTemp("", ""); err != nil {
		return
	}

	err = writeAndResetTestFile(tmp, str)
	return
}

func createTestClosedTempfile() (tmp *os.File, err error) {
	if tmp, err = createTestTempFile("hello world"); err != nil {
		return
	}

	err = tmp.Close()
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
