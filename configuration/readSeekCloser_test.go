package configuration

import (
	"bytes"
	"io"
	"testing"
)

func Test_readSeekCloser_Close(t *testing.T) {
	type fields struct {
		ReadSeeker io.ReadSeeker
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				ReadSeeker: bytes.NewReader([]byte("hello world")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := readSeekCloser{
				ReadSeeker: tt.fields.ReadSeeker,
			}

			if err := r.Close(); (err != nil) != tt.wantErr {
				t.Errorf("readSeekCloser.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
