// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewId(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := NewId()
		if len(id) > 26 {
			t.Fatal("ids shouldn't be longer than 26 chars")
		}
	}
}

func TestRandomString(t *testing.T) {
	for i := 0; i < 1000; i++ {
		r := NewRandomString(32)
		if len(r) != 32 {
			t.Fatal("should be 32 chars")
		}
	}
}

func TestGetMillisForTime(t *testing.T) {
	thisTimeMillis := int64(1471219200000)
	thisTime := time.Date(2016, time.August, 15, 0, 0, 0, 0, time.UTC)

	result := GetMillisForTime(thisTime)

	if thisTimeMillis != result {
		t.Fatalf(fmt.Sprintf("millis are not the same: %d and %d", thisTimeMillis, result))
	}
}

func TestPadDateStringZeros(t *testing.T) {
	for _, testCase := range []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "Valid date",
			Input:    "2016-08-01",
			Expected: "2016-08-01",
		},
		{
			Name:     "Valid date but requires padding of zero",
			Input:    "2016-8-1",
			Expected: "2016-08-01",
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			assert.Equal(t, testCase.Expected, PadDateStringZeros(testCase.Input))
		})
	}
}

func TestAppError(t *testing.T) {
	err := NewAppError("TestAppError", "message", nil, "", http.StatusInternalServerError)
	json := err.ToJson()
	rerr := AppErrorFromJson(strings.NewReader(json))
	require.Equal(t, err.Message, rerr.Message)

	t.Log(err.Error())
}

func TestAppErrorJunk(t *testing.T) {
	rerr := AppErrorFromJson(strings.NewReader("<html><body>This is a broken test</body></html>"))
	require.Equal(t, "body: <html><body>This is a broken test</body></html>", rerr.DetailedError)
}

func TestCopyStringMap(t *testing.T) {
	itemKey := "item1"
	originalMap := make(map[string]string)
	originalMap[itemKey] = "val1"

	copyMap := CopyStringMap(originalMap)
	copyMap[itemKey] = "changed"

	assert.Equal(t, "val1", originalMap[itemKey])
}

func TestMapJson(t *testing.T) {

	m := make(map[string]string)
	m["id"] = "test_id"
	json := MapToJson(m)

	rm := MapFromJson(strings.NewReader(json))

	if rm["id"] != "test_id" {
		t.Fatal("map should be valid")
	}

	rm2 := MapFromJson(strings.NewReader(""))
	if len(rm2) > 0 {
		t.Fatal("make should be ivalid")
	}
}

func TestIsValidEmail(t *testing.T) {
	for _, testCase := range []struct {
		Input    string
		Expected bool
	}{
		{
			Input:    "corey",
			Expected: false,
		},
		{
			Input:    "corey@example.com",
			Expected: true,
		},
		{
			Input:    "corey+test@example.com",
			Expected: true,
		},
		{
			Input:    "@corey+test@example.com",
			Expected: false,
		},
		{
			Input:    "firstname.lastname@example.com",
			Expected: true,
		},
		{
			Input:    "firstname.lastname@subdomain.example.com",
			Expected: true,
		},
		{
			Input:    "123454567@domain.com",
			Expected: true,
		},
		{
			Input:    "email@domain-one.com",
			Expected: true,
		},
		{
			Input:    "email@domain.co.jp",
			Expected: true,
		},
		{
			Input:    "firstname-lastname@domain.com",
			Expected: true,
		},
		{
			Input:    "@domain.com",
			Expected: false,
		},
		{
			Input:    "Billy Bob <billy@example.com>",
			Expected: false,
		},
		{
			Input:    "email.domain.com",
			Expected: false,
		},
		{
			Input:    "email.@domain.com",
			Expected: false,
		},
		{
			Input:    "email@domain@domain.com",
			Expected: false,
		},
		{
			Input:    "(email@domain.com)",
			Expected: false,
		},
		{
			Input:    "email@汤.中国",
			Expected: true,
		},
		{
			Input:    "email1@domain.com, email2@domain.com",
			Expected: false,
		},
	} {
		t.Run(testCase.Input, func(t *testing.T) {
			assert.Equal(t, testCase.Expected, IsValidEmail(testCase.Input))
		})
	}
}

func TestValidLower(t *testing.T) {
	if !IsLower("corey+test@hulen.com") {
		t.Error("should be valid")
	}

	if IsLower("Corey+test@hulen.com") {
		t.Error("should be invalid")
	}
}

func TestEtag(t *testing.T) {
	etag := Etag("hello", 24)
	require.NotEqual(t, "", etag)
}

