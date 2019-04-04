package sql

type Type interface {
	isSQLType()
}

// Table represents a sql table
type Table struct{}

func (Table) isSQLType() {}
func (Table) String() string {
	return "table"
}

// View  represents a sql view
type View struct{}

func (View) isSQLType() {}
func (View) String() string {
	return "view"
}

// Function represents a sql view
type Function struct{}

func (Function) isSQLType() {}
func (Function) String() string {
	return "function"
}

// Object represents a sql object and its name.
type Object struct {
	Type Type
	Name string
}

type Objects []Object

type Dependecies Objects
type Definitions Objects

func (d Dependecies) Unique() Dependecies { return Dependecies(Objects(d).Unique()) }
func (d Definitions) Unique() Definitions { return Definitions(Objects(d).Unique()) }

func (o Objects) Unique() Objects {
	objMap := map[Object]struct{}{}
	for _, obj := range o {
		if _, has := objMap[obj]; has {
			continue
		}
		objMap[obj] = struct{}{}
	}
	objs := Objects{}
	for obj := range objMap {
		objs = append(objs, obj)
	}
	return objs
}
