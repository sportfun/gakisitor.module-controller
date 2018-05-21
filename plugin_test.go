package main

import (
	"context"
	"github.com/onsi/gomega"
	"github.com/sportfun/gakisitor/plugin/plugin_test"
	periph "periph.io/x/periph/conn/gpio"
	"testing"
	"time"
)

type testingGPIO struct{}

func (*testingGPIO) register([4]string) error { return nil }
func (*testingGPIO) edge(ctx context.Context) [4]<-chan periph.Level {
	var levels [4]<-chan periph.Level

	levels[0] = make(<-chan periph.Level)
	levels[1] = make(<-chan periph.Level)
	levels[2] = make(<-chan periph.Level)

	out := make(chan periph.Level)
	go func(out chan<- periph.Level) {
		for {
			out <- periph.High
			time.Sleep(50 * time.Millisecond)
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}(out)
	levels[3] = out

	return levels
}

func TestPlugin(t *testing.T) {
	engine.gpio = &testingGPIO{}

	desc := plugin_test.PluginTestDesc{
		ConfigJSON:   `{"gpio":{"pin_0": "NONE", "pin_1": "NONE", "pin_2": "NONE", "pin_3": "NONE"}}`,
		ValueChecker: gomega.Equal(3),
	}
	plugin_test.PluginValidityChecker(t, &Plugin, desc)
}
