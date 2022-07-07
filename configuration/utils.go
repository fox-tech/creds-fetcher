package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/BurntSushi/toml"
)

func getConfigurations(overrideLocation string) (cfgs map[string]*Configuration, err error) {
	var (
		source io.ReadSeekCloser
		key    string
	)

	if source, key, err = getSource(overrideLocation); err != nil {
		return
	}
	defer source.Close()

	if cfgs, err = parseReader(source); err != nil {
		err = fmt.Errorf("error parsing <%s>: %v", key, err)
		return
	}

	return
}

func getSource(overrideLocation string) (r io.ReadSeekCloser, key string, err error) {
	if len(overrideLocation) > 0 {
		key = overrideLocation
		r, err = getReader(key)
		return
	}

	for _, source := range sources {
		if source, err = replaceTilde(source); err != nil {
			return
		}

		if r, err = getReader(source); err == nil {
			key = source
			return
		}
	}

	err = ErrNoConfiguration
	return
}

func getReader(src string) (r io.ReadSeekCloser, err error) {
	if src == "stdin" {
		return getStdinReader()
	}

	return os.Open(src)
}

func getStdinReader() (r io.ReadSeekCloser, err error) {
	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, os.Stdin); err != nil {
		return
	}
	reader := bytes.NewReader(buf.Bytes())
	r = makeReadSeekCloser(reader)
	return
}

func getReaderLength(r io.ReadSeeker) (n int64, err error) {
	if n, err = r.Seek(0, io.SeekEnd); err != nil {
		return
	}

	_, err = r.Seek(0, io.SeekStart)
	return
}

func parseReader(r io.ReadSeeker) (cfgs map[string]*Configuration, err error) {
	var length int64
	if length, err = getReaderLength(r); err != nil {
		return
	}

	if length == 0 {
		err = ErrEmptyConfigurationFile
		return
	}

	for _, d := range decoders {
		if cfgs, err = d.decodeOrReset(r); err == nil {
			return
		}
	}

	err = ErrCannotParseConfigurationFile
	return
}

func decodeAsTOML(r io.Reader) (cfgs map[string]*Configuration, err error) {
	_, err = toml.NewDecoder(r).Decode(&cfgs)
	return
}

func decodeAsJSON(r io.Reader) (cfgs map[string]*Configuration, err error) {
	err = json.NewDecoder(r).Decode(&cfgs)
	return
}

func replaceTilde(str string) (out string, err error) {
	if strings.IndexByte(str, '~') == -1 {
		out = str
		return
	}

	var u *user.User
	if u, err = user.Current(); err != nil {
		return
	}

	dir := u.HomeDir
	out = strings.Replace(str, "~", dir, 1)
	return
}
