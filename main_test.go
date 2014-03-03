package rfc6902

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func Test_ParsePatch_NilReader(t *testing.T) {
	_, err := ParsePatch(nil)
	if err == nil {
		t.Fatalf("Expected error for nil reader.")
	}
}

func Test_ParsePatch_EmptyReader(t *testing.T) {
	_, err := ParsePatch(strings.NewReader(""))
	if err == nil {
		t.Fatalf("Parsing must return an error for \"\": %s", err)
	}

}

func Test_ParsePatch_MissingRequiredElements(t *testing.T) {
	tests := []struct {
		invalidPatchDoc string
		expected        error
	}{
		{"[{\"path\": \"/a/b/c/\"}]", fmt.Errorf("rfc6902: missing op at 0 (section 4 Operations)")},
		{"[{\"op\": \"add\"}]", fmt.Errorf("rfc6902: missing path at 0 (section 4 Operations)")},
		{"[{\"op\": \"add\", \"path\": \"/a/b/c\"}]", fmt.Errorf("rfc6902: missing value for add op (section 4.1 add)")},
	}

	for _, test := range tests {
		_, err := ParsePatch(strings.NewReader(test.invalidPatchDoc))
		if err.Error() != test.expected.Error() {
			t.Errorf("For doc: %q (actual) %q != %q (expected)", test.invalidPatchDoc, err, test.expected)
		}
	}
}

func Test_OperationAdd(t *testing.T) {

	patch := `[{ "op": "add", "path": "/baz", "value": "qux" }]`

	p, err := ParsePatch(strings.NewReader(patch))
	if err != nil {
		t.Errorf("Failed parsing: %q. %s", patch, err)
	}

	if len(p.ops) != 1 {
		t.Errorf("Mismatched number of ops (actual) %d != 1 (expected)", len(p.ops))
	}

	if p.ops[0].Op != "add" {
		t.Errorf("(actual) %q != 'add' (expected)", p.ops[0].Op)
	}

	if p.ops[0].Path != "/baz" {
		t.Errorf("(actual) %q != '/baz/' (expected)", p.ops[0].Path)
	}
}

func Test_PatchApply_Empty(t *testing.T) {
	p, _ := ParsePatch(strings.NewReader("[{\"op\": \"add\", \"path\":\"/a/b/c\", \"value\":\"foobar\"}]"))
	_, err := p.Apply([]byte{})
	if err == nil {
		t.Fatalf("Missing error for empty document.")
	}
}

func Test_RFC6902_A1(t *testing.T) {

	tests := []struct {
		rfcTitle, target, patch, expect string
	}{
		{
			rfcTitle: `A.1. Adding an Object Member`,
			target:   `{ "foo": "bar"}`,
			expect:   `{ "baz": "qux", "foo": "bar" }`,
			patch:    `[ { "op": "add", "path": "/baz", "value": "qux" } ]`,
		},
		{
			rfcTitle: "A.2. Adding an Array Element",
			target:   `{ "foo": [ "bar", "baz" ] }`,
			patch:    `[ { "op": "add", "path": "/foo/1", "value": "qux" } ]`,
			expect:   `{ "foo": [ "bar", "qux", "baz" ] }`,
		},
		{
			rfcTitle: "Extra Credit: Array Document",
			target:   `[ "bar", "baz" ] `,
			patch:    `[ { "op": "add", "path": "/1", "value": "qux" } ]`,
			expect:   `[ "bar", "qux", "baz" ]`,
		},
		{
			rfcTitle: "A.3. Remove an Object Member",
			target: `{ "baz": "qux", "foo": "bar" }`,
			patch: `[ { "op": "remove", "path": "/baz" } ]`,
			expect: `{ "foo": "bar" }`,
		},
		{
			rfcTitle: "A.4. Removing an Array Element",
			target: `{ "foo": [ "bar", "qux", "baz" ] }`,
			patch: `[ { "op": "remove", "path": "/foo/1" } ]`,
			expect: `{ "foo": [ "bar", "baz" ] }`,
		},
		{
			rfcTitle: "A.5. Replace a Value",
			target: `{ "baz": "qux", "foo": "bar" }`,
			patch: `[ { "op": "replace", "path": "/baz", "value": "boo" } ]`,
			expect: `{ "baz": "boo", "foo": "bar" }`,
		},
		{
			rfcTitle: "Extra Credit. Replace a Value in an array",
			target: `{ "foo": ["qux", "bar"]}`,
			patch: `[ { "op": "replace", "path": "/foo/0", "value": "baz" } ]`,
			expect: `{ "foo": ["baz", "bar" ]}`,
		},

	}

	for _, test := range tests {
		p, err := ParsePatch(strings.NewReader(test.patch))
		if err != nil {
			t.Fatalf("Failed parsing: %q. %s", test.patch, err)
		}

		result, err := p.Apply([]byte(test.target))
		if err != nil {
			t.Fatalf("%s: Unable to apply patch %s", test.rfcTitle, err)
		}
		if !jsonEqual(result, []byte(test.expect)) {
			t.Errorf("%s\npatch failed => %s\n\nactual:\n%s\n\nexpected:\n%s", test.rfcTitle, test.patch, prettyPrintJson(result), prettyPrintJson([]byte(test.expect)))
		}
	}
}

func jsonEqual(left, right []byte) bool {
	var l, r interface{}
	json.Unmarshal(left, &l)
	json.Unmarshal(right, &r)
	return reflect.DeepEqual(l, r)
}

func prettyPrintJson(src []byte) string {
	b := new(bytes.Buffer)
	json.Indent(b, src, "", "  ")
	return b.String()
}
