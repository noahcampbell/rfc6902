package rfc6902

import (
	"strconv"
	"errors"
	"fmt"
)

type Inserter interface {
	Insert(v interface{})
}

type Overwriter interface {
	Overwrite(v interface{})
}

type Remover interface {
	Remove()
}

type Valuer interface {
	Value() interface{}
}

type PtrMutator interface {
	Inserter
	Remover
	Overwriter
	Valuer
}

type mapPointer struct {
	mm map[string]interface{}
	el string
}

func (m *mapPointer) Value() interface{} {
	return m.mm[m.el]
}

func (m *mapPointer) Insert(v interface{}) {
	m.mm[m.el] = v
}

func (m *mapPointer) Remove() {
	delete(m.mm, m.el)
}

type refPointer struct {
	value interface{}
}

func (r *refPointer) Value() interface{} {
	return r.value
}

func (r *refPointer) Remove() {
	panic("cannot remove element from refPointer")
}

func (r *refPointer) Insert(v interface{}) {
	panic("cannot insert into refPointer")
}

type arrayPointer struct {
	a []interface{}
	idx int
	last *mapPointer
}

func (r arrayPointer) Value() interface{} {
	return r.a[r.idx]
}

func (r *arrayPointer) Insert(v interface{}) {
	if r.last == nil {
		panic("cannot support top level JSON arrays")
	}
	r.last.Insert(append(r.a[:r.idx], append([]interface{}{v}, r.a[r.idx:]...)...))

}

func (r *arrayPointer) Remove() {
	panic("cannot remove element from arrayPointer")
}

var ErrorInvalidJSONPath = errors.New("invalid pointer path")

// todo remove
type ValueInserter interface {
	Valuer
	Inserter
	Remover
}

type patcher struct {
	pointer jsonptr
	jsonObject interface{}
}

func (p *patcher) existingValue() (interface{}, error) {
	v, err := value(p.pointer, p.jsonObject)
	return *v, err
}

func (p *patcher) existing() bool {
	_, err := value(p.pointer, p.jsonObject)
	return err == nil
}

func (p *patcher) parentValue() (*interface{}, error) {
	return value(p.pointer[:len(p.pointer)-1], p.jsonObject)
}

func value(fields jsonptr, ref interface{}) (value *interface{}, err error) {

	value = &ref
	for _, field := range fields {
		el := field.token()
		switch t := ref.(type) {
		case map[string]interface{}:
			vv, ok := t[el]
			if !ok {
				return nil, ErrorInvalidJSONPath
			}
			value = &vv
			ref = vv
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
			ref = vv
		default:
			panic(fmt.Sprintf("unknown type %T\n", t))
		}
	}
	return
}


func patch(fields jsonptr, doc interface{}) (value ValueInserter, err error) {

	ref := doc
	value = &refPointer{doc}
	var last *mapPointer // needed to support updating arrays until I figure out how to inplace update an array within a struct (may not be possible.  This also don't work when the JSON object is an array.
	for _, field := range fields {
		el := field.token()
		switch t := ref.(type) {
		case map[string]interface{}:
			vv, ok := t[el]
			last = &mapPointer{t, el}
			value = last
			if !ok {
				return value, ErrorInvalidJSONPath
			}
			ref = vv
		case []interface{}:
			idx, err := strconv.Atoi(string(el))
			if err != nil {
				panic("unable to convert to integer")
			}
			value = &arrayPointer{t, idx, last}
			vv := t[idx]
			ref = vv
		default:
			panic(fmt.Sprintf("unknown type %T\n", t))
		}
	}
	return
}


