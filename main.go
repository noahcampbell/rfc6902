package rfc6902

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type op struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value interface{}
}

func (o *op) apply(v interface{}) (interface{}, error) {
	ptr, err := newJSONPointer(o.Path)
	if err != nil {
		return nil, err
	}

	switch o.Op {
	case "add":
		return o.add(ptr, v)
	case "remove":
		return o.remove(ptr, v)
	case "replace":
		return o.replace(ptr, v)
	default:
		return nil, errors.New("rfc6902: unknown operation")
	}
}

func (o *op) add(ptr jsonptr, v interface{}) (interface{}, error) {
	p := patcher{ptr, v}
	p.setExistingValue(o.Value)
	return p.jsonObject, nil
}

func (o *op) remove(ptr jsonptr, v interface{}) (interface{}, error) {
	p := patcher{ptr, v}
	p.remove()
	return p.jsonObject, nil
}

func (o *op) replace(ptr jsonptr, v interface{}) (interface{}, error) {
	p := patcher{ptr, v}
	p.replace(o.Value)
	return p.jsonObject, nil
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
		v, err = op.apply(v)
		if err != nil {
			panic(err)
		}
	}

	return json.Marshal(v)
}
