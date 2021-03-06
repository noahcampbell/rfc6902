package rfc6902

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrorInvalidJSONPath = errors.New("Invalid JSON Path")

/*
Patcher to maniplate a json doc.
*/
type patcher struct {
	pointer    jsonptr
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

func (p *patcher) exists() bool {
	_, err := value(p.pointer, &p.jsonObject)
	return err == nil
}

func (p *patcher) setExistingValue(v interface{}) error {
	ref, err := p.parentValue()
	if err != nil {
		return err
	}

	switch t := ref.(type) {
	case map[string]interface{}:
		t[p.pointer[len(p.pointer)-1].token()] = v
	case []interface{}:

		parent, err := p.parentValue()
		if err != nil {
			return err
		}

		pa, ok := parent.([]interface{})
		if !ok {
			panic("unable to convert to []interface{}")
		}

		index := p.pointer[len(p.pointer)-1].token()
		var na []interface{}
		if index == "-" {
			na = append(pa, v)
		} else {
			i, err := strconv.Atoi(index)
			if err != nil {
				return err
			}
			na = append(pa[:i], append([]interface{}{v}, pa[i:]...)...)
		}

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
	if err != nil {
		return nil, err
	}
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

	parent := p.parent()
	switch t := ref.(type) {
	case map[string]interface{}:
		delete(t, p.pointer[len(p.pointer)-1].token())
	case []interface{}:
		i, err := strconv.Atoi(p.pointer[len(p.pointer)-1].token())
		if err != nil {
			return err
		}
		parentValue, err := parent.value()
		if err != nil {
			return err
		}
		container, ok := parentValue.([]interface{})
		if !ok {
			panic("Unable to convert parent to []interface{}")
		}

		newArray := make([]interface{}, 0)
		newArray = append(newArray, container[:i]...)
		newArray = append(newArray, container[i+1:]...)
		parent.setExistingValue(newArray)
	default:
		panic("Unknown type for removal")
	}
	return nil
}

func (p *patcher) replace(o interface{}) error {
	ref, err := p.parentValue()
	if err != nil {
		return err
	}

	switch t := ref.(type) {
	case map[string]interface{}:
		_ = t
		p.setExistingValue(o)
	case []interface{}:
		parent := p.parent()
		i, err := strconv.Atoi(p.pointer[len(p.pointer)-1].token())
		if err != nil {
			return err
		}
		parentValue, err := parent.value()
		if err != nil {
			return err
		}
		container, ok := parentValue.([]interface{})
		if !ok {
			panic("Unable to convert parent to []interface{}")
		}
		container[i] = o
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
