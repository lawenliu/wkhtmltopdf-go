package wkhtmltopdf

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

// A Document represents a single pdf document.
type Document struct {
	pages   []*Page
	options []string

	tmp string // temp directory
}

// NewDocument creates a new document.
func NewDocument(opts ...Option) *Document {

	doc := &Document{pages: []*Page{}, options: []string{}}
	doc.AddOptions(opts...)
	return doc
}

// AddPages to the document. Pages will be included in
// the final pdf in the order they are added.
func (doc *Document) AddPages(pages ...*Page) {
	doc.pages = append(doc.pages, pages...)
}

// AddCover adds a cover page to the document.
func (doc *Document) AddCover(cover *Page) {
	doc.pages = append(doc.pages, cover)
	cover.cover = true
}

// AddOptions allows the setting of options after document creation.
func (doc *Document) AddOptions(opts ...Option) {

	for _, opt := range opts {
		doc.options = append(doc.options, opt.opts()...)
	}
}

// args calculates the args needed to run wkhtmltopdf
func (doc *Document) args() []string {

	args := []string{}
	args = append(args, doc.options...)

	// pages
	for _, pg := range doc.pages {
		args = append(args, pg.args()...)
	}

	return args
}

// readers counts the number of pages using a reader
// as a source
func (doc *Document) readers() int {

	n := 0
	for _, pg := range doc.pages {
		if pg.reader {
			n++
		}
	}
	return n
}

// writeTempPages writes the pages generated by a reader
// to a set of pages within a temp directory.
func (doc *Document) writeTempPages() error {

	var err error
	doc.tmp, err = ioutil.TempDir(TempDir, "temp")
	if err != nil {
		return fmt.Errorf("Error creating temp directory")
	}

	for n, pg := range doc.pages {
		if !pg.reader {
			continue
		}

		pg.filename = fmt.Sprintf("%v/%v/page%08d.html", TempDir, doc.tmp, n)
		err := ioutil.WriteFile(pg.filename, pg.buf.Bytes(), 0666)
		if err != nil {
			return fmt.Errorf("Error writing temp file: %v", err)
		}
	}

	return nil
}

// createPDF creates the pdf and writes it to the buffer,
// which can then be written to file or writer.
func (doc *Document) createPDF() (*bytes.Buffer, error) {

	var stdin io.Reader
	switch {
	case doc.readers() == 1:

		// Pipe through stdin for a single reader.
		for _, pg := range doc.pages {
			if pg.reader {
				stdin = pg.buf
				pg.filename = "-"
				break
			}
		}

	case doc.readers() > 1:

		// Write multiple readers to temp files
		err := doc.writeTempPages()
		if err != nil {
			return nil, fmt.Errorf("Error writing temp files: %v", err)
		}
	}

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(stdin)
	buf, err := CreatePDFNormal(buffer)
	if err != nil {
		buf, err = CreatePDFXvfb(buffer)
		if err != nil {
			return nil, err
		}
	}

	if doc.tmp != "" {
		err = os.RemoveAll(TempDir + "/" + doc.tmp)
	}
	return buf, err

}

// CreatePDFXvfb create with xvfb to start wkhtmltopdf to avoid headless (without display)
func (doc *Document) CreatePDFXvfb(buffer *bytes.Buffer) (*bytes.Buffer, error) {
	stdin := bytes.NewReader(buffer.Bytes())
	args := append(doc.args(), "-")
	buf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}
	args = append([]string {Executable_Wkhtmltopdf}, args...)
	cmd := exec.Command(Executable_Xvfb, args...)
	cmd.Stdin = stdin
	cmd.Stdout = buf
	cmd.Stderr = errbuf

	err := cmd.Run()
	if err != nil {
        	return nil, fmt.Errorf("First: Error running wkhtmltopdf: %v", errbuf.String())
	}
	
	return buf, nil
}

// CreatePDFNormal create with wkhtmltopdf directly
func (doc *Document) CreatePDFNormal(buffer *bytes.Buffer) (*bytes.Buffer, error) {
	stdin := bytes.NewReader(buffer.Bytes())
	args := append(doc.args(), "-")
	buf := &bytes.Buffer{}
	errbuf := &bytes.Buffer{}
	cmd := exec.Command(Executable_Wkhtmltopdf, args...)
	cmd.Stdin = stdin
	cmd.Stdout = buf
	cmd.Stderr = errbuf

	err := cmd.Run()
	if err != nil {
        	return nil, fmt.Errorf("First: Error running wkhtmltopdf: %v", errbuf.String())
	}
	
	return buf, nil
}

// WriteToFile creates the pdf document and writes it
// to the specified filename.
func (doc *Document) WriteToFile(filename string) error {

	buf, err := doc.createPDF()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, buf.Bytes(), 0666)
	if err != nil {
		return fmt.Errorf("Error creating file: %v", err)
	}

	return nil
}

// Write creates the pdf document and writes it
// to the provided reader.
func (doc *Document) Write(w io.Writer) error {

	buf, err := doc.createPDF()
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("Error writing to writer: %v", err)
	}

	return nil
}