var hashtags = map[string]string{
	"#test":           "#test",
	"test":            "",
	"#test123":        "#test123",
	"#123test123":     "",
	"#test-test":      "#test-test",
	"#test?":          "#test",
	"hi #there":       "#there",
	"#bug #idea":      "#bug #idea",
	"#bug or #gif!":   "#bug #gif",
	"#hüllo":          "#hüllo",
	"#?test":          "",
	"#-test":          "",
	"#yo_yo":          "#yo_yo",
	"(#brakets)":      "#brakets",
	")#stekarb(":      "#stekarb",
	"<#less_than<":    "#less_than",
	">#greater_than>": "#greater_than",
	"-#minus-":        "#minus",
	"_#under_":        "#under",
	"+#plus+":         "#plus",
	"=#equals=":       "#equals",
	"%#pct%":          "#pct",
	"&#and&":          "#and",
	"^#hat^":          "#hat",
	"##brown#":        "#brown",
	"*#star*":         "#star",
	"|#pipe|":         "#pipe",
	":#colon:":        "#colon",
	";#semi;":         "#semi",
	"#Mötley;":        "#Mötley",
	".#period.":       "#period",
	"¿#upside¿":       "#upside",
	"\"#quote\"":      "#quote",
	"/#slash/":        "#slash",
	"\\#backslash\\":  "#backslash",
	"#a":              "",
	"#1":              "",
	"foo#bar":         "",
}

func TestStringArray_Equal(t *testing.T) {
	for name, tc := range map[string]struct {
		Array1   StringArray
		Array2   StringArray
		Expected bool
	}{
		"Empty": {
			nil,
			nil,
			true,
		},
		"EqualLength_EqualValue": {
			StringArray{"123"},
			StringArray{"123"},
			true,
		},
		"DifferentLength": {
			StringArray{"123"},
			StringArray{"123", "abc"},
			false,
		},
		"DifferentValues_EqualLength": {
			StringArray{"123"},
			StringArray{"abc"},
			false,
		},
		"EqualLength_EqualValues": {
			StringArray{"123", "abc"},
			StringArray{"123", "abc"},
			true,
		},
		"EqualLength_EqualValues_DifferentOrder": {
			StringArray{"abc", "123"},
			StringArray{"123", "abc"},
			false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.Array1.Equals(tc.Array2))
		})
	}
}

func TestParseHashtags(t *testing.T) {
	for input, output := range hashtags {
		if o, _ := ParseHashtags(input); o != output {
			t.Fatal("failed to parse hashtags from input=" + input + " expected=" + output + " actual=" + o)
		}
	}
}

func TestIsValidAlphaNum(t *testing.T) {
	cases := []struct {
		Input  string
		Result bool
	}{
		{
			Input:  "test",
			Result: true,
		},
		{
			Input:  "test-name",
			Result: true,
		},
		{
			Input:  "test--name",
			Result: true,
		},
		{
			Input:  "test__name",
			Result: true,
		},
		{
			Input:  "-",
			Result: false,
		},
		{
			Input:  "__",
			Result: false,
		},
		{
			Input:  "test-",
			Result: false,
		},
		{
			Input:  "test--",
			Result: false,
		},
		{
			Input:  "test__",
			Result: false,
		},
		{
			Input:  "test:name",
			Result: false,
		},
	}

	for _, tc := range cases {
		actual := IsValidAlphaNum(tc.Input)
		if actual != tc.Result {
			t.Fatalf("case: %v\tshould returned: %#v", tc, tc.Result)
		}
	}
}

func TestGetServerIpAddress(t *testing.T) {
	if len(GetServerIpAddress()) == 0 {
		t.Fatal("Should find local ip address")
	}
}

func TestIsValidAlphaNumHyphenUnderscore(t *testing.T) {
	casesWithFormat := []struct {
		Input  string
		Result bool
	}{
		{
			Input:  "test",
			Result: true,
		},
		{
			Input:  "test-name",
			Result: true,
		},
		{
			Input:  "test--name",
			Result: true,
		},
		{
			Input:  "test__name",
			Result: true,
		},
		{
			Input:  "test_name",
			Result: true,
		},
		{
			Input:  "test_-name",
			Result: true,
		},
		{
			Input:  "-",
			Result: false,
		},
		{
			Input:  "__",
			Result: false,
		},
		{
			Input:  "test-",
			Result: false,
		},
		{
			Input:  "test--",
			Result: false,
		},
		{
			Input:  "test__",
			Result: false,
		},
		{
			Input:  "test:name",
			Result: false,
		},
	}

	for _, tc := range casesWithFormat {
		actual := IsValidAlphaNumHyphenUnderscore(tc.Input, true)
		if actual != tc.Result {
			t.Fatalf("case: %v\tshould returned: %#v", tc, tc.Result)
		}
	}

	casesWithoutFormat := []struct {
		Input  string
		Result bool
	}{
		{
			Input:  "test",
			Result: true,
		},
		{
			Input:  "test-name",
			Result: true,
		},
		{
			Input:  "test--name",
			Result: true,
		},
		{
			Input:  "test__name",
			Result: true,
		},
		{
			Input:  "test_name",
			Result: true,
		},
		{
			Input:  "test_-name",
			Result: true,
		},
		{
			Input:  "-",
			Result: true,
		},
		{
			Input:  "_",
			Result: true,
		},
		{
			Input:  "test-",
			Result: true,
		},
		{
			Input:  "test--",
			Result: true,
		},
		{
			Input:  "test__",
			Result: true,
		},
		{
			Input:  ".",
			Result: false,
		},

		{
			Input:  "test,",
			Result: false,
		},
		{
			Input:  "test:name",
			Result: false,
		},
	}

	for _, tc := range casesWithoutFormat {
		actual := IsValidAlphaNumHyphenUnderscore(tc.Input, false)
		if actual != tc.Result {
			t.Fatalf("case: '%v'\tshould returned: %#v", tc.Input, tc.Result)
		}
	}
}

