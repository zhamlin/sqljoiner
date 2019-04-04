package main // import sqljoiner

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"sqljoiner/pkg/sql"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	sqlDirectory = flag.String("dir", "", "directory containing schema")
	printJSON    = flag.Bool("json", false, "print out tree representation in json")
	ignoreFiles  arrayFlags
)

// globalHas is a list of objects that already exist in postgres.
var globalHas = map[string]struct{}{
	"pg_catalog": struct{}{},
}

func main() {
	flag.Var(&ignoreFiles, "ignore", "files to exclude")
	flag.Parse()

	if *sqlDirectory == "" {
		panic("must specify directory")
	}
	graph := &sql.DependecyGraph{}
	defToFile := map[string]string{}
	fileCache := map[string]string{}

	toAddEdges := map[string][]func(*sql.DependecyGraph){}
	addEdgeDelayed := func(path, dep string) {
		funcs, has := toAddEdges[dep]
		if !has {
			funcs = []func(*sql.DependecyGraph){}
		}

		funcs = append(funcs, func(g *sql.DependecyGraph) {
			filename, has := defToFile[dep]
			if !has {
				if _, has := globalHas[dep]; !has {
					panic("don't have:" + dep)
				}
			}
			graph.CreateNode(filename)
			fileNode := graph.Node(filename)
			node := graph.Node(path)
			node.AddEdge(fileNode)
		})

		toAddEdges[dep] = funcs
	}

	err := filepath.Walk(*sqlDirectory, func(path string, f os.FileInfo, err error) error {
		for _, ignore := range ignoreFiles {
			if strings.Contains(path, ignore) {
				return nil
			}
		}

		if f.IsDir() {
			return nil
		}

		parsedFile, err := sql.ParseFile(path)
		if err != nil {
			panic(fmt.Sprintf("failed to parse sql file: %v: %v", path, err))
		}
		fileCache[path] = parsedFile.Content
		tree := parsedFile.Tree

		if *printJSON {
			fmt.Println(parsedFile.JSON())
		}

		graph.CreateNode(path)
		node := graph.Node(path)
		if node == nil {
			panic("node not found: " + path)
		}

		defs := sql.GetDefs(tree).Unique()
		for _, def := range defs {
			defToFile[def.Name] = path
		}

		deps := sql.GetDeps(tree).Unique()
		for _, dep := range deps {
			filename, has := defToFile[dep.Name]
			if !has {
				addEdgeDelayed(path, dep.Name)
			} else if filename != path {
				graph.CreateNode(filename)
				node.AddEdge(graph.Node(filename))
			}
		}

		return nil
	})

	if err != nil {
		panic("error walking files: " + err.Error())
	}

	for _, funcs := range toAddEdges {
		for _, f := range funcs {
			f(graph)
		}
	}

	resolved := &sql.DependecyGraph{}
	cycle, err := graph.TopSort(resolved)
	if cycle {
		log.Fatal(err)
	}

	builder := strings.Builder{}
	for _, n := range resolved.Items() {
		builder.WriteString(fileCache[n.Name] + "\n")
	}
	fmt.Println(builder.String())
}
