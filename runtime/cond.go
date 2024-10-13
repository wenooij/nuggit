package runtime

import (
	"net/url"
	"regexp"

	"github.com/wenooij/nuggit"
)

type runCond struct {
	AlwaysEnabled bool
	Hosts         map[string]struct{}
	Pattern       *regexp.Regexp
}

func newRunCond(rc nuggit.RunCondition) (*runCond, error) {
	if rc == (nuggit.RunCondition{}) {
		return nil, nil
	}
	hosts := make(map[string]struct{})
	if rc.Host != "" {
		hosts[rc.Host] = struct{}{}
	}
	c := &runCond{
		AlwaysEnabled: rc.AlwaysEnabled,
		Hosts:         hosts,
	}
	if pattern := rc.URLPattern; pattern != "" {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		c.Pattern = r
	}
	return c, nil
}

func (c runCond) Test(u *url.URL) bool {
	if c.AlwaysEnabled {
		return true
	}
	hostname := u.Hostname()
	if _, ok := c.Hosts[hostname]; ok {
		return true
	}
	return c.Pattern.MatchString(u.String())
}
