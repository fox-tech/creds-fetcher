package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

func getConfiguration() (cfg *Configuration, err error) {
	var (
		source io.ReadSeekCloser
		key    string
	)

	if source, key, err = getSource(); err != nil {
		return
	}
	defer source.Close()

	if cfg, err = parseReader(source); err != nil {
		err = fmt.Errorf("error parsing <%s>: %v", key, err)
		return
	}

	return
}

func getSource() (r io.ReadSeekCloser, key string, err error) {
	for _, source := range sources {
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

	var n int64
	if n, err = io.Copy(buf, os.Stdin); err != nil {
		return
	}

	if n == 0 {
		err = io.EOF
		return
	}

	reader := bytes.NewReader(buf.Bytes())
	r = makeReadSeekCloser(reader)
	return
}

func parseReader(r io.ReadSeeker) (cfg *Configuration, err error) {
	for _, decoder := range decoders {
		if cfg, err = decoder(r); err == nil {
			return
		}

		if _, err = r.Seek(0, 0); err != nil {
			return
		}
	}

	err = ErrCannotParseConfigurationFile
	return
}

func decodeAsTOML(r io.Reader) (cfg *Configuration, err error) {
	var c Configuration
	if _, err = toml.NewDecoder(r).Decode(&c); err != nil {
		return
	}

	cfg = &c
	return
}

func decodeAsJSON(r io.Reader) (cfg *Configuration, err error) {
	var c Configuration
	if err = json.NewDecoder(r).Decode(&c); err != nil {
		return
	}

	cfg = &c
	return
}

type decoder func(io.Reader) (*Configuration, error)
