package registry

type Registry struct {
	opts options
}

func NewRegistry(opts ...Options) *Registry {
	v := &Registry{}

	for _, o := range opts {
		o(&v.opts)
	}

	return v
}

func (r *Registry) Nodes() []string {
	return r.opts.nodes
}
func (r *Registry) Region() string {
	return r.opts.region
}
func (r *Registry) Zone(zone string) string {
	return r.opts.zone
}
func (r *Registry) Env(env string) string {
	return r.opts.env
}
func (r *Registry) Host(host string) string {
	return r.opts.host
}