func TestIsValidId(t *testing.T) {
	cases := []struct {
		Input  string
		Result bool
	}{
		{
			Input:  NewId(),
			Result: true,
		},
		{
			Input:  "",
			Result: false,
		},
		{
			Input:  "junk",
			Result: false,
		},
		{
			Input:  "qwertyuiop1234567890asdfg{",
			Result: false,
		},
		{
			Input:  NewId() + "}",
			Result: false,
		},
	}

	for _, tc := range cases {
		actual := IsValidId(tc.Input)
		if actual != tc.Result {
			t.Fatalf("case: %v\tshould returned: %#v", tc, tc.Result)
		}
	}
}

func TestNowhereNil(t *testing.T) {
	t.Parallel()

	var nilStringPtr *string
	var nonNilStringPtr *string = new(string)
	var nilSlice []string
	var nilStruct *struct{}
	var nilMap map[bool]bool

	var nowhereNilStruct = struct {
		X *string
		Y *string
	}{
		nonNilStringPtr,
		nonNilStringPtr,
	}
	var somewhereNilStruct = struct {
		X *string
		Y *string
	}{
		nonNilStringPtr,
		nilStringPtr,
	}

	var privateSomewhereNilStruct = struct {
		X *string
		y *string
	}{
		nonNilStringPtr,
		nilStringPtr,
	}

	testCases := []struct {
		Description string
		Value       interface{}
		Expected    bool
	}{
		{
			"nil",
			nil,
			false,
		},
		{
			"empty string",
			"",
			true,
		},
		{
			"non-empty string",
			"not empty!",
			true,
		},
		{
			"nil string pointer",
			nilStringPtr,
			false,
		},
		{
			"non-nil string pointer",
			nonNilStringPtr,
			true,
		},
		{
			"0",
			0,
			true,
		},
		{
			"1",
			1,
			true,
		},
		{
			"0 (int64)",
			int64(0),
			true,
		},
		{
			"1 (int64)",
			int64(1),
			true,
		},
		{
			"true",
			true,
			true,
		},
		{
			"false",
			false,
			true,
		},
		{
			"nil slice",
			nilSlice,
			// A nil slice is observably the same as an empty slice, so allow it.
			true,
		},
		{
			"empty slice",
			[]string{},
			true,
		},
		{
			"slice containing nils",
			[]*string{nil, nil},
			true,
		},
		{
			"nil map",
			nilMap,
			false,
		},
		{
			"non-nil map",
			make(map[bool]bool),
			true,
		},
		{
			"non-nil map containing nil",
			map[bool]*string{true: nilStringPtr, false: nonNilStringPtr},
			// Map values are not checked
			true,
		},
		{
			"nil struct",
			nilStruct,
			false,
		},
		{
			"empty struct",
			struct{}{},
			true,
		},
		{
			"struct containing no nil",
			nowhereNilStruct,
			true,
		},
		{
			"struct containing nil",
			somewhereNilStruct,
			false,
		},
		{
			"struct pointer containing no nil",
			&nowhereNilStruct,
			true,
		},
		{
			"struct pointer containing nil",
			&somewhereNilStruct,
			false,
		},
		{
			"struct containing private nil",
			privateSomewhereNilStruct,
			true,
		},
		{
			"struct pointer containing private nil",
			&privateSomewhereNilStruct,
			true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic: %v", r)
				}
			}()

			t.Parallel()
			require.Equal(t, testCase.Expected, checkNowhereNil(t, "value", testCase.Value))
		})
	}
}

// checkNowhereNil checks that the given interface value is not nil, and if a struct, that all of
// its public fields are also nowhere nil
func checkNowhereNil(t *testing.T, name string, value interface{}) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	switch v.Type().Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			t.Logf("%s was nil", name)
			return false
		}

		return checkNowhereNil(t, fmt.Sprintf("(*%s)", name), v.Elem().Interface())

	case reflect.Map:
		if v.IsNil() {
			t.Logf("%s was nil", name)
			return false
		}

		// Don't check map values
		return true

	case reflect.Struct:
		nowhereNil := true
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			// Ignore unexported fields
			if v.Type().Field(i).PkgPath != "" {
				continue
			}

			nowhereNil = nowhereNil && checkNowhereNil(t, fmt.Sprintf("%s.%s", name, v.Type().Field(i).Name), f.Interface())
		}

		return nowhereNil

	case reflect.Array:
		fallthrough
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		fallthrough
	case reflect.Interface:
		fallthrough
	case reflect.UnsafePointer:
		t.Logf("unhandled field %s, type: %s", name, v.Type().Kind())
		return false

	default:
		return true
	}
}
