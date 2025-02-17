// © 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
// ISC License

package testpackage

import (
	"bufio"
	"errors"
	"iter"
	"os"
)

// LineReader provides an iterator reading a file line-by-line
type LineReader struct {
	// the file lines are being read from
	filename string
	// a pointer to store occurring errors
	errp *error
	// the open file
	osFile *os.File
}

// NewLineReader returns an iterator over the lines of a file
//   - [LineReader.Lines] is iterator function
func NewLineReader(fieldp *LineReader, filename string, errp *error) (lineReader *LineReader) {
	if fieldp != nil {
		lineReader = fieldp
	} else {
		lineReader = &LineReader{}
	}
	lineReader.filename = filename
	lineReader.errp = errp

	return
}

// Lines is a single-value string iterator
func (r *LineReader) Lines(yield func(line string) (keepGoing bool)) {
	var err error
	defer r.cleanup(&err)

	if r.osFile, err = os.Open(r.filename); err != nil {
		return // i/o error
	}
	var scanner = bufio.NewScanner(r.osFile)
	for scanner.Scan() {
		if !yield(scanner.Text()) {
			return // iteration canceled by break or such
		}
	}
	// reached end of file
}

// LineReader.Lines is iter.Seq
var _ iter.Seq[string] = (&LineReader{}).Lines

// cleanup is invoked on iteration end or any panic
//   - errp: possible error from Lines
func (r *LineReader) cleanup(errp *error) {
	var err error
	if r.osFile != nil {
		err = r.osFile.Close()
	}
	if err != nil || *errp != nil {
		// aggregate errors in order of occurrence
		*r.errp = errors.Join(*r.errp, *errp, err)
	}
}
