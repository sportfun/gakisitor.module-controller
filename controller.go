package main

import (
	"context"
	periph "periph.io/x/periph/conn/gpio"
)

type GPIO interface {
	register([4]string) error
	edge(context.Context) [4]<-chan periph.Level
}

type controller struct {
	context.Context
	gpio GPIO

	io   chan interface{}
	kill chan interface{}
}

func (rpm *controller) configure(pin [4]string) (<-chan interface{}, error) {
	if err := rpm.gpio.register(pin); err != nil {
		return nil, err
	}
	rpm.io = make(chan interface{})
	return rpm.io, nil
}

func (rpm *controller) start() {
	rpm.kill = make(chan interface{})

	go func(kill <-chan interface{}) {

		ctx, cancel := context.WithCancel(context.Background())
		edges := rpm.gpio.edge(ctx)
		for {
			select {
			case <-kill:
				cancel()
				return

			case isPushed := <-edges[0]:
				if isPushed {
					rpm.io <- 0
				}
			case isPushed := <-edges[1]:
				if isPushed {
					rpm.io <- 1
				}
			case isPushed := <-edges[2]:
				if isPushed {
					rpm.io <- 2
				}
			case isPushed := <-edges[3]:
				if isPushed {
					rpm.io <- 3
				}
			}
		}
	}(rpm.kill)

	log.Debug("start session")
}

func (rpm *controller) stop() {
	safelyClose(&rpm.kill)
	log.Debug("stop session")
}

func (rpm *controller) close() {
	safelyClose(&rpm.kill)
	safelyClose(&rpm.io)
	log.Debug("exit plugin")
}

func safelyClose(c *chan interface{}) {
	if *c != nil {
		close(*c)
	}
	*c = nil
}
