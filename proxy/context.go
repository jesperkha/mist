package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Context struct {
	r *http.Request
	w http.ResponseWriter
}

// JSON marshals and writes json data to the response writer.
func (c *Context) JSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.w.Write(b)
	return nil
}

// String is a shorthand for writing bytes to the response writer.
func (c *Context) String(s string) {
	c.w.Write([]byte(s))
}

// FileDownload writes the file from reader to the response writer and marks
// it as an attachement (downloaded file).
func (c *Context) FileDownload(filename string, r io.Reader) error {
	c.w.Header().Set("Content-Disposition", fmt.Sprintf("attachement;filename=%s", filename))
	_, err := io.Copy(c.w, r)
	return err
}

// Serve static file.
func (c *Context) File(filename string) {
	http.ServeFile(c.w, c.r, filename)
}
