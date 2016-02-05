package echo

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type (
	// Binder is the interface that wraps the Bind method.
	Binder interface {
		Bind(*http.Request, interface{}) error
	}

	binder struct {
		maxMemory int64
	}
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

// SetMaxBodySize sets multipart forms max body size
func (b *binder) SetMaxMemory(size int64) {
	b.maxMemory = size
}

// MaxBodySize return multipart forms max body size
func (b *binder) MaxMemory() int64 {
	return b.maxMemory
}

func (b *binder) Bind(r *http.Request, i interface{}) (err error) {
	if r.Body == nil {
		err = NewHTTPError(http.StatusBadRequest, "Request body can't be nil")
		return
	}
	defer r.Body.Close()
	ct := r.Header.Get(ContentType)
	err = ErrUnsupportedMediaType
	switch {
	case strings.HasPrefix(ct, ApplicationJSON):
		if err = json.NewDecoder(r.Body).Decode(i); err != nil {
			err = NewHTTPError(http.StatusBadRequest, err.Error())
		}
	case strings.HasPrefix(ct, ApplicationXML):
		if err = xml.NewDecoder(r.Body).Decode(i); err != nil {
			err = NewHTTPError(http.StatusBadRequest, err.Error())
		}
	case strings.HasPrefix(ct, ApplicationForm):
		if err = b.bindForm(r, i); err != nil {
			err = NewHTTPError(http.StatusBadRequest, err.Error())
		}
	case strings.HasPrefix(ct, MultipartForm):
		if err = b.bindMultiPartForm(r, i); err != nil {
			err = NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}
	return
}

func (binder) bindForm(r *http.Request, i interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return mapForm(i, r.Form)
}

func (b binder) bindMultiPartForm(r *http.Request, i interface{}) error {
	if b.maxMemory == 0 {
		b.maxMemory = defaultMaxMemory
	}
	if err := r.ParseMultipartForm(b.maxMemory); err != nil {
		return err
	}
	return mapForm(i, r.Form)
}

func mapForm(ptr interface{}, form map[string][]string) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}

		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get("form")
		if inputFieldName == "" {
			inputFieldName = typeField.Name

			// if "form" tag is nil, we inspect if the field is a struct.
			// this would not make sense for JSON parsing but it does for a form
			// since data is flatten
			if structFieldKind == reflect.Struct {
				err := mapForm(structField.Addr().Interface(), form)
				if err != nil {
					return err
				}
				continue
			}
		}
		inputValue, exists := form[inputFieldName]
		if !exists {
			continue
		}

		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for i := 0; i < numElems; i++ {
				if err := setWithProperType(sliceOf, inputValue[i], slice.Index(i)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else {
			if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
				return err
			}
		}
	}
	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("Unknown type")
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
