package gateway

type bufpool struct {
	size int64
}

func (t *bufpool) Get() []byte {
	return make([]byte, t.size)
}

func (t *bufpool) Put([]byte) {
}
