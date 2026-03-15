package pkl

import (
	"io"
	"testing"
)

type writerMock struct {
	err   error
	bytes int
}

func (w writerMock) Write(_ []byte) (n int, err error) {
	return w.bytes, w.err
}

func TestLogger(t *testing.T) {

	tests := map[string]struct {
		writerMock io.Writer
		msg        string
		frameURI   string
	}{
		"should successfully log a message as trace and warn": {
			writerMock: writerMock{err: nil, bytes: 20},
			msg:        "test message",
			frameURI:   "test",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lgr := NewLogger(tc.writerMock)
			lgr.Trace(tc.msg, tc.frameURI)
			lgr.Warn(tc.msg, tc.frameURI)
		})
	}
}
