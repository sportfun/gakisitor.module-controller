package main

import (
	"context"
	"github.com/pkg/errors"
	periph "periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

type gpio struct {
	pins [4]periph.PinIO
}

func (gpio *gpio) register(pins [4]string) error {
	if _, err := host.Init(); err != nil {
		return err
	}

	for idx, pin := range pins {
		gpio.pins[idx] = gpioreg.ByName(pin)
		if gpio.pins[idx] == nil {
			return errors.New("invalid pin: " + pin)
		}
		if err := gpio.pins[idx].In(periph.PullDown, periph.RisingEdge); err != nil {
			return err
		}
	}
	return nil
}

func (gpio *gpio) edge(ctx context.Context) [4]<-chan periph.Level {
	var levels [4]<-chan periph.Level

	for idx, pin := range gpio.pins {
		ch := make(chan periph.Level)
		go func(pin periph.PinIO, levels chan<- periph.Level) {
			for {
				pin.WaitForEdge(-1)
				select {
				case <-ctx.Done():
				case levels <- pin.Read():
				}
			}
		}(pin, ch)
		levels[idx] = ch
	}
	return levels
}
