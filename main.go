package rfc6902

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type op struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value interface{}
}

func (o *op) apply(v interface{}) error {
	switch o.Op {
	case "add":
		return o.add(v)
	default:
		return errors.New("rfc6902: unknown operation")
	}
}

func (o *op) add(v interface{}) error {
	v, leaf := o.traversePath(v)
	switch t := v.(type) {
	case map[string]interface{}:
		t[leaf] = o.Value
	default:
		panic("Unknown")
	}
	return nil
}

func (o *op) traversePath(v interface{}) (ref interface{}, leaf string) {
	var (
		miss bool
	)
	for _, el := range strings.Split(o.Path, "/") {
		if len(el) <= 0 {
			continue
		}
		leaf = el
		if miss {
			panic("too many levels.") // todo fix this up
		}

		ref = v
		switch t := v.(type) {
		case map[string]interface{}:
			vv, ok := t[el]
			if !ok {
				miss = true
				continue
			}
			ref = vv
		case []interface{}:
			panic("arrays...don't know what to do.")
		case string:
			panic("strings...don't know what to do.")
		case interface{}:
			panic("object...don't know what to do.")
		default:
			fmt.Printf("Unknown type %T\n", t)
			panic("wtf")
		}

	}
	return
}

type Patcher struct {
	ops []op
}

func ParsePatch(r io.Reader) (*Patcher, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}

	b := new(bytes.Buffer)
	b.ReadFrom(r)
	p := new(Patcher)
	p.ops = make([]op, 0)
	err := json.Unmarshal(b.Bytes(), &p.ops)
	if err != nil {
		return nil, err
	}

	for pos, op := range p.ops {
		if len(op.Op) == 0 {
			return nil, fmt.Errorf("rfc6902: missing op at %d (section 4 Operations)", pos)
		}
		if len(op.Path) == 0 {
			return nil, fmt.Errorf("rfc6902: missing path at %d (section 4 Operations)", pos)
		}
		switch op.Op {
		case "add":
			if op.Value == nil {
				return nil, fmt.Errorf("rfc6902: missing value for add op (section 4.1 add)")
			}
		}
	}

	return p, nil
}

func (p *Patcher) Apply(b []byte) ([]byte, error) {
	if len(b) <= 0 {
		return nil, errors.New("rfc6902: empty JSON document")
	}
	var v interface{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}

	for _, op := range p.ops {
		err := op.apply(v)
		if err != nil {
			panic(err)
		}
	}

	return json.Marshal(v)
}
