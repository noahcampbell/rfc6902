package rfc6902

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"errors"
)

// decode according to Section 3. Syntax
func decode(in string) string {
	in = strings.Replace(in, "~1", "/", -1)
	return strings.Replace(in, "~0", "~", -1)
}

type Inserter interface {
	Insert(v interface{})
}

type Overwriter interface {
	Overwrite(v interface{})
}

type Valuer interface {
	Value() interface{}
}

type PtrMutator interface {
	Inserter
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

type refPointer struct {
	value interface{}
}

func (r *refPointer) Value() interface{} {
	return r.value
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

var ErrorInvalidJSONPath = errors.New("invalid pointer path")

// todo remove
type ValueInserter interface {
	Valuer
	Inserter
}

func jsonPointer(path string, doc interface{}) (value ValueInserter, err error) {

	if len(path) > 0 && path[0] == '#' {
		path, err := url.QueryUnescape(path[1:])
		if err != nil {
			return nil, err
		}
		return jsonPointer(path, doc)
	}

	ref := doc
	value = &refPointer{doc}
	fields := jsonPathFields(path)
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

type jsonptr []reftoken

// return the escaped path (see section 3. Syntax)
func (j jsonptr) remainder(i int) (s string) {
	for _, ref := range j[i:] {
		s += "/" + string(ref)
	}
	s = s[1:]
	return
}

func (j jsonptr) String() (s string) {
	for _, ref := range j {
		s += "/" + ref.token()
	}
	return
}

type reftoken string

func (r reftoken) token() string {
	return decode(string(r))
}

func newRefToken(in string) reftoken {
	if len(in) <= 0 {
		panic("jsonptr cannot be formed from zero length string")
	}
	if in[0] != '/' {
		panic("jsonptr must contain a leading '/'")
	}
	return reftoken(in[1:])
}

func jsonPathFields(path string) (head jsonptr) {
	s := path
	for len(s) > 0 {
		if s[0] == '/' {
			if len(s) == 1 {
				head = append(head, newRefToken(s))
				break
			}

			next := strings.Index(s[1:], "/")
			if next == -1 {
				head = append(head, newRefToken(s))
				break
			}

			next += 1
			head = append(head, newRefToken(s[:next]))
			s = s[next:]
		} else {
			panic(fmt.Sprintf("field must start with '/': %q", s))
		}
	}
	return
}
