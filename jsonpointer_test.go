package rfc6902

import (
	"encoding/json"
	"reflect"
	"testing"
)

var target = `
{
	"foo": ["bar", "baz"],
	"": 0,
	"a/b": 1,
	"c%d": 2,
	"e^f": 3,
	"g|h": 4,
	"i\\j": 5,
	"k\"l": 6,
	" ": 7,
	"m~n": 8
}
`

func Test_JSONPointer_RFCExamples(t *testing.T) {
	tests := []struct {
		path, expected string
	}{
		// 5. JSON String Representation
		{"", target},
		{"/foo", "[\"bar\", \"baz\"]"},
		{"/foo/0", "\"bar\""},
		{"/", "0"},
		{"/a~1b", "1"},
		{"/c%d", "2"},
		{"/e^f", "3"},
		{"/g|h", "4"},
		{"/i\\j", "5"},
		{"/k\"l", "6"},
		{"/ ", "7"},
		{"/m~0n", "8"},
		// 6. URI Fragment Identifier Representation
		{"#", target},
		{"#/foo", "[\"bar\", \"baz\"]"},
		{"#/foo/0", "\"bar\""},
		{"#/", "0"},
		{"#/a~1b", "1"},
		{"#/c%25d", "2"},
		{"#/e%5Ef", "3"},
		{"#/g%7Ch", "4"},
		{"#/i%5Cj", "5"},
		{"#/k%22l", "6"},
		{"#/%20", "7"},
		{"#/m~0n", "8"}}

	var v interface{}
	err := json.Unmarshal([]byte(target), &v)
	if err != nil {
		t.Fatalf("Unable to parse target json.")
	}

	for i, test := range tests {
		t.Logf("Testing path: %q", test.path)
		f, _ := newJSONPointer(test.path)
		v, _ := jsonPointer(f, v)
		if !objectJsonCompare(v.Value(), []byte(test.expected)) {
			t.Errorf("%d. %s failed: (actual) %#v != %s (expected)", i, test.path, v, test.expected)
		}
	}

	_ = tests
}

var errorTarget = `{
	"a": "b",
	"d": {"e": "f"}
}`

func Test_JSONPointer_ErrorHandling(t *testing.T) {
	tests := []struct {
		path, element string
	}{
		{"/c", "c"},
		{"/ccc", "ccc"},
		{"/d/g", "g"},
		{"/d/h/j", "h/j"},
		{"/x~1y", "x~1y"},
	}

	var v interface{}
	json.Unmarshal([]byte(errorTarget), &v)
	for _, test := range tests {
		f, _ := newJSONPointer(test.path)
		_, err := jsonPointer(f, v)
		if err != ErrorInvalidJSONPath {
			t.Errorf("%s: is a missing path but retunred err:", test.path, err)
		}
	}
}

func objectJsonCompare(obj interface{}, js []byte) bool {
	var v interface{}
	json.Unmarshal(js, &v)
	return reflect.DeepEqual(obj, v)
}

func Test_ParseIntoFields(t *testing.T) {
	tests := []struct {
		path     string
		length   int
		elements string
	}{
		{"", 0, ""},
		{"/", 1, ""},
		{"/a", 1, "a"},
		{"/a/", 2, "a"},
		{"/a/b", 2, "ab"},
		{"/a/b/", 3, "ab"},
		{"/a/b/c", 3, "abc"},
		{"/a~0/b/c", 3, "a~bc"},
		{"/a~0/b~1~1/c", 3, "a~b//c"},
	}

	for _, test := range tests {
		actual, _ := newJSONPointer(test.path)

		if test.length != len(actual) {
			t.Errorf("%s: field parsing return (actual) %d != %d (expected) elements", test.path, len(actual), test.length)
		}
		acc := ""
		for _, el := range actual {
			acc += el.token()
		}
		if test.elements != acc {
			t.Errorf("%s: elements not recorded correctly (actual) %q != %q (expected)", test.path, acc, test.elements)
		}
	}
}
