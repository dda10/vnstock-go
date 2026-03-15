package exporter

import (
	"fmt"
	"io"
	"reflect"
	"time"

	vnstock "github.com/dda10/vnstock-go"
	"github.com/xuri/excelize/v2"
)

// ExportExcel writes records to w as an Excel (.xlsx) file with one worksheet.
// records must be a slice of structs. Returns *Error{Code: SerialiseError} on failure.
func ExportExcel(w io.Writer, records any) error {
	v := reflect.ValueOf(records)
	if v.Kind() != reflect.Slice {
		return &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: "records must be a slice",
		}
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	index, err := f.GetSheetIndex(sheetName)
	if err != nil {
		return &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: fmt.Sprintf("failed to get sheet index: %v", err),
			Cause:   err,
		}
	}
	if index == -1 {
		index, err = f.NewSheet(sheetName)
		if err != nil {
			return &vnstock.Error{
				Code:    vnstock.SerialiseError,
				Message: fmt.Sprintf("failed to create sheet: %v", err),
				Cause:   err,
			}
		}
	}
	f.SetActiveSheet(index)

	if v.Len() == 0 {
		// Write header only for empty slice
		if err := writeExcelHeader(f, sheetName, v.Type().Elem()); err != nil {
			return err
		}
	} else {
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

		for col, h := range header {
			cell, _ := excelize.CoordinatesToCellName(col+1, 1)
			if err := f.SetCellValue(sheetName, cell, h); err != nil {
				return &vnstock.Error{
					Code:    vnstock.SerialiseError,
					Message: fmt.Sprintf("failed to write header: %v", err),
					Cause:   err,
				}
			}
		}

		// Write data rows
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}

			col := 0
			for j := 0; j < elem.NumField(); j++ {
				field := elem.Type().Field(j)
				if !field.IsExported() {
					continue
				}

				fieldVal := elem.Field(j)
				cell, _ := excelize.CoordinatesToCellName(col+1, i+2)

				val := formatExcelField(fieldVal)
				if err := f.SetCellValue(sheetName, cell, val); err != nil {
					return &vnstock.Error{
						Code:    vnstock.SerialiseError,
						Message: fmt.Sprintf("failed to write cell %s: %v", cell, err),
						Cause:   err,
					}
				}
				col++
			}
		}
	}

	if err := f.Write(w); err != nil {
		return &vnstock.Error{
			Code:    vnstock.SerialiseError,
			Message: fmt.Sprintf("failed to write Excel file: %v", err),
			Cause:   err,
		}
	}

	return nil
}

func writeExcelHeader(f *excelize.File, sheetName string, elemType reflect.Type) error {
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

	for col, h := range header {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			return &vnstock.Error{
				Code:    vnstock.SerialiseError,
				Message: fmt.Sprintf("failed to write header: %v", err),
				Cause:   err,
			}
		}
	}

	return nil
}

func formatExcelField(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Bool:
		return v.Bool()
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
