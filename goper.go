package goper

import (
	"errors"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

const (
	flag_ch_closed = 0
	flag_ch_open   = 1
)

var (
	defaultCPU      = runtime.NumCPU()
	defaultGoNum    = defaultCPU
	defaultChanSize = defaultCPU * 64
)

type empty struct{}

// Handler user handler fuction type.
type Handler func(interface{})

type task struct {
	arg interface{}
}

// Goper a simple half sync half async goroutine pool.
//
// equire to go func().
// Goper is safety to close all goroutine before done
// the task using Put() to channal.
type Goper struct {
	flag uint32 // channal flag indicate goper if running or closed.
	name string // use for debug

	hd     Handler   // user handler function.
	taskCh chan task // send to handler channal.

	mux sync.Mutex
	wg  sync.WaitGroup
}

func (g *Goper) lazyInit() {
	if g.taskCh == nil {
		g.taskCh = make(chan task, defaultChanSize)
	}
}

// Handler user handler function.
// it will call after recive an value
// from channal by Send(v).
func (g *Goper) Handler(hd Handler) { g.hd = hd }

// Name register pool name.
func (g *Goper) Name(name string) { g.name = name }

// String get pool name.
func (g *Goper) String() string { return g.name }

// Send send a value to handle,
// if pool not run,it return an error.
func (g *Goper) Put(arg interface{}) error {
	return g.put(arg, nil)
}

func (g *Goper) put(arg interface{}, ch chan error) error {
	if g.isStop() {
		return poolError{name: g.String(), reason: errNotRun}
	}
	// BUG channal meby closed by Stop().
	g.taskCh <- task{arg: arg}
	return nil
}

// Close() stop runner,
// wait for all work and goroutine finish.
func (g *Goper) Close() {
	if g.isStop() {
		return
	}
	g.mux.Lock()
	defer g.mux.Unlock()

	if g.taskCh == nil {
		return
	}
	g.stopFlag()

	// TODO wait Send() finish send i.
	// waitting by done rest task
	for len(g.taskCh) > 0 {
		runtime.Gosched()
	} // end waitting Send().

	close(g.taskCh)

	g.wg.Wait()
	g.taskCh = nil
}

// Default run maxGo num of goroutine.
// if maxGo<1, it will use numcpu.
func (g *Goper) Default(maxGo int, hd Handler) error {
	return g.run(maxGo, hd)
}

func (g *Goper) run(maxGo int, hd Handler) error {
	if hd == nil {
		return poolError{name: g.String(), reason: errHdNotSet}
	}
	if g.isRun() {
		return nil
	}
	g.mux.Lock()
	defer g.mux.Unlock()

	// dobul check Run() if running.
	if g.isRun() {
		return nil
	}
	g.runFlag()
	g.lazyInit()

	// initialize
	if maxGo < 1 {
		maxGo = defaultGoNum
	}
	g.hd = hd
	g.shced(maxGo)

	return nil
}

func (g *Goper) shced(num int) {
	for i := 0; i < num; i++ {
		g.goroutine()
	}
}

func (g *Goper) goroutine() {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		for {
			t, ok := <-g.taskCh
			if !ok {
				return
			}
			safeCall(g.hd, t.arg, g.String())
		}
	}()
}

func safeCall(f Handler, arg interface{}, n string) {
	defer func() {
		if e := recover(); e != nil {
			err := poolError{name: n, reason: errSafeCall}
			er := errors.New(err.Error() + e.(error).Error())
			log.Println(er.Error())
		}
	}()
	f(arg)
}

func (g *Goper) stopFlag() {
	atomic.StoreUint32(&g.flag, flag_ch_closed)
}

func (g *Goper) runFlag() {
	atomic.StoreUint32(&g.flag, flag_ch_open)
}

func (g *Goper) isRun() bool {
	return atomic.LoadUint32(&g.flag) == flag_ch_open
}

func (g *Goper) isStop() bool {
	return atomic.LoadUint32(&g.flag) == flag_ch_closed
}
