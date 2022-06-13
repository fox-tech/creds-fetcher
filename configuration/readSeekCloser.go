package configuration

import "io"

func makeReadSeekCloser(r io.ReadSeeker) (rsc io.ReadSeekCloser) {
	return readSeekCloser{r}
}

type readSeekCloser struct {
	io.ReadSeeker
}

func (readSeekCloser) Close() error { return nil }
