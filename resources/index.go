package resources

import (
	"encoding/json"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"gopkg.in/yaml.v3"
)

type Index struct {
	Entries       map[integrity.NameDigest]*api.Resource
	EntriesByName map[string][]*api.Resource
	Pipes         map[integrity.NameDigest]nuggit.Pipe
	Views         map[integrity.NameDigest]*api.View
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

func (x *Index) GetUniquePipes() map[integrity.NameDigest]nuggit.Pipe {
	m := make(map[integrity.NameDigest]nuggit.Pipe, len(x.Entries))
	for nd := range x.Entries {
		if pipe := x.GetUnique(nd.GetName()).GetPipe(); pipe != nil {
			m[nd] = *pipe
			// NB: Digest is omitted here because we want one pipe per name.
			m[integrity.NameDigest{Name: nd.Name}] = *pipe
		}
	}
	return m
}

func (x *Index) GetUniqueViews(name string) map[integrity.NameDigest]*api.View {
	m := make(map[integrity.NameDigest]*api.View, len(x.Entries))
	for nd := range x.Entries {
		if c := x.GetUnique(nd.Name).GetView(); c != nil {
			m[nd] = c
			m[integrity.NameDigest{Name: nd.Name}] = c
		}
	}
	return m
}

func (x *Index) Add(r *api.Resource) error {
	nd, err := integrity.NewNameDigest(r)
	if err != nil {
		return err
	}
	if x.Entries == nil {
		x.Entries = make(map[integrity.NameDigest]*api.Resource, 64)
	}
	x.Entries[nd] = r
	if x.EntriesByName == nil {
		x.EntriesByName = make(map[string][]*api.Resource, 64)
	}
	x.EntriesByName[nd.Name] = append(x.EntriesByName[nd.Name], r)
	switch r.GetKind() {
	case "pipe":
		if x.Pipes == nil {
			x.Pipes = make(map[integrity.NameDigest]nuggit.Pipe, 32)
		}
		pipe := r.GetPipe()
		x.Pipes[nd] = *pipe
	case "views":
		if x.Views == nil {
			x.Views = make(map[integrity.NameDigest]*api.View, 4)
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
