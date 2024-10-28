package pipes

import (
	"sync"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

type Index struct {
	pipes             map[integrity.NameDigest]nuggit.Pipe
	pipeDigestsByName map[integrity.Name][]string
	once              sync.Once
}

func (i *Index) Reset() {
	i.pipes = make(map[integrity.NameDigest]nuggit.Pipe)
	i.pipeDigestsByName = make(map[integrity.Name][]string)
}

func (i *Index) Add(name, digest string, pipe nuggit.Pipe) {
	i.once.Do(i.Reset)
	key := integrity.KeyLit(name, digest)
	if _, ok := i.pipes[key]; ok {
		return
	}
	i.pipes[key] = pipe
	key = integrity.KeyLit(name, "")
	i.pipeDigestsByName[key] = append(i.pipeDigestsByName[key], digest)
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
	return digests[0], true
}
