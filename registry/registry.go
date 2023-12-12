package registry

type registry struct {
	opts options
}

func NewRegistry(opts ...Options) *registry {
	v := &registry{}

	for _, o := range opts {
		o(&v.opts)
	}

	return v
}

func (r *registry) Nodes() []string {
	return r.opts.nodes
}
func (r *registry) Region() string {
	return r.opts.region
}
func (r *registry) Zone(zone string) string {
	return r.opts.zone
}
func (r *registry) Env(env string) string {
	return r.opts.env
}
func (r *registry) Host(host string) string {
	return r.opts.host
}
