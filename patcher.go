package rfc6902

import (
	"strconv"
	"errors"
	"fmt"
)

var ErrorInvalidJSONPath = errors.New("Invalid JSON Path")

/*
Patcher to maniplate a json doc.
*/
type patcher struct {
	pointer jsonptr
	jsonObject interface{}
}

func (p *patcher) parent() *patcher {
	if len(p.pointer) == 1 {
		return nil
	}
	return &patcher{jsonptr(p.pointer[:len(p.pointer)-1]), p.jsonObject}
}

func (p *patcher) value() (interface{}, error) {
	v, err := value(p.pointer, &p.jsonObject)
	if err != nil {
		return nil, err
	}
	return *v, nil
}

func (p *patcher) existing() bool {
	_, err := value(p.pointer, &p.jsonObject)
	return err == nil
}

func(p *patcher) setExistingValue(v interface{}) error {
	ref, err := p.parentValue()
	if err != nil {
		return err
	}

	switch t := ref.(type) {
	case map[string]interface{}:
		t[p.pointer[len(p.pointer)-1].token()] = v
	case []interface{}:
		i, err := strconv.Atoi(p.pointer[len(p.pointer)-1].token())
		if err != nil {
			return err
		}
		parent, err := p.parentValue()	
		if err != nil {
			return err
		}
		pa, ok := parent.([]interface{})
		if !ok {
			panic("unable to convert to []interface{}")
		}
		na := append(pa[:i], append([]interface{}{v}, pa[i:]...)...)

		parentRef := p.parent()
		if parentRef == nil {
			p.setParentValue(na)
			return nil
		} 

		parentRef.setExistingValue(na)

	default:
		panic("Unknown type.")
	}
	return nil
}

func (p *patcher) parentValue() (interface{}, error) {
	v, err := value(p.pointer[:len(p.pointer)-1], &p.jsonObject)
	return *v, err
}

func (p *patcher) setParentValue(v interface{}) {
	ref, _ := value(p.pointer[:len(p.pointer)-1], &p.jsonObject)
	*ref = v
}

func (p *patcher) remove() error {
	ref, err := p.parentValue()
	if err != nil {
		return err
	}

	switch t := ref.(type) {
	case map[string]interface{}:
		delete(t, p.pointer[len(p.pointer)-1].token())
	}
	return nil
}

func value(fields jsonptr, ref *interface{}) (value *interface{}, err error) {

	value = ref
	for _, field := range fields {
		el := field.token()
		switch t := (*ref).(type) {
		case map[string]interface{}:
			vv, ok := t[el]
			if !ok {
				return nil, ErrorInvalidJSONPath
			}
			value = &vv
			ref = &vv
		case []interface{}:
			idx, err := strconv.Atoi(string(el))
			if err != nil {
				panic("unable to convert to integer")
			}
			if idx >= len(t) {
				return nil, ErrorInvalidJSONPath
			}
			vv := t[idx]
			value = &vv
			ref = &vv
		default:
			panic(fmt.Sprintf("unknown type %T\n", t))
		}
	}
	return
}
