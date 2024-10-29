package resources

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"iter"
	"log"
	"maps"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/pipes"
	"github.com/wenooij/nuggit/views"
	"gopkg.in/yaml.v3"
)

type Index struct {
	entries       map[integrity.NameDigest]*api.Resource
	entriesByName map[string][]*api.Resource
	views         *views.Index
	pipes         *pipes.Index
	once          sync.Once
}

func (x *Index) Reset() {
	x.entries = make(map[integrity.NameDigest]*api.Resource, 64)
	x.entriesByName = make(map[string][]*api.Resource, 64)
	x.views = new(views.Index)
	x.pipes = new(pipes.Index)
	x.views.Reset()
	x.pipes.Reset()
}

func (x *Index) Pipes() *pipes.Index { return x.pipes }
func (x *Index) Views() *views.Index { return x.views }

func (x *Index) All() iter.Seq2[integrity.NameDigest, *api.Resource] { return maps.All(x.entries) }
func (x *Index) Keys() iter.Seq[integrity.NameDigest]                { return maps.Keys(x.entries) }
func (x *Index) Values() iter.Seq[*api.Resource]                     { return maps.Values(x.entries) }

func (x *Index) Get(nd integrity.NameDigest) (*api.Resource, bool) {
	r, ok := x.entries[nd]
	return r, ok
}

func (x *Index) Add(r *api.Resource) error {
	x.once.Do(x.Reset)
	key := integrity.Key(r)
	x.entries[key] = r
	if x.entriesByName == nil {
		x.entriesByName = make(map[string][]*api.Resource, 64)
	}
	x.entriesByName[key.GetName()] = append(x.entriesByName[key.GetName()], r)
	switch r.GetKind() {
	case api.KindPipe:
		x.pipes.Add(key.GetName(), key.GetDigest(), *r.GetPipe())
		return nil
	case api.KindView:
		if r.GetMetadata().GetUUID() == "" {
			x.views.Add(r.GetName(), r.GetMetadata().GetUUID(), *r.GetView())
			return nil
		}
		x.views.AddName(r.GetName(), *r.GetView())
		return nil
	case api.KindRule: // No rules index (yet).
		return nil
	default:
		return fmt.Errorf("unsupported resource kind (%q)", r.GetKind())
	}
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
