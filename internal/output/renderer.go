package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Renderer struct {
	out io.Writer
}

func NewRenderer(out io.Writer) *Renderer {
	return &Renderer{out: out}
}

func (r *Renderer) WriteJSON(value any) error {
	encoder := json.NewEncoder(r.out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func (r *Renderer) WriteTable(headers []string, rows [][]string) error {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i := range headers {
			if i < len(row) && len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}

	if err := r.writeRow(headers, widths); err != nil {
		return err
	}
	for _, row := range rows {
		if err := r.writeRow(row, widths); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) writeRow(values []string, widths []int) error {
	for i := range widths {
		value := ""
		if i < len(values) {
			value = values[i]
		}
		if i == len(widths)-1 {
			if _, err := fmt.Fprint(r.out, value); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprint(r.out, value); err != nil {
			return err
		}
		if _, err := fmt.Fprint(r.out, strings.Repeat(" ", widths[i]-len(value)+2)); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(r.out)
	return err
}
