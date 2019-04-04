package sql

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"

	pg_query "github.com/lfittl/pg_query_go"
	nodes "github.com/lfittl/pg_query_go/nodes"
)

// File represents a parsed sql file.
type File struct {
	Path        string
	Tree        pg_query.ParsetreeList
	Content     string
	jsonContent string
}

// JSON is used to get a json representation of the psql syntax tree.
func (f *File) JSON() string {
	if f.jsonContent != "" {
		return f.jsonContent
	}

	strTree, err := pg_query.ParseToJSON(f.Content)
	if err != nil {
		panic(err)
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, []byte(strTree), "", " ")
	if err != nil {
		panic(err)
	}

	f.jsonContent = prettyJSON.String()
	return f.jsonContent
}

// ParseFile loads and parses the psql file.
func ParseFile(path string) (*File, error) {
	b, err := ioutil.ReadFile(path) // just pass the file name
	if err != nil {
		return nil, err
	}

	tree, err := pg_query.Parse(string(b))
	if err != nil {
		return nil, err
	}

	return &File{
		Path:    path,
		Content: string(b),
		Tree:    tree,
	}, nil
}

// GetDefs  gets every object that it provides
// views, functions, tables, triggers, etc
func GetDefs(tree pg_query.ParsetreeList) Definitions {
	defs := Definitions{}
	for _, s := range tree.Statements {
		rawStmt, ok := s.(nodes.RawStmt)
		if !ok {
			log.Println("skipping...")
			continue
		}
		switch obj := rawStmt.Stmt.(type) {
		case nodes.CreateStmt:
			defs = append(defs, Object{
				Type: Table{},
				Name: *obj.Relation.Relname,
			})
		case nodes.ViewStmt:
			defs = append(defs, Object{
				Type: View{},
				Name: *obj.View.Relname,
			})
		case nodes.CreateFunctionStmt:
			node := obj.Funcname.Items[0]
			if stringNode, ok := node.(nodes.String); ok {
				defs = append(defs, Object{
					Type: Function{},
					Name: stringNode.Str,
				})
			}
		}
	}
	return defs
}

// GetDeps gets every object that it depends on
// views, functions, tables, triggers, etc
func GetDeps(tree pg_query.ParsetreeList) Dependecies {
	deps := Dependecies{}
	for _, s := range tree.Statements {
		rawStmt, ok := s.(nodes.RawStmt)
		if !ok {
			log.Println("skipping...")
			continue
		}
		switch obj := rawStmt.Stmt.(type) {
		case nodes.SelectStmt:
			deps = append(deps, depSelectStmt(obj)...)
		case nodes.CreateStmt:
			deps = append(deps, depCreateStatement(obj)...)
		case nodes.ViewStmt:
			switch queryStmt := obj.Query.(type) {
			case nodes.SelectStmt:
				deps = append(deps, depSelectStmt(queryStmt)...)
			}
		}
	}
	return deps
}

func depCreateStatement(stmt nodes.CreateStmt) Dependecies {
	deps := Dependecies{}
	for _, node := range stmt.TableElts.Items {
		switch obj := node.(type) {
		case nodes.ColumnDef:
			for _, constraint := range obj.Constraints.Items {
				switch constraintObj := constraint.(type) {
				case nodes.Constraint:
					if constraintObj.Pktable == nil {
						continue
					}
					deps = append(deps, Object{
						Type: Table{},
						Name: *constraintObj.Pktable.Relname,
					})
				}
			}
		}
	}
	return deps
}

func depSelectStmt(stmt nodes.SelectStmt) Dependecies {
	deps := Dependecies{}
	for _, item := range stmt.TargetList.Items {
		switch itemObj := item.(type) {
		case nodes.ResTarget:
			if funcCallNode, ok := itemObj.Val.(nodes.FuncCall); ok {
				if len(funcCallNode.Funcname.Items) < 1 {
					continue
				}
				n := funcCallNode.Funcname.Items[0]
				if stringNode, ok := n.(nodes.String); ok {
					deps = append(deps, Object{
						Type: Function{},
						Name: stringNode.Str,
					})
				}

			}
		}
	}

	for _, item := range stmt.FromClause.Items {
		switch itemObj := item.(type) {
		case nodes.RangeVar:
			deps = append(deps, Object{
				Type: Table{},
				Name: *itemObj.Relname,
			})
		case nodes.JoinExpr:
			switch lObj := itemObj.Larg.(type) {
			case nodes.RangeVar:
				deps = append(deps, Object{
					Type: Table{},
					Name: *lObj.Relname,
				})
			}
			switch rObj := itemObj.Rarg.(type) {
			case nodes.RangeVar:
				deps = append(deps, Object{
					Type: Table{},
					Name: *rObj.Relname,
				})
			}
		}
	}
	return deps
}
