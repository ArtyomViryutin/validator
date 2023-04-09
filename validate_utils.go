package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func contains[T comparable](s []T, needle T) bool {
	for _, v := range s {
		if v == needle {
			return true
		}
	}
	return false
}

func parseIntSlice(s string) ([]int, error) {
	strVals := strings.Split(s, ",")
	intVals := make([]int, 0, len(strVals))
	for _, v := range strVals {
		val, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		intVals = append(intVals, val)
	}
	return intVals, nil
}

type fieldValidator interface {
	validate(reflect.Value) error
}

type fieldValidatorCreator func(string) (fieldValidator, error)

type intMinValidator struct {
	min int
}

func (v intMinValidator) validate(i reflect.Value) error {
	val := int(i.Int())
	if val < v.min {
		return fmt.Errorf("%d is less than min allowed %d", val, v.min)
	}
	return nil
}

func newIntMinValidator(s string) (fieldValidator, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return intMinValidator{val}, nil
}

type intMaxValidator struct {
	max int
}

func (v intMaxValidator) validate(i reflect.Value) error {
	val := int(i.Int())
	if val > v.max {
		return fmt.Errorf("%d is higher than max allowed %d", val, v.max)
	}
	return nil
}

func newIntMaxValidator(s string) (fieldValidator, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return intMaxValidator{val}, nil
}

type intInValidator struct {
	in []int
}

func (v intInValidator) validate(i reflect.Value) error {
	val := int(i.Int())
	if !contains(v.in, val) {
		return fmt.Errorf("%d is not in %v", val, v.in)
	}
	return nil
}

func newIntInValidator(s string) (fieldValidator, error) {
	vals, err := parseIntSlice(s)
	if err != nil {
		return nil, err
	}
	return intInValidator{vals}, nil
}

type strLenValidator struct {
	l int
}

func (v strLenValidator) validate(s reflect.Value) error {
	val := s.String()
	if len(val) != v.l {
		return fmt.Errorf("len of %s is not equal to %d", val, v.l)
	}
	return nil
}

func newStrLenValidator(s string) (fieldValidator, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return strLenValidator{val}, nil
}

type strMinValidator struct {
	min int
}

func (v strMinValidator) validate(s reflect.Value) error {
	val := s.String()
	if len(val) < v.min {
		return fmt.Errorf("len of %s is less than min allowed %d", val, v.min)
	}
	return nil
}

func newStrMinValidator(s string) (fieldValidator, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return strMinValidator{val}, nil
}

type strMaxValidator struct {
	max int
}

func (v strMaxValidator) validate(s reflect.Value) error {
	val := s.String()
	if len(val) > v.max {
		return fmt.Errorf("len of %s is higher than max allowed %d", val, v.max)
	}
	return nil
}

func newStrMaxValidator(s string) (fieldValidator, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return strMaxValidator{val}, nil
}

type strInValidator struct {
	in []string
}

func (v strInValidator) validate(s reflect.Value) error {
	val := s.String()
	if !contains(v.in, val) {
		return fmt.Errorf("%s is not in %v", val, v.in)
	}
	return nil
}

func newStrInValidator(s string) (fieldValidator, error) {
	return strInValidator{strings.Split(s, ",")}, nil
}
