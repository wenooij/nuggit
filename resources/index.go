package resources

import (
	"encoding/json"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/wenooij/nuggit/api"
	"gopkg.in/yaml.v3"
)

type Index struct {
	Entries       map[api.NameDigest]*api.Resource
	EntriesByName map[string][]*api.Resource
	Pipes         map[api.NameDigest]*api.Pipe
	Views         map[api.NameDigest]*api.View
}

func (x *Index) Reset() {
	x.Entries = nil
	x.Pipes = nil
	x.Views = nil
}

func (x *Index) GetUnique(name string) *api.Resource {
	entries := x.EntriesByName[name]
	if len(entries) == 0 || len(entries) > 1 {
		return nil
	}
	return entries[0]
}

func (x *Index) GetUniquePipes() map[api.NameDigest]*api.Pipe {
	m := make(map[api.NameDigest]*api.Pipe, len(x.Entries))
	for nd := range x.Entries {
		if pipe := x.GetUnique(nd.GetName()).GetPipe(); pipe != nil {
			m[nd] = pipe
			m[api.NameDigest{Name: nd.Name}] = pipe
		}
	}
	return m
}

func (x *Index) GetUniqueViews(name string) map[api.NameDigest]*api.View {
	m := make(map[api.NameDigest]*api.View, len(x.Entries))
	for nd := range x.Entries {
		if c := x.GetUnique(nd.Name).GetView(); c != nil {
			m[nd] = c
			m[api.NameDigest{Name: nd.Name}] = c
		}
	}
	return m
}

func (x *Index) Add(r *api.Resource) error {
	nd, err := api.NewNameDigest(r)
	if err != nil {
		return err
	}
	if x.Entries == nil {
		x.Entries = make(map[api.NameDigest]*api.Resource, 64)
	}
	x.Entries[nd] = r
	if x.EntriesByName == nil {
		x.EntriesByName = make(map[string][]*api.Resource, 64)
	}
	x.EntriesByName[nd.Name] = append(x.EntriesByName[nd.Name], r)
	switch r.GetKind() {
	case "pipe":
		if x.Pipes == nil {
			x.Pipes = make(map[api.NameDigest]*api.Pipe, 32)
		}
		pipe := r.GetPipe()
		x.Pipes[nd] = pipe
	case "views":
		if x.Views == nil {
			x.Views = make(map[api.NameDigest]*api.View, 4)
		}
		c := r.GetView()
		x.Views[nd] = c
	}
	return nil
}

func (x *Index) AddFS(fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		var unmarshal func([]byte, any) error
		switch ext {
		case ".json":
			unmarshal = json.Unmarshal
		case ".yaml":
			unmarshal = yaml.Unmarshal
		default:
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		var r api.Resource
		if err := unmarshal(data, &r); err != nil {
			log.Println(err)
			return nil
		}

		return x.Add(&r)
	})
}
