package configuration

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_decoder_decodeOrReset(t *testing.T) {
	type args struct {
		getReader func() (io.ReadSeeker, error)
	}

	tests := []struct {
		name    string
		d       decoder
		args    args
		wantCfg *Configuration
		wantErr bool
	}{
		{
			name: "success",
			d:    decodeAsJSON,
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					r = strings.NewReader(exampleJSON)
					return
				},
			},
			wantCfg: exampleConfiguration,
		},
		{
			name: "failure (closed)",
			d:    decodeAsJSON,
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					return createTestClosedTempfile()
				},
			},
			wantErr: true,
		},
		{
			name: "failure (decode fail, reset success)",
			d:    decodeAsJSON,
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					r = strings.NewReader(exampleTOML)
					return
				},
			},
			wantErr: true,
		},
		{
			name: "failure (error seeking)",
			d:    decodeAsJSON,
			args: args{
				getReader: func() (r io.ReadSeeker, err error) {
					r = &testBadSeeker{
						ReadSeeker: strings.NewReader(exampleTOML),
					}

					return
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
				t.Error()
			}

			if f, ok := r.(*os.File); ok {
				defer os.Remove(f.Name())
				defer f.Close()
			}

			gotCfg, err := tt.d.decodeOrReset(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("decoder.decodeOrReset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(gotCfg, tt.wantCfg) {
				t.Errorf("decoder.decodeOrReset() = %v, want %v", gotCfg, tt.wantCfg)
			}
		})
	}
}

type testBadSeeker struct {
	io.ReadSeeker
}

func (t *testBadSeeker) Seek(offset int64, whence int) (n int64, err error) {
	err = os.ErrClosed
	return
}
