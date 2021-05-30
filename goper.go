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
	defaultNumCPU   = runtime.NumCPU()
	defaultGoNum    = defaultNumCPU
	defaultChanSize = defaultNumCPU * 64
)

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
	flag uint32    // channal flag indicate goper if running or closed.
	name string    // use for debug
	hd   Handler   // user handler function.
	task chan task // send to handler channal.

	mux sync.Mutex
	wg  sync.WaitGroup
}

func (g *Goper) lazyInit() {
	if g.task == nil {
		g.task = make(chan task, defaultChanSize)
	}
}

// Name register Goper name.
func (g *Goper) Name(name string) { g.name = name }

// String get Goper name.
func (g *Goper) String() string { return g.name }

// Deliver send an arg to handler,
// if Goper not run, it return an error.
func (g *Goper) Deliver(arg interface{}) error {
	return g.put(arg, nil)
}

func (g *Goper) put(arg interface{}, ch chan error) error {
	if g.isStop() {
		return poolError{name: g.String(), reason: errNotRun}
	}
	// BUG channal meby closed by Stop().
	g.task <- task{arg: arg}
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

	if g.task == nil {
		return
	}
	g.stopFlag()

	// TODO wait Put() finish send i.
	// waitting by done rest task
	for len(g.task) > 0 {
		runtime.Gosched()
	} // end waitting Put().

	close(g.task)

	g.wg.Wait()
	g.task = nil
}

// Default run maxGo num of goroutine.
// if maxGo<1, it will use numcpu.
func (g *Goper) Default(maxGo int, hd Handler) error {
	return g.runner(maxGo, hd)
}

func (g *Goper) runner(maxGo int, hd Handler) error {
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

// run num's goroutine
func (g *Goper) shced(maxgo int) {
	for i := 0; i < maxgo; i++ {
		g.goroutine()
	}
}

func (g *Goper) goroutine() {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		for {
			t, ok := <-g.task
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
