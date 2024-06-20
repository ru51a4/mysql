package main

import (
	"fmt"
	"strings"
)

type table struct {
	cols []string
	row  [][]string
}

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
			s = s + strings.ToUpper(string(str[i]))
		}
	}
	return res
}

type ttoken struct {
	str       string
	subquery  *Query
	_subQuery *baseQuery
	_func     *_func
}
type _func struct {
	opertator string
	alias     string
	args      *[]ttoken
}
type Query struct {
	tokens *[]ttoken
	alias  string
}

type tQuery struct {
	columns        []ttoken
	fromSources    []ttoken
	joins          []ttoken
	whereClauses   []ttoken
	havingClauses  []ttoken
	groupByColumns []ttoken
	sortColumns    []ttoken
	limit          []ttoken
}
type fromSources struct {
	table ttoken
	alias string
}
type columns struct {
	col   ttoken
	alias string
	_func *_func
}
type where struct {
	left  ttoken
	right ttoken
	ttype ttoken
}
type join struct {
	table ttoken
	alias string
	exp   *where
}
type baseQuery struct {
	fromSources *fromSources
	columns     []*columns
	where       []*where
	join        []*join
	alias       string
}

func subQuery(_node *node) []*Query {
	var res []*Query

	var buildFunc func(node *node, alias string, operator string) _func
	buildFunc = func(node *node, alias string, operator string) _func {
		var str []ttoken

		query := _func{
			opertator: operator,
			alias:     alias,
			args:      &str,
		}
		t := node
		for t.next != nil {
			if t.nextnext != nil {
				if t.nextnext.next == nil || t.nextnext.next.next == nil || t.nextnext.next.next.token != "SELECT" {
					ft := buildFunc(t.nextnext, "", t.nextnext.token)
					str = append(str, ttoken{_func: &ft})
					t = t.next
				}
			} else {
				//todo
				if t.next.token != "" {
					str = append(str, ttoken{str: t.next.token})
				}
				t = t.next
			}
		}

		return query

	}

	var deep func(node *node, alias string) Query
	deep = func(node *node, alias string) Query {
		var str []ttoken
		query := Query{
			tokens: &str,
			alias:  alias,
		}
		res = append(res, &query)
		t := node
		for t.next != nil {
			if t.nextnext != nil {
				if t.nextnext.next == nil || t.nextnext.next.next == nil || t.nextnext.next.next.token != "SELECT" {
					ft := buildFunc(t.nextnext, "", t.nextnext.token)
					str = append(str, ttoken{_func: &ft})
					t = t.next
					continue
				}
				if t.next.next.token == "AS" {
					st := deep(t.nextnext, t.next.next.next.token)
					str = append(str, ttoken{
						subquery: &st,
					})
					str = append(str, ttoken{str: "AS"})

				} else {
					st := deep(t.nextnext, t.next.next.token)
					str = append(str, ttoken{
						subquery: &st,
					})
					str = append(str, ttoken{str: t.next.next.token})
				}
				t = t.next.next.next
			} else {
				str = append(str, ttoken{str: t.token})
				t = t.next
			}

		}
		return query
	}
	deep(_node.next.nextnext, "main")
	return res
}

func buildQuery(b []*Query, item *Query, alias string) baseQuery {
	node := tQuery{}
	tokens := *item.tokens
	isColumns := false
	isFromSources := false
	isJoin := false
	isWhere := false
	for i := 0; i < len(tokens); i++ {
		if tokens[i].str == "SELECT" {
			isColumns = true
			continue
		}
		if tokens[i].str == "FROM" {
			isColumns = false
			isFromSources = true
			continue
		}
		if tokens[i].str == "JOIN" {
			isJoin = true
			isColumns = false
			isFromSources = false
			continue
		}
		if tokens[i].str == "WHERE" {
			isJoin = false
			isColumns = false
			isFromSources = false
			isWhere = true
			continue
		}
		if tokens[i].subquery == nil && tokens[i].str == "" && tokens[i]._func == nil {
			continue
		}
		if tokens[i].subquery != nil {
			_subQuery := buildQuery(b, tokens[i].subquery, tokens[i].subquery.alias)
			if isColumns {
				node.columns = append(node.columns, ttoken{_subQuery: &_subQuery})
			}
			if isFromSources {
				node.fromSources = append(node.fromSources, ttoken{_subQuery: &_subQuery})
			}
			if isJoin {
				node.joins = append(node.joins, ttoken{_subQuery: &_subQuery})
			}
			if isWhere {
				node.whereClauses = append(node.whereClauses, ttoken{_subQuery: &_subQuery})
			}
			continue
		}
		if isColumns {
			node.columns = append(node.columns, tokens[i])
		}
		if isFromSources {
			node.fromSources = append(node.fromSources, tokens[i])
		}
		if isJoin {
			node.joins = append(node.joins, tokens[i])
		}
		if isWhere {
			node.whereClauses = append(node.whereClauses, tokens[i])
		}
	}
	base := baseQuery{
		alias: alias,
	}
	t := fromSources{}
	for i := 0; i < len(node.fromSources); i++ {
		t.table = node.fromSources[i]
		t.alias = node.fromSources[i+1].str
		i++
	}
	base.fromSources = &t
	var tc []*columns
	for i := 0; i < len(node.columns); i++ {
		if i+1 < len(node.columns) && node.columns[i+1].str == "AS" {
			tc = append(tc, &columns{
				col:   node.columns[i],
				alias: node.columns[i+2].str,
				_func: node.columns[i]._func,
			})
			i++
			i++
		} else {
			tc = append(tc, &columns{
				col:   node.columns[i],
				_func: node.columns[i]._func,
			})
		}
	}
	base.columns = tc

	var tw []*where
	for i := 0; i < len(node.whereClauses); i++ {
		tw = append(tw, &where{left: node.whereClauses[i], right: node.whereClauses[i+2], ttype: node.whereClauses[i+1]})
		i++
		i++
	}
	base.where = tw

	var tj []*join
	for i := 0; i < len(node.joins); i++ {
		tj = append(tj, &join{
			table: node.joins[i],
			alias: node.joins[i+1].str,
			exp:   &where{left: node.joins[i+3], ttype: node.joins[i+4], right: node.joins[i+5]},
		})
		i++
		i++
		i++
		i++
		i++
	}
	base.join = tj
	return base
}
func indexOf(arr []string, need string) int {
	for ind, val := range arr {
		if val == need {
			return ind
		}
	}
	return -1
}

func eval(aquery string, arr []baseQuery, _table map[string]*table) string {
	// todo
	return ""
}
func main() {
	sql := "( SELECT id, max(id) FROM diaries a JOIN ( SELECT id, max(max(id)) FROM posts p ) gg ON a.id = gg.di where a.id = 1337 )"
	_table := make(map[string]*table)
	_table["POSTS"] = &table{
		cols: []string{"ID", "USER_ID"},
		row:  [][]string{[]string{"1", "1"}, []string{"2", "2"}},
	}
	_table["USERS"] = &table{
		cols: []string{"ID", "LOGIN"},
		row:  [][]string{[]string{"1", "admin"}, []string{"2", "user"}},
	}
	a := lexer(sql)
	b := subQuery(a)
	tree := buildQuery(b, b[0], "main")

	fmt.Print(tree)
}
