package graphs

// Sort nodes using a topological sort.
func Sort(g *Graph, keys []string) error {
	visitKeys := make(map[string]struct{})
	for _, k := range keys {
		visitKeys[k] = struct{}{}
	}
	res := make([]string, 0, len(keys))
	if err := Visit(g, func(k string) error {
		if _, ok := visitKeys[k]; ok {
			res = append(res, k)
		}
		return nil
	}); err != nil {
		return err
	}
	copy(keys, res)
	return nil
}
