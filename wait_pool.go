package waitPool

import (
	"context"
	"sync"
)

type WaitPool struct {
	pSize  int
	pSync  chan struct{}
	wSize  int
	waited chan *job
	wg     sync.WaitGroup
	once   sync.Once
}

type job struct {
	work func()
	ctx  context.Context
	done *sync.WaitGroup
}

func New(pSize, wSize int) *WaitPool {
	wChanSize := wSize
	if wChanSize < 0 {
		wChanSize = 10000
	}
	return (&WaitPool{
		pSize:  pSize,
		pSync:  make(chan struct{}, pSize),
		wSize:  wSize,
		waited: make(chan *job, wChanSize),
	}).run()
}

func (wp *WaitPool) Add(ctx context.Context, task func()) *sync.WaitGroup {
	done := &sync.WaitGroup{}
	done.Add(1)
	wp.wg.Add(1)
	j := &job{
		work: task,
		ctx:  ctx,
		done: done,
	}
	if wp.wSize < 0 {
		go wp.add(j)
	} else {
		wp.add(j)
	}
	return done
}

func (wp *WaitPool) add(j *job) {
	wp.waited <- j
}

func (wp *WaitPool) Close() {
	close(wp.waited)
	wp.wg.Wait()
}

func (wp *WaitPool) Len() (int, int) {
	return len(wp.pSync), len(wp.waited)
}

func (wp *WaitPool) run() *WaitPool {
	go wp.once.Do(func() {
		for {
			wp.pSync <- struct{}{}
			for {
				job, ok := <-wp.waited
				if !ok {
					<-wp.pSync
					return
				}
				canceled := false
				select {
				case <-job.ctx.Done():
					canceled = true
					wp.wg.Done()
				default:
					go func() {
						job.work()
						job.done.Done()
						wp.wg.Done()
						<-wp.pSync
					}()
				}
				if !canceled {
					break
				}
			}
		}
	})
	return wp
}
