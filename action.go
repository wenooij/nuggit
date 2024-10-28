package nuggit

type Action map[string]string

func (a Action) GetSpec() any      { return a }
func (a Action) GetAction() string { return a.GetOrDefaultArg("action") }
func (a Action) GetArg(arg string) (string, bool) {
	v, ok := a[arg]
	return v, ok
}
func (a Action) GetOrDefaultArg(arg string) string {
	v, _ := a[arg]
	return v
}

func (a Action) SetAction(action string) { a.Set("action", action) }
func (a Action) Set(key, value string) {
	if a != nil {
		a[key] = value
	}
}
func (a Action) SetName(name string)     { a.Set("name", name) }
func (a Action) SetDigest(digest string) { a.Set("digest", digest) }
