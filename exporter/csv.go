package exporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

// ExportCSV writes records to w in CSV format with a header row.
// records must be a slice of structs. Returns *Error{Code: SerialiseError} on failure.
func ExportCSV(w io.Writer, records any) error {
	v := reflect.ValueOf(records)
	if v.Kind() != reflect.Slice {
		return &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "records must be a slice",
		}
	}

	if v.Len() == 0 {
		// Write header only for empty slice
		return writeCSVHeader(w, v.Type().Elem())
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	header := make([]string, 0, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if field.IsExported() {
			header = append(header, field.Name)
		}
	}

	if err := writer.Write(header); err != nil {
		return &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: fmt.Sprintf("failed to write CSV header: %v", err),
			Cause:   err,
		}
	}

	// Write data rows
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		row := make([]string, 0, len(header))
		for j := 0; j < elem.NumField(); j++ {
			field := elem.Type().Field(j)
			if !field.IsExported() {
				continue
			}

			fieldVal := elem.Field(j)
			row = append(row, formatField(fieldVal))
		}

		if err := writer.Write(row); err != nil {
			return &vnstock.Error{
				Code:    vnstock.SerialiseError,
				Message: fmt.Sprintf("failed to write CSV row %d: %v", i, err),
				Cause:   err,
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: fmt.Sprintf("CSV writer error: %v", err),
			Cause:   err,
		}
	}

	return nil
}

func writeCSVHeader(w io.Writer, elemType reflect.Type) error {
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := make([]string, 0, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if field.IsExported() {
			header = append(header, field.Name)
		}
	}

	if err := writer.Write(header); err != nil {
		return &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: fmt.Sprintf("failed to write CSV header: %v", err),
			Cause:   err,
		}
	}

	return nil
}

func formatField(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return t.Format(time.RFC3339)
		}
		return fmt.Sprintf("%v", v.Interface())
	case reflect.Map:
		return fmt.Sprintf("%v", v.Interface())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
