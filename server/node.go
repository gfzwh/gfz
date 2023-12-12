package server

type nodeopts struct {
	zone string
	env  string
	name string
}

type NodeOption func(*nodeopts)

func Zone(zone string) NodeOption {
	return func(c *nodeopts) {
		c.zone = zone
	}
}

func Env(env string) NodeOption {
	return func(c *nodeopts) {
		c.env = env
	}
}

func Name(name string) NodeOption {
	return func(c *nodeopts) {
		c.name = name
	}
}

type node struct {
	opts nodeopts
}

func Node(opts ...NodeOption) *node {
	n := &node{}
	for _, o := range opts {
		o(&n.opts)
	}

	return n
}
