package protocol

import (
	"fmt"
	"reflect"
)

type ResponseFuture struct {
	c chan any
}

func NewResponseFuture() *ResponseFuture {
	return &ResponseFuture{
		c: make(chan any, 16),
	}
}

func (r *ResponseFuture) Close() {
	close(r.c)
}

func (r *ResponseFuture) ReceiveResponse(resp any) {
	r.c <- resp
}

func (r *ResponseFuture) WaitResponse(resp any) error {
	v, ok := <-r.c
	if !ok {
		return nil
	}
	// Set the value pointed to by resp to v
	switch ptr := resp.(type) {
	case *any:
		*ptr = v
	default:
		rv := reflect.ValueOf(resp)
		if rv.Kind() == reflect.Ptr && rv.Elem().CanSet() {
			rv.Elem().Set(reflect.ValueOf(v))
		} else {
			return fmt.Errorf("needs a pointer")
		}
	}
	return nil
}
