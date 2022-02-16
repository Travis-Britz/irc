/*
Package ircdebug contains helper functions that are useful while writing an IRC client.
*/
package ircdebug

import (
	"io"
)

// WriteTo returns a new io.ReadWriteCloser that copies all reads/writes for rwc to w.
// Reads and Writes are prefixed with inPrefix and outPrefix respectively.
// This is mainly useful while developing an IRC client like a bot,
// e.g. for writing to os.Stdout or a file.
// todo: it's not safe for concurrent usage, so replies are sometimes mixed in with connection reads
func WriteTo(w io.Writer, rwc io.ReadWriteCloser, outPrefix string, inPrefix string) io.ReadWriteCloser {
	return &debugConn{
		ReadWriteCloser: rwc,
		r:               io.TeeReader(rwc, &writePrefixer{w: w, prefix: inPrefix}),
		w:               io.MultiWriter(rwc, &writePrefixer{w: w, prefix: outPrefix}),
	}
}

type debugConn struct {
	io.ReadWriteCloser
	r io.Reader
	w io.Writer
}

func (dc *debugConn) Read(p []byte) (int, error) {
	return dc.r.Read(p)
}
func (dc *debugConn) Write(p []byte) (int, error) {
	return dc.w.Write(p)
}

type writePrefixer struct {
	w      io.Writer
	prefix string
}

func (wp *writePrefixer) Write(p []byte) (n int, err error) {
	n, err = wp.w.Write(append([]byte(wp.prefix), p...))

	// since this writePrefixer is only ever used for a MultiWriter, we need to lie about how many bytes
	// were written so that the MultiWriter doesn't have an error for different byte counts on each of its writers.
	return n - len(wp.prefix), err
}
