package configuration

import "io"

type decoder func(io.Reader) (*Configuration, error)

func (d decoder) decodeOrReset(r io.ReadSeeker) (cfg *Configuration, err error) {
	if cfg, err = d(r); err == nil {
		return
	}

	if _, seekErr := r.Seek(0, 0); seekErr != nil {
		err = seekErr
		return
	}

	return
}
