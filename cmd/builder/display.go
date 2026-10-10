package main

import (
	"io"
	"strings"
)

type progressWriter struct {
	output  io.Writer
	pending string
}

func (w *progressWriter) Write(p []byte) (int, error) {
	const prefix = "[railpack]"
	text := w.pending + string(p)
	w.pending = ""
	// Retain only a possible prefix split across progress writes.
	for length := min(len(text), len(prefix)-1); length > 0; length-- {
		if strings.HasSuffix(text, prefix[:length]) {
			w.pending = text[len(text)-length:]
			text = text[:len(text)-length]
			break
		}
	}
	text = strings.ReplaceAll(text, prefix, "[deplopack]")
	n, err := io.WriteString(w.output, text)
	if err != nil {
		return 0, err
	}
	if n != len(text) {
		return 0, io.ErrShortWrite
	}
	return len(p), nil
}

func (w *progressWriter) Flush() error {
	_, err := io.WriteString(w.output, w.pending)
	w.pending = ""
	return err
}
