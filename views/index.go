package views

import (
	"sync"

	"github.com/wenooij/nuggit/api"
)

type Index struct {
	viewsByUUID    map[string]api.View
	viewsByName    map[string][]api.View
	viewUUIDByName map[string][]string
	once           sync.Once
}

func (i *Index) Reset() {
	i.viewsByUUID = make(map[string]api.View)
	i.viewsByName = make(map[string][]api.View)
	i.viewUUIDByName = make(map[string][]string)
}

func (i *Index) Add(name, uuid string, view api.View) {
	i.once.Do(i.Reset)
	if _, ok := i.viewsByUUID[uuid]; ok {
		return
	}
	i.viewsByUUID[uuid] = view
	i.viewsByName[name] = append(i.viewsByName[name], view)
	i.viewUUIDByName[name] = append(i.viewUUIDByName[name], uuid)
}

func (i *Index) AddName(name string, view api.View) {
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

func (i *Index) Get(uuid string) (api.View, bool) {
	view, found := i.viewsByUUID[uuid]
	return view, found
}

func (i *Index) GetUnique(name string) (view api.View, ok bool) {
	views := i.viewsByName[name]
	if len(views) != 1 {
		return api.View{}, false
	}
	return views[0], true
}
