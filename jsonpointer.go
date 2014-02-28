package rfc6902

import (
	"fmt"
	"net/url"
	"strings"
)

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

// decode according to Section 3. Syntax
func decode(in string) string {
	in = strings.Replace(in, "~1", "/", -1)
	return strings.Replace(in, "~0", "~", -1)
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

func newJSONPointer(path string) (head jsonptr, err error) {
	if len(path) > 0 && path[0] == '#' {
		path, err = url.QueryUnescape(path[1:])
		if err != nil {
			return nil, err
		}
	}

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
