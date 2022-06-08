package aws

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func Test_ReadCredentialsFile(t *testing.T) {
	tempExistingDir := "awstest"
	tempExistingFile := "testcred"
	testData := []byte("test aws credentials")

	type args struct {
		mockDir  string
		mockName string
	}

	type expect struct {
		err  error
		data []byte
	}

	tests := []struct {
		name   string
		args   args
		expect expect
	}{
		{
			name: "directory doesn't exists: creates directory and file",
			args: args{
				mockDir:  "tempTestDir",
				mockName: "tempTestFile",
			},
			expect: expect{
				err:  nil,
				data: []byte{},
			},
		},
		{
			name: "directory exists, file doesn't: creates file",
			args: args{
				mockDir:  "tempTestDir",
				mockName: "tempTestFile2",
			},
			expect: expect{
				err:  nil,
				data: []byte{},
			},
		},
		{
			name: "file exists: reads file data",
			args: args{
				mockDir:  tempExistingDir,
				mockName: tempExistingFile,
			},
			expect: expect{
				err:  nil,
				data: testData,
			},
		},
	}

	// Create file to test data
	h, _ := os.UserHomeDir()
	os.Chdir(h)
	os.Mkdir(tempExistingDir, 0766)
	f, _ := os.Create(filepath.Join(tempExistingDir, tempExistingFile))
	f.Write(testData)
	f.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := readCredentialsFile(tt.args.mockDir, tt.args.mockName)

			if err != tt.expect.err {
				t.Errorf("readCredentials() expected error: %s, got: %s", tt.expect.err, err)
			}

			if !bytes.Equal(data, tt.expect.data) {
				t.Errorf("readCredentials() expected data: %s, got: %s", tt.expect.err, err)
			}
		})
	}

	os.RemoveAll(tempExistingDir)
	os.RemoveAll("tempTestDir")
}
