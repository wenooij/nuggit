package pipes

import (
	"iter"
	"maps"
	"slices"
	"sync"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

type Index struct {
	pipes             map[integrity.NameDigest]nuggit.Pipe
	pipeDigestsByName map[integrity.Name]map[string]struct{}
	once              sync.Once
}

func (i *Index) All() iter.Seq2[integrity.NameDigest, nuggit.Pipe] { return maps.All(i.pipes) }
func (i *Index) Keys() iter.Seq[integrity.NameDigest]              { return maps.Keys(i.pipes) }
func (i *Index) Values() iter.Seq[nuggit.Pipe]                     { return maps.Values(i.pipes) }

func (i *Index) Reset() {
	i.pipes = make(map[integrity.NameDigest]nuggit.Pipe)
	i.pipeDigestsByName = make(map[integrity.Name]map[string]struct{})
}

func (i *Index) Remove(name, digest string) {
	i.once.Do(i.Reset)
	delete(i.pipes, integrity.KeyLit(name, digest))
	i.removeDigestByName(name, digest)
}

func (i *Index) removeDigestByName(name, digest string) {
	digests := i.pipeDigestsByName[integrity.KeyLit(name, "")]
	if digests == nil {
		return
	}
	delete(digests, digest)
}

func (i *Index) Add(name, digest string, pipe nuggit.Pipe) {
	i.once.Do(i.Reset)
	key := integrity.KeyLit(name, digest)
	if _, ok := i.pipes[key]; ok {
		return
	}
	i.pipes[key] = pipe
	i.addDigestByName(name, digest)
}

func (i *Index) addDigestByName(name, digest string) {
	key := integrity.KeyLit(name, "")
	digests := i.pipeDigestsByName[key]
	if digests == nil {
		digests = make(map[string]struct{})
	}
	digests[digest] = struct{}{}
	i.pipeDigestsByName[key] = digests
}

func (i *Index) Has(name, digest string) bool {
	_, found := i.pipes[integrity.KeyLit(name, digest)]
	return found
}

func (i *Index) HasName(name string) bool {
	digests := i.pipeDigestsByName[integrity.KeyLit(name, "")]
	return len(digests) > 0
}

func (i *Index) Get(name, digest string) (nuggit.Pipe, bool) {
	pipe, found := i.pipes[integrity.KeyLit(name, digest)]
	return pipe, found
}

func (i *Index) GetUnique(name string) (digest string, ok bool) {
	digests := i.pipeDigestsByName[integrity.KeyLit(name, "")]
	if len(digests) != 1 {
		return "", false
	}
	return slices.Collect(maps.Keys(digests))[0], true
}

func (i *Index) GetUniquePipe(name string) (nuggit.Pipe, bool) {
	digest, ok := i.GetUnique(name)
	if !ok {
		return nuggit.Pipe{}, false
	}
	return i.Get(name, digest)
}
