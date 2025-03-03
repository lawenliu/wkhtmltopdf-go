package wkhtmltopdf

/* This example creates an http server, which returns a simple
   pdf document with a title and the path of the request.
*/

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

const page = `
<html>
  <body>
    <h1>Test Page</h1>

	<p>Path: {{.}}</p>
  </body>
</html>`

func handler(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.New("page").Parse(page))
	buf := &bytes.Buffer{}
	tmpl.Execute(buf, r.URL.String())

	doc := NewDocument()
	pg, err := NewPageReader(buf)
	if err != nil {
		log.Fatal("Error reading page buffer")
	}
	doc.AddPages(pg)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="test.pdf"`)
	err = doc.Write(w, false)
	if err != nil {
		log.Fatal("Error serving pdf")
	}
}

func Example() {

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)

}
