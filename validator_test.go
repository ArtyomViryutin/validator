package validator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		checkErr func(err error) bool
	}{
		{
			name: "invalid struct: interface",
			args: args{
				v: new(any),
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, ErrNotStruct)
			},
		},
		{
			name: "invalid struct: map",
			args: args{
				v: map[string]string{},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, ErrNotStruct)
			},
		},
		{
			name: "invalid struct: string",
			args: args{
				v: "some string",
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, ErrNotStruct)
			},
		},
		{
			name: "valid struct with no fields",
			args: args{
				v: struct{}{},
			},
			wantErr: false,
		},
		{
			name: "valid struct with untagged fields",
			args: args{
				v: struct {
					f1 string
					f2 string
				}{},
			},
			wantErr: false,
		},
		{
			name: "valid struct with unexported fields",
			args: args{
				v: struct {
					foo string `validate:"len:10"`
				}{},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				e := &ValidationErrors{}
				return errors.As(err, e) && e.Error() == ErrValidateForUnexportedFields.Error()
			},
		},
		{
			name: "invalid validator syntax",
			args: args{
				v: struct {
					Foo string `validate:"len:abcdef"`
				}{},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				e := &ValidationErrors{}
				return errors.As(err, e) && e.Error() == ErrInvalidValidatorSyntax.Error()
			},
		},
		{
			name: "valid struct with tagged fields",
			args: args{
				v: struct {
					Len       string `validate:"len:20"`
					LenZ      string `validate:"len:0"`
					InInt     int    `validate:"in:20,25,30"`
					InNeg     int    `validate:"in:-20,-25,-30"`
					InStr     string `validate:"in:foo,bar"`
					MinInt    int    `validate:"min:10"`
					MinIntNeg int    `validate:"min:-10"`
					MinStr    string `validate:"min:10"`
					MinStrNeg string `validate:"min:-1"`
					MaxInt    int    `validate:"max:20"`
					MaxIntNeg int    `validate:"max:-2"`
					MaxStr    string `validate:"max:20"`
				}{
					Len:       "abcdefghjklmopqrstvu",
					LenZ:      "",
					InInt:     25,
					InNeg:     -25,
					InStr:     "bar",
					MinInt:    15,
					MinIntNeg: -9,
					MinStr:    "abcdefghjkl",
					MinStrNeg: "abc",
					MaxInt:    16,
					MaxIntNeg: -3,
					MaxStr:    "abcdefghjklmopqrst",
				},
			},
			wantErr: false,
		},
		{
			name: "wrong length",
			args: args{
				v: struct {
					Lower    string `validate:"len:24"`
					Higher   string `validate:"len:5"`
					Zero     string `validate:"len:3"`
					BadSpec  string `validate:"len:%12"`
					Negative string `validate:"len:-6"`
				}{
					Lower:    "abcdef",
					Higher:   "abcdef",
					Zero:     "",
					BadSpec:  "abc",
					Negative: "abcd",
				},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				assert.Len(t, err.(ValidationErrors), 5)
				return true
			},
		},
		{
			name: "wrong in",
			args: args{
				v: struct {
					InA     string `validate:"in:ab,cd"`
					InB     string `validate:"in:aa,bb,cd,ee"`
					InC     int    `validate:"in:-1,-3,5,7"`
					InD     int    `validate:"in:5-"`
					InEmpty string `validate:"in:"`
				}{
					InA:     "ef",
					InB:     "ab",
					InC:     2,
					InD:     12,
					InEmpty: "",
				},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				assert.Len(t, err.(ValidationErrors), 5)
				return true
			},
		},
		{
			name: "wrong min",
			args: args{
				v: struct {
					MinA string `validate:"min:12"`
					MinB int    `validate:"min:-12"`
					MinC int    `validate:"min:5-"`
					MinD int    `validate:"min:"`
					MinE string `validate:"min:"`
				}{
					MinA: "ef",
					MinB: -22,
					MinC: 12,
					MinD: 11,
					MinE: "abc",
				},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				assert.Len(t, err.(ValidationErrors), 5)
				return true
			},
		},
		{
			name: "wrong max",
			args: args{
				v: struct {
					MaxA string `validate:"max:2"`
					MaxB string `validate:"max:-7"`
					MaxC int    `validate:"max:-12"`
					MaxD int    `validate:"max:5-"`
					MaxE int    `validate:"max:"`
					MaxF string `validate:"max:"`
				}{
					MaxA: "efgh",
					MaxB: "ab",
					MaxC: 22,
					MaxD: 12,
					MaxE: 11,
					MaxF: "abc",
				},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				assert.Len(t, err.(ValidationErrors), 6)
				return true
			},
		},
		{
			name: "valid struct with multiple tags",
			args: args{
				v: struct {
					MinMaxInt int    `validate:"min:0;max:10"`
					MinInInt  int    `validate:"min:0;in:2,4,6"`
					MaxInInt  int    `validate:"max:10;in:2,4,6"`
					LenInStr  string `validate:"len:1;in:a,b,c"`
					MinMaxStr string `validate:"min:0;max:10"`
					MinInStr  string `validate:"min:0;in:a,b,c"`
					MaxInStr  string `validate:"max:10;in:a,b,c"`
				}{
					MinMaxInt: 5,
					MinInInt:  4,
					MaxInInt:  4,
					MinMaxStr: "abc",
					MinInStr:  "b",
					MaxInStr:  "b",
					LenInStr:  "b",
				},
			},
			wantErr: false,
		},
		{
			name: "single wrong tag",
			args: args{
				v: struct {
					MinInt int    `validate:"min:5;max:10;in:2,4,12"`
					MaxInt int    `validate:"min:5;max:10;in:2,4,12"`
					InInt  int    `validate:"min:5;max:10;in:2,4,12"`
					LenStr string `validate:"len:2;in:a,b,c"`
					MinStr string `validate:"min:2;max:10;in:a,b,c"`
					MaxStr string `validate:"min:0;max:1;in:aa,bb,cc"`
					InStr  string `validate:"min:1;max:10;in:a,b,c"`
				}{
					MinInt: 2,
					MaxInt: 12,
					InInt:  4,
					LenStr: "a",
					MinStr: "a",
					MaxStr: "bb",
					InStr:  "gg",
				},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				assert.Len(t, err.(ValidationErrors), 7)
				return true
			},
		},
		{
			name: "multiple wrong tags",
			args: args{
				v: struct {
					MinInInt int    `validate:"min:5;max:10;in:2,4,12"`
					MaxInInt int    `validate:"min:5;max:10;in:2,4,12"`
					LenInStr string `validate:"len:2;in:a,b,c"`
					MinInStr string `validate:"min:2;max:10;in:a,b,c"`
					MaxInStr string `validate:"min:0;max:1;in:aa,bb,cc"`
				}{
					MinInInt: 1,
					MaxInInt: 14,
					LenInStr: "g",
					MinInStr: "g",
					MaxInStr: "gg",
				},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				assert.Len(t, err.(ValidationErrors), 10)
				return true
			},
		},
		{
			name: "valid struct with nested",
			args: args{
				v: struct {
					A      int    `validate:"min:0;max:5"`
					B      string `validate:"len:2"`
					Nested struct {
						A int    `validate:"min:0;max:5"`
						B string `validate:"len:2"`
					}
				}{
					A: 2,
					B: "bb",
					Nested: struct {
						A int    `validate:"min:0;max:5"`
						B string `validate:"len:2"`
					}{
						A: 2,
						B: "bb",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "struct with wrong nested",
			args: args{
				v: struct {
					A      int    `validate:"min:0;max:5"`
					B      string `validate:"len:2"`
					Nested struct {
						A int    `validate:"min:0;max:5"`
						B string `validate:"len:2"`
					}
				}{
					A: 2,
					B: "bb",
					Nested: struct {
						A int    `validate:"min:0;max:5"`
						B string `validate:"len:2"`
					}{
						A: 10,
						B: "bbb",
					},
				},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				assert.Len(t, err.(ValidationErrors), 2)
				return true
			},
		},
		{
			name: "struct with double nested",
			args: args{
				v: struct {
					OuterNested struct {
						InnerNested struct {
							A int `validate:"min:-1;max:1"`
						}
					}
				}{},
			},
			wantErr: false,
		},
		{
			name: "struct with wrong double nested",
			args: args{
				v: struct {
					OuterNested struct {
						InnerNested struct {
							A int `validate:"min:1;max:2"`
						}
					}
				}{},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				assert.Len(t, err.(ValidationErrors), 1)
				return true
			},
		},
		{
			name: "struct with unexported nested",
			args: args{
				v: struct {
					nested struct {
						A int `validate:"min:0;max:5"`
					}
				}{},
			},
			wantErr: true,
			checkErr: func(err error) bool {
				assert.Len(t, err.(ValidationErrors), 1)
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.args.v)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, tt.checkErr(err), "test expect an error, but got wrong error type")
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
