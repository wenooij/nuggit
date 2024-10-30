package packages

import (
	"sync"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

type Index struct {
	entries            map[integrity.NameDigest]nuggit.Package
	packagesByResource map[integrity.NameDigest]integrity.NameDigest
	resourcesByPackage map[integrity.NameDigest]integrity.NameDigest

	once sync.Once
}

func (i *Index) Reset() {
	i.entries = make(map[integrity.NameDigest]nuggit.Package)
	i.packagesByResource = make(map[integrity.NameDigest]integrity.NameDigest)
	i.resourcesByPackage = make(map[integrity.NameDigest]integrity.NameDigest)
}

func (i *Index) Add(name string, p nuggit.Package) error {
	digest, err := integrity.GetDigest(integrity.DummySpec{Spec: p})
	if err != nil {
		return err
	}
	key := integrity.KeyLit(name, digest)
	i.entries[key] = p
	for k, err := range Keys(p) {
		if err != nil {
			return err
		}
		i.packagesByResource[k] = key
		i.resourcesByPackage[key] = k
	}
	return nil
}
