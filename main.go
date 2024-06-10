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

type Query struct {
	tokens *[]string
	alias  string
}

type tQuery struct {
	columns        []string
	fromSources    []string
	joins          []string
	whereClauses   []string
	havingClauses  []string
	groupByColumns []string
	sortColumns    []string
	limit          []string
}
type fromSources struct {
	table    string
	subquery *Query
	alias    string
}
type columns struct {
	col      string
	subquery *Query
	alias    string
}
type where struct {
	left  string
	right string
	ttype string
}
type join struct {
	table    string
	subquery *Query
	alias    string
	exp      *where
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
	var deep func(node *node, alias string)
	deep = func(node *node, alias string) {
		var str []string
		query := Query{
			tokens: &str,
			alias:  alias,
		}
		res = append(res, &query)
		t := node
		for t.next != nil {
			if t.nextnext != nil {
				if t.next.next.token == "AS" {
					deep(t.nextnext, t.next.next.next.token)
					str = append(str, "@subquery")
					str = append(str, "AS")

				} else {
					deep(t.nextnext, t.next.next.token)
					str = append(str, "@subquery")
					str = append(str, t.next.next.token)
				}
				t = t.next.next.next
			} else {
				str = append(str, t.token)
				t = t.next
			}

		}
	}
	deep(_node.next.nextnext, "main")
	return res
}

func buildQuery(item *Query, alias string) baseQuery {
	node := tQuery{}
	tokens := *item.tokens
	isColumns := false
	isFromSources := false
	isJoin := false
	isWhere := false
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "SELECT" {
			isColumns = true
			continue
		}
		if tokens[i] == "FROM" {
			isColumns = false
			isFromSources = true
			continue
		}
		if tokens[i] == "JOIN" {
			isJoin = true
			isColumns = false
			isFromSources = false
			continue
		}
		if tokens[i] == "WHERE" {
			isJoin = false
			isColumns = false
			isFromSources = false
			isWhere = true
			continue
		}
		if tokens[i] == "" {
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
		t.alias = node.fromSources[i+1]
		i++
	}
	base.fromSources = &t
	var tc []*columns
	for i := 0; i < len(node.columns); i++ {
		if i+1 < len(node.columns) && node.columns[i+1] == "AS" {
			tc = append(tc, &columns{
				col:   node.columns[i],
				alias: node.columns[i+2],
			})
			i++
			i++
		} else {
			tc = append(tc, &columns{
				col: node.columns[i],
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
			alias: node.joins[i+1],
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
	sql := "( SELECT * FROM diaries a JOIN ( SELECT * FROM posts p ) gg ON a.id = gg.di where a.id = 1337 )"
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
	var queries []baseQuery
	for _, item := range b {
		queries = append(queries, buildQuery(item, item.alias))
	}

	fmt.Print(1)
}
