package nuggit

type Action map[string]string

func (a Action) SetAction(action string) bool { return a.Set("action", action) }

func (a Action) Set(key, value string) bool {
	if a == nil {
		return false
	}
	a[key] = value
	return true
}

func (a Action) GetAction() string { return a.GetOrDefaultArg("action") }

func (a Action) GetArg(arg string) (string, bool) {
	v, ok := a[arg]
	return v, ok
}

func (a Action) GetOrDefaultArg(arg string) string {
	v, _ := a[arg]
	return v
}
