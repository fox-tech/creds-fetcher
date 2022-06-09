package fsmanager

import "path"

type mockFileSystem struct {
	files    map[string][]byte
	readErr  error
	writeErr error
}

func NewMock(files map[string][]byte, re, we error) mockFileSystem {
	return mockFileSystem{
		files:    files,
		writeErr: we,
		readErr:  re,
	}
}

func (m mockFileSystem) ReadFile(dir, filename string) ([]byte, error) {
	fp := path.Join(dir, filename)

	if m.readErr != nil {
		return []byte{}, m.readErr
	}

	if data, ok := m.files[fp]; ok {
		return data, nil
	}

	return []byte{}, nil

}

func (m mockFileSystem) WriteFile(name string, data []byte) error {
	if m.writeErr != nil {
		return m.writeErr
	}

	m.files[name] = data
	return nil
}
