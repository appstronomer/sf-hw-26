package pipe

import "sync"

type Pipe[T any] struct {
	chNext <-chan T
	wg     sync.WaitGroup
}

func (p *Pipe[T]) check() {
	if p.chNext == nil {
		panic("called builder method on finished pipe")
	}
}

func (p *Pipe[T]) Next(chOutCap int, dataProcLoop func(chIn <-chan T, chOut chan<- T)) *Pipe[T] {
	p.check()
	chIn := p.chNext
	chOut := make(chan T, chOutCap)
	p.chNext = chOut
	go func(chInArg <-chan T, chOutArg chan<- T) {
		defer close(chOutArg)
		dataProcLoop(chInArg, chOutArg)
	}(chIn, chOut)
	return p
}

func (p *Pipe[T]) Last(dataDstLoop func(chIn <-chan T)) {
	p.check()
	p.wg.Add(1)
	go func(wg *sync.WaitGroup, chInArg <-chan T) {
		defer wg.Done()
		dataDstLoop(chInArg)
	}(&p.wg, p.chNext)
	p.chNext = nil
}

func (p *Pipe[T]) Wait() {
	if p.chNext != nil {
		panic("called Wait method on unfinished pipe")
	}
	p.wg.Wait()
}

func NewPipe[T any](chOutCap int, dataSrcLoop func(chOut chan<- T)) *Pipe[T] {
	chOut := make(chan T, chOutCap)
	pipe := &Pipe[T]{chNext: chOut, wg: sync.WaitGroup{}}
	go func(chOutArg chan<- T) {
		defer close(chOutArg)
		dataSrcLoop(chOutArg)
	}(chOut)
	return pipe
}
