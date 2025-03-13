package helper

import (
	"errors"
	"io"
	"log/slog"
)

type DupReadCloser struct {
	oRdr         io.ReadCloser
	dupRdr       io.Reader
	dupWtr       io.WriteCloser
	dupWtrClosed bool
}

func NewDupReadCloser(oRdr io.ReadCloser) *DupReadCloser {
	r, w := io.Pipe()
	return &DupReadCloser{
		oRdr:         oRdr,
		dupRdr:       r,
		dupWtr:       w,
		dupWtrClosed: false,
	}
}

func (d DupReadCloser) DupRdr() io.Reader {
	return d.dupRdr
}

func (d *DupReadCloser) Read(p []byte) (n int, err error) {
	n, err = d.oRdr.Read(p)
	if n > 0 && !d.dupWtrClosed {
		nW, errW := d.dupWtr.Write(p[:n])
		if n != nW {
			slog.Warn("Incomplete Write to DupW", "exp", n, "act", nW)
		}
		if errW != nil {
			slog.Warn("Error Writing to DupW", "error", errW)
		}
	}
	if err != nil {
		if errors.Is(err, io.EOF) && !d.dupWtrClosed {
			d.dupWtr.Close()
			d.dupWtrClosed = true
		} else {
			slog.Warn("Error Reading to ORdr", "error", err)
		}
	}
	return
}
