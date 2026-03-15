package exporter

import (
	"encoding/json"
	"fmt"
	"io"

	vnstock "github.com/dda10/vnstock-go"
)

// ExportJSON writes records to w as a pretty-printed JSON array.
// records must be a slice. Returns *Error{Code: SerialiseError} on failure.
func ExportJSON(w io.Writer, records any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(records); err != nil {
		return &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: fmt.Sprintf("failed to encode JSON: %v", err),
			Cause:   err,
		}
	}

	return nil
}
