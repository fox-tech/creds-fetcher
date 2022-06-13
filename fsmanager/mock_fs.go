package fsmanager

import "path"

type MockFileSystem struct {
	Files    map[string][]byte
	ReadErr  error
	WriteErr error
}

func NewMock() MockFileSystem {
	return MockFileSystem{
		Files: map[string][]byte{},
	}
}

func (m MockFileSystem) ReadFile(dir, filename string) ([]byte, error) {
	fp := path.Join(dir, filename)

	if m.ReadErr != nil {
		return []byte{}, m.ReadErr
	}

	if data, ok := m.Files[fp]; ok {
		return data, nil
	}

	return []byte{}, nil

}

func (m MockFileSystem) WriteFile(name string, data []byte) error {
	if m.WriteErr != nil {
		return m.WriteErr
	}

	m.Files[name] = data
	return nil
}
