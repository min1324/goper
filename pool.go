package goper

import (
	"reflect"
	"sync"
)

// Worker gopool interface.
type Worker interface {
	Close()
	Put(arg interface{}) error
	String() string
}

// New return an Goper using default.
func New(hd Handler) Worker {
	var g Goper
	g.Default(1, hd)
	return &g
}

// Put implicate Worker.
func (g *Goper) Put(arg interface{}) error {
	return g.put(arg)
}

// Pool goper pool
type Pool struct {
	pools sync.Map
}

// Register put Worker into pool.
func (p *Pool) Register(name string, w Worker) error {
	_, loaded := p.pools.LoadOrStore(name, w)
	if loaded {
		return poolError{name: name, reason: nameExistsErr}
	}
	return nil
}

// Default Register a Worker run in default mode.
func (p *Pool) Default(name string, maxgo int, hd Handler) error {
	var g Goper
	g.Name(name)

	err := p.Register(name, &g)
	if err == nil {
		return g.Default(maxgo, hd)
	}
	return err
}

// Get return name's Worker.
func (p *Pool) Get(name string) (t Worker, ok bool) {
	v, loaded := p.pools.Load(name)
	if loaded {
		t, ok = v.(Worker)
		return t, ok
	}
	return nil, false
}

// Put send i to name's worker or group.
// if to group,arg must be Function type.
func (mgr *Pool) Put(name string, arg interface{}) error {
	v, ok := mgr.pools.Load(name)
	if !ok {
		return poolError{name: name, reason: nameNotExistsErr}
	}
	s, _ := v.(Worker)
	f, ok := arg.(Function)
	if ok {
		return s.Put(f)
	}
	return s.Put(arg)
}

// Close stop name's worker.
func (mgr *Pool) Close(name string) {
	v, loaded := mgr.pools.LoadAndDelete(name)
	if loaded {
		s, _ := v.(Worker)
		s.Close()
	}
}

// Shutdown stop all in pool.
func (mgr *Pool) Shutdown() {
	mgr.pools.Range(func(k, v interface{}) bool {
		s, ok := v.(Worker)
		if !ok {
			return true
		}
		s.Close()
		mgr.pools.Delete(k)
		return true
	})
}

// Groud Register a group.
// send to group's arg must be Function.
func (p *Pool) Groud(name string, maxgo int) error {
	var g Goper
	g.Name(name)

	err := p.Register(name, &g)
	if err == nil {
		return g.Default(maxgo, funcHandler)
	}
	return err
}

// GroupPut send i to name's group.
func (mgr *Pool) GroupPut(name string, fn Function) error {
	v, ok := mgr.pools.Load(name)
	if !ok {
		return poolError{name: name, reason: nameNotExistsErr}
	}
	s, _ := v.(Worker)
	return s.Put(fn)
}

// Function a func type for group.
type Function func()

// funcAdapter for group
func funcHandler(i interface{}) {
	f, ok := (i).(func())
	if ok {
		f()
		return
	}
	// can't convert to func,
	// try reflect to call.
	rt := reflect.TypeOf(i)
	if rt.Kind() == reflect.Func {
		if rt.NumIn() == 0 {
			rv := reflect.ValueOf(i)
			rv.Call(nil)
			return
		}
	}
	//panic("invalid function.")
}
