package try

import (
	"errors"
	"runtime/debug"
)

type tryCatch struct {
	tryf   func() error
	catchf func(error, string)
}

func (t *tryCatch) run() {
	defer func() {
		if r := recover(); r != nil {
			t.catchf(errors.New(r.(string)), string(debug.Stack()))
		}
	}()

	if e := t.tryf(); e != nil {
		t.catchf(e, string(debug.Stack()))
	}
}
