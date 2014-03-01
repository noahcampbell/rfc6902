package rfc6902

import (
	"encoding/json"
"fmt"
	"reflect"
	"testing"
)

func Test_Patcher_ExistingValue(t *testing.T) {
	tests := []struct {
		path     string
		target   string
		expected interface{}
	}{
		{"", "{}", map[string]interface{}{}},
		{"/", `{"": "whiteshadow"}`, "whiteshadow"},
		{"/a", `{"a": "b"}`, "b"},
		{"/1", `["a", "b"]`, "b"},
		{"/0", `["c", "d"]`, "c"},
	}
	for i, test := range tests {
		p := &patcher{ptr(test.path), um(test.target)}
		v, _ := p.existingValue()
		if !reflect.DeepEqual(v, test.expected) {
			t.Errorf("%d: existingValue() (actual) %q != %q (expected)", i, v, test.expected)
		}
		if !p.existing() {
			t.Errorf("existing (actual) false != true (expected)")
		}
	}
}

func Test_Patcher_NotExistingValue(t *testing.T) {
	tests := []struct {
		path     string
		target   string
		expected interface{}
	}{
		{"/a", "{}", map[string]interface{}{}},
		{"/", `{"a": "whiteshadow"}`, map[string]interface{}{"a": "whiteshadow"}},
		{"/c", `{"a": "b"}`, map[string]interface{}{"a": "b"}},
		{"/", `{"z": ["a", "b"]}`, map[string]interface{}{"z":[]interface{}{"a", "b"}}},
		{"/2", `["c", "d"]`, []interface{}{"c", "d"}},
	}
	for i, test := range tests {
		p := &patcher{ptr(test.path), um(test.target)}
		if p.existing() {
			t.Errorf("%d: existing (actual) true != false (expected)", i)
		}

		v, _ := p.parentValue()
		if !reflect.DeepEqual(*v, test.expected) {
			t.Errorf("%d: parentValue() (actual) %#v != %#v (expected)", i, v, test.expected)
		}

		fmt.Printf("p.jsonObject: %#v\n", p.jsonObject)

	}
}

func ptr(j string) jsonptr {
	p, _ := newJSONPointer(j)
	return p
}

func um(j string) interface{} {
	var v interface{}
	err := json.Unmarshal([]byte(j), &v)
	if err != nil {
		panic("No one expects an error: " + err.Error())
	}
	return v
}
