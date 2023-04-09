package validator

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"reflect"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

type ValidationError struct {
	Field string
	Err   error
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Err)
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 1 {
		return v[0].Err.Error()
	}
	s := make([]string, 0, len(v))
	for _, err := range v {
		s = append(s, err.Error())
	}
	return strings.Join(s, "\n")
}

var intValidators = map[string]fieldValidatorCreator{
	"min": newIntMinValidator,
	"max": newIntMaxValidator,
	"in":  newIntInValidator,
}

var strValidators = map[string]fieldValidatorCreator{
	"len": newStrLenValidator,
	"min": newStrMinValidator,
	"max": newStrMaxValidator,
	"in":  newStrInValidator,
}

func parseValidators(t reflect.Type, tag string) ([]fieldValidator, error) {
	var fieldValidators map[string]fieldValidatorCreator
	switch t.Kind() {
	case reflect.Int:
		fieldValidators = intValidators
	case reflect.String:
		fieldValidators = strValidators
	default:
		log.Panicf("unsupported type: %s", t.String())
	}

	kvs := strings.Split(tag, ";")
	validators := make([]fieldValidator, 0, len(kvs))
	for _, kv := range kvs {
		vals := strings.Split(kv, ":")
		if len(vals) != 2 || len(vals[0]) == 0 || len(vals[1]) == 0 {
			return nil, ErrInvalidValidatorSyntax
		}
		k, v := vals[0], vals[1]
		createValidator, ok := fieldValidators[k]
		if !ok {
			return nil, ErrInvalidValidatorSyntax
		}
		validator, err := createValidator(v)
		if err != nil {
			return nil, ErrInvalidValidatorSyntax
		}
		validators = append(validators, validator)
	}
	return validators, nil
}

func needValidation(f reflect.StructField) bool {
	switch f.Type.Kind() {
	case reflect.Int, reflect.String:
		if _, ok := f.Tag.Lookup("validate"); ok {
			return true
		}
	case reflect.Struct:
		for i := 0; i < f.Type.NumField(); i++ {
			if needValidation(f.Type.Field(i)) {
				return true
			}
		}
	}
	return false
}

func Validate(v any) error {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	vv := reflect.ValueOf(v)
	errs := make(ValidationErrors, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !needValidation(f) {
			continue
		}
		if !f.IsExported() {
			errs = append(errs, ValidationError{f.Name, ErrValidateForUnexportedFields})
			continue
		}
		if f.Type.Kind() == reflect.Struct {
			nestedErrs := Validate(vv.Field(i).Interface())
			if nestedErrs == nil {
				continue
			}
			for _, err := range nestedErrs.(ValidationErrors) {
				errs = append(errs, ValidationError{f.Name, err})
			}
			continue
		}
		validators, err := parseValidators(f.Type, f.Tag.Get("validate"))
		if err != nil {
			errs = append(errs, ValidationError{f.Name, err})
			continue
		}
		fv := vv.Field(i)
		for _, validator := range validators {
			err = validator.validate(fv)
			if err != nil {
				errs = append(errs, ValidationError{f.Name, err})
			}
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}
