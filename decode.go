package gocsv

import (
	"fmt"
	"io"
	"reflect"
)

type decoder struct {
	in io.Reader
}

func newDecoder(in io.Reader) *decoder {
	return &decoder{in}
}

func (decode *decoder) getCSVRows() ([][]string, error) {
	return getCSVReader(decode.in).ReadAll()
}

func (decode *decoder) readTo(out interface{}) error {
	outValue, outType := getConcreteReflectValueAndType(out) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
	if err := ensureOutType(outType); err != nil {
		return err
	}
	outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
	if err := ensureOutInnerType(outInnerType); err != nil {
		return err
	}
	csvRows, err := decode.getCSVRows() // Get the CSV csvRows
	if err != nil {
		return err
	}
	if err := ensureOutCapacity(&outValue, len(csvRows)); err != nil { // Ensure the container is big enough to hold the CSV content
		return err
	}
	outInnerStructInfo := getStructInfo(outInnerType)                            // Get the inner struct info to get CSV annotations
	csvHeadersLabels := make(map[int]*fieldInfo, len(outInnerStructInfo.Fields)) // Used to store the correspondance header <-> position in CSV
	for i, csvRow := range csvRows {                                             // Iterate over csv rows
		if i == 0 { // First line of CSV is the header line
			for j, csvColumnHeader := range csvRow {
				if fieldInfo := getCSVFieldPosition(csvColumnHeader, outInnerStructInfo); fieldInfo != nil {
					csvHeadersLabels[j] = fieldInfo
				}
			}
		} else {
			if len(outInnerStructInfo.Fields) != len(csvHeadersLabels) {
				return fmt.Errorf("not all struct tags were matched to a csv header")
			}
			outInner := createNewOutInner(outInnerWasPointer, outInnerType)
			for j, csvColumnContent := range csvRow {
				if fieldInfo, ok := csvHeadersLabels[j]; ok { // Position found accordingly to header name
					if err := setInnerField(&outInner, outInnerWasPointer, fieldInfo.Num, csvColumnContent); err != nil { // Set field of struct
						return err
					}
				}
			}
			outValue.Index(i - 1).Set(outInner)
		}
	}
	return nil
}

// Check if the outType is an array or a slice
func ensureOutType(outType reflect.Type) error {
	switch outType.Kind() {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		return nil
	}
	return fmt.Errorf("cannot use " + outType.String() + ", only slice or array supported")
}

// Check if the outInnerType is of type struct
func ensureOutInnerType(outInnerType reflect.Type) error {
	switch outInnerType.Kind() {
	case reflect.Struct:
		return nil
	}
	return fmt.Errorf("cannot use " + outInnerType.String() + ", only struct supported")
}

func ensureOutCapacity(out *reflect.Value, csvLen int) error {
	switch out.Kind() {
	case reflect.Array:
		if out.Len() < csvLen-1 { // Array is not big enough to hold the CSV content (arrays are not addressable)
			return fmt.Errorf("array capacity problem: cannot store %d %s in %s", csvLen-1, out.Type().Elem().String(), out.Type().String())
		}
	case reflect.Slice:
		if !out.CanAddr() && out.Len() < csvLen-1 { // Slice is not big enough tho hold the CSV content and is not addressable
			return fmt.Errorf("slice capacity problem and is not addressable (did you forget &?)")
		} else if out.CanAddr() && out.Len() < csvLen-1 {
			out.Set(reflect.MakeSlice(out.Type(), csvLen-1, csvLen-1)) // Slice is not big enough, so grows it
		}
	}
	return nil
}

func getCSVFieldPosition(key string, structInfo *structInfo) *fieldInfo {
	for _, field := range structInfo.Fields {
		if field.Key == key {
			return &field
		}
	}
	return nil
}

func createNewOutInner(outInnerWasPointer bool, outInnerType reflect.Type) reflect.Value {
	if outInnerWasPointer {
		return reflect.New(outInnerType)
	}
	return reflect.New(outInnerType).Elem()
}

func setInnerField(outInner *reflect.Value, outInnerWasPointer bool, fieldPosition int, value string) error {
	if outInnerWasPointer {
		return setField(outInner.Elem().Field(fieldPosition), value)
	}
	return setField(outInner.Field(fieldPosition), value)
}
