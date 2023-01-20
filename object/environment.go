package object

import (
	"io"
	"sort"
)





func NewEnclosedEnvironment(outer *Environment, args []Object) *Environment {
	env := NewEnvironment(outer.Writer, outer.Dir, outer.Version)
	env.outer = outer
	env.CurrentArgs = args
	return env
}





func NewEnvironment(w io.Writer, dir string, version string) *Environment {
	s := make(map[string]Object)
	
	
	e := &Environment{store: s, outer: nil, Writer: w, Dir: dir, Version: version}
	e.Set("ANK_VERSION", &String{Value: e.Version})

	return e
}




type Environment struct {
	store map[string]Object
	
	
	
	
	
	
	
	
	
	CurrentArgs []Object
	outer       *Environment
	
	
	Writer io.Writer
	
	
	
	
	
	
	
	
	
	Dir string
	
	Version string
}


func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}



func (e *Environment) GetKeys() []string {
	keys := make([]string, 0, len(e.store))
	for k := range e.store {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}


func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}


func (e *Environment) Delete(name string) {
	delete(e.store, name)
}
