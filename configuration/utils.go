package configuration

import (
	"encoding/json"
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

func getConfiguration() (cfg *Configuration, err error) {
	for _, fn := range getFuncs {
		cfg, err = fn()
		switch {
		case err == nil:
			return
		case os.IsNotExist(err):

		default:
			return
		}
	}

	err = ErrNoConfiguration
	return
}

func getAsJSON() (cfg *Configuration, err error) {
	var (
		c Configuration
		r io.ReadCloser
	)

	if r, err = os.Open("./configuration.json"); err != nil {
		return
	}
	defer r.Close()

	if err = json.NewDecoder(r).Decode(&c); err != nil {
		return
	}

	cfg = &c
	return
}

func getAsTOML() (cfg *Configuration, err error) {
	var (
		c Configuration
		r io.ReadCloser
	)

	if r, err = os.Open("./configuration.toml"); err != nil {
		return
	}
	defer r.Close()

	if _, err = toml.NewDecoder(r).Decode(&c); err != nil {
		return
	}

	cfg = &c
	return
}
