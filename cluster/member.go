package cluster

type Member struct {
	ID       string   `gorethink:"id,omitempty"`
	BindPort string   `gorethink:"bind_port,omitempty"`
	BindAddr []string `gorethink:"bind_addr,omitempty"`
}
