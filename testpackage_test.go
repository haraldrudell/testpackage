package testpackage_test

import (
	"bufio"
	"errors"
	"fmt"
	"iter"
	"os"
)

// main_8_lineReader demonstrates an iterator that
//   - can handle internal failure, for block panic and
//   - provide error value to outside the for statement
func Example() {

	// create test file
	var filename = "test.txt"
	if err := os.WriteFile(filename, []byte("one\ntwo\n"), 0600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := iterateLines(filename); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// iterateLines demonstrates the four needs of an iterator:
//   - 1 receive and maintain internal state: filename, errp, osFile
//   - 2 provide iteration values and determine end of iteration: [LineReader.Lines]
//   - 3 release resources upon end of iteration: [LineReader.cleanup]
//   - 4 propagate error outside the for statement: errp
//
// the iterator LineReader is allocated on the stack
//   - stack allocation is faster than heap allocation
//   - on stack even if NewLineReader is in another module
//   - LineReader pointer receiver is more performant
func iterateLines(filename string) (err error) {
	for line := range NewLineReader(&LineReader{}, filename, &err).Lines {
		println("iterator line:", line)
	}

	return
}

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
