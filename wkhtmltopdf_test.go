package wkhtmltopdf

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWriteToFile(t *testing.T) {

	testcases := []struct {
		Case     string
		Pages    []string
		Filename string
		Err      string
	}{
		{"Simple", []string{"test_data/simple.html"}, "test_data/simple.pdf", ""},
		{"Missing", []string{"test_data/missing.html"}, "test_data/missing.pdf", "Error running wkhtmltopdf"},
		{"BadFile", []string{"test_data/simple.html"}, "<>/!//bad.pdf", "Error creating file"},
	}

	for _, tc := range testcases {

		doc := NewDocument()
		for _, pg := range tc.Pages {
			doc.AddPages(NewPage(pg))
		}
		err := doc.WriteToFile(tc.Filename, false)
		switch {
		case err == nil && tc.Err != "":
			t.Errorf("%v. Wrong error produced. Expected: %v, Got: %v", tc.Case, tc.Err, err)
		case err == nil:
			continue
		case !strings.HasPrefix(err.Error(), tc.Err):
			t.Errorf("%v. Wrong error produced. Expected: %v, Got: %v", tc.Case, tc.Err, err)
		}
	}
}

type BadWriter struct{}

func (w BadWriter) Write(p []byte) (int, error) {
	return 0, errors.New("Bad writer doesn't write")
}

func TestWriteToReader(t *testing.T) {

	testcases := []struct {
		Case   string
		Pages  []string
		Writer io.Writer
		Err    string
	}{
		{"Simple", []string{"test_data/simple.html"}, &bytes.Buffer{}, ""},
		{"Missing", []string{"test_data/missing.html"}, &bytes.Buffer{}, "Error running wkhtmltopdf"},
		{"Bad Writer", []string{"test_data/simple.html"}, BadWriter{}, "Error writing to writer"},
	}

	for _, tc := range testcases {

		doc := NewDocument()
		for _, pg := range tc.Pages {
			doc.AddPages(NewPage(pg))
		}
		err := doc.Write(tc.Writer, false)
		switch {
		case err == nil && tc.Err != "":
			t.Errorf("%v. Wrong error produced. Expected: %v, Got: %v", tc.Case, tc.Err, err)
		case err == nil:
			continue
		case !strings.HasPrefix(err.Error(), tc.Err):
			t.Errorf("%v. Wrong error produced. Expected: %v, Got: %v", tc.Case, tc.Err, err)
		}
	}
}

func TestWriteFromReader(t *testing.T) {

	f, err := os.Open("test_data/simple.html")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(f)
	f.Close()

	pg, err := NewPageReader(buf)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	doc := NewDocument()
	doc.AddPages(pg)

	output := &bytes.Buffer{}
	err = doc.Write(output, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestMultipleReaders(t *testing.T) {

	pages := []*Page{}
	for n := 0; n < 5; n++ {
		f, err := os.Open("test_data/simple.html")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		buf := &bytes.Buffer{}
		buf.ReadFrom(f)
		f.Close()

		pg, err := NewPageReader(buf)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		pages = append(pages, pg)
	}

	doc := NewDocument()
	doc.AddCover(pages[0])
	doc.AddPages(pages[1:]...)

	output := &bytes.Buffer{}
	err := doc.Write(output, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
