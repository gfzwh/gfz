package registry

type Options func(*options)

type options struct {
	url    string
	nodes  []string
	region string
	zone   string
	env    string
	host   string
}

func Url(url string) Options {
	return func(c *options) {
		c.url = url
	}
}

func Nodes(nodes []string) Options {
	return func(c *options) {
		c.nodes = nodes
	}
}
func Region(region string) Options {
	return func(c *options) {
		c.region = region
	}
}
func Zone(zone string) Options {
	return func(c *options) {
		c.zone = zone
	}
}
func Env(env string) Options {
	return func(c *options) {
		c.env = env
	}
}
func Host(host string) Options {
	return func(c *options) {
		c.host = host
	}
}
