package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Context struct {
	R *http.Request
	W http.ResponseWriter
}

// JSON marshals and writes json data to the response writer.
func (c *Context) JSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.W.Write(b)
	return nil
}

// String is a shorthand for writing bytes to the response writer.
func (c *Context) String(s string) {
	c.W.Write([]byte(s))
}

// FileDownload writes the file from reader to the response writer and marks
// it as an attachement (downloaded file).
func (c *Context) FileDownload(filename string, r io.Reader) error {
	c.W.Header().Set("Content-Disposition", fmt.Sprintf("attachement;filename=%s", filename))
	_, err := io.Copy(c.W, r)
	return err
}

// Serve static file.
func (c *Context) File(filename string) {
	http.ServeFile(c.W, c.R, filename)
}
