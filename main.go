package main

import (
	"fmt"
)

type node struct {
	token    string
	next     *node
	nextnext *node
}

func lexer(str string) *node {
	var s string
	var r = &node{}
	res := r
	var stack []*node

	for i := 0; i < len(str); i++ {
		if string(str[i]) == "(" {
			_t := node{
				token: s,
			}
			s = ""
			t := node{
				nextnext: &_t,
			}
			r.next = &t
			r = &_t
			stack = append(stack, &t)
		} else if string(str[i]) == ")" {
			r.next = &node{
				token: s,
			}
			s = ""
			r = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		} else if string(str[i]) == " " {
			r.next = &node{
				token: s,
			}
			r = r.next
			s = ""
		} else {
			s = s + string(str[i])
		}
	}
	return res
}

type Query struct {
	tokens *string
	alias  string
}

func subQuery(_node *node) []*Query {
	var res []*Query
	var deep func(node *node, alias string)
	deep = func(node *node, alias string) {
		str := ""
		query := Query{
			tokens: &str,
			alias:  alias,
		}
		res = append(res, &query)
		t := node
		for t.next != nil {
			if t.nextnext != nil {
				deep(t.nextnext, t.next.next.token)
				str += "( " + t.next.next.token + " )"
				t = t.next.next.next
			} else {
				str += t.token + " "
				t = t.next
			}

		}
	}
	deep(_node.next.nextnext, "main")
	return res
}
func main() {
	sql := "(SELECT * FROM diaries a JOIN (SELECT diary_id as di, id as kek FROM posts p group by diary_id ) gg ON a.id = gg.di ORDER BY gg.kek desc)"
	a := lexer(sql)
	b := subQuery(a)
	fmt.Print(b)
}
