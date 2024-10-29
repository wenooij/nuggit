package views

import (
	"sync"

	"github.com/wenooij/nuggit"
)

type Index struct {
	viewsByUUID    map[string]nuggit.View
	viewsByName    map[string][]nuggit.View
	viewUUIDByName map[string][]string
	once           sync.Once
}

func (i *Index) Reset() {
	i.viewsByUUID = make(map[string]nuggit.View)
	i.viewsByName = make(map[string][]nuggit.View)
	i.viewUUIDByName = make(map[string][]string)
}

func (i *Index) Add(name, uuid string, view nuggit.View) {
	i.once.Do(i.Reset)
	if _, ok := i.viewsByUUID[uuid]; ok {
		return
	}
	i.viewsByUUID[uuid] = view
	i.viewsByName[name] = append(i.viewsByName[name], view)
	i.viewUUIDByName[name] = append(i.viewUUIDByName[name], uuid)
}

func (i *Index) AddName(name string, view nuggit.View) {
	i.once.Do(i.Reset)
	i.viewsByName[name] = append(i.viewsByName[name], view)
}

func (i *Index) Has(uuid string) bool {
	_, found := i.viewsByUUID[uuid]
	return found
}

func (i *Index) HasName(name string) bool {
	views := i.viewsByName[name]
	return len(views) > 0
}

func (i *Index) Get(uuid string) (nuggit.View, bool) {
	view, found := i.viewsByUUID[uuid]
	return view, found
}

func (i *Index) GetUnique(name string) (view nuggit.View, ok bool) {
	views := i.viewsByName[name]
	if len(views) != 1 {
		return nuggit.View{}, false
	}
	return views[0], true
}
