package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sportfun/gakisitor/plugin"
	"github.com/sportfun/gakisitor/profile"
)

var engine = controller{gpio: &gpio{}}
var log = logrus.WithField("plugin", "controller")

var Plugin = plugin.Plugin{
	Name: "Controller Plugin",
	Instance: func(ctx context.Context, profile profile.Plugin, channels plugin.Chan) error {
		// prepare plugin env
		state := plugin.IdleState
		btn, err := configure(profile)
		if err != nil {
			return errors.WithMessage(err, "configuration failed")
		}

		// process
		defer func() { engine.close() }()
		for {
			select {
			case <-ctx.Done():
				return nil

			case instruction, valid := <-channels.Instruction:
				if !valid {
					return nil
				}

				switch instruction {
				case plugin.StatusPluginInstruction:
					channels.Status <- state
				case plugin.StartSessionInstruction:
					if state == plugin.InSessionState {
						break
					}
					state = plugin.InSessionState
					engine.start()
				case plugin.StopSessionInstruction:
					if state == plugin.IdleState {
						break
					}
					state = plugin.IdleState
					engine.stop()
				}

			case btnId, open := <-btn:
				if !open {
					continue
				}

				go func() { channels.Data <- btnId }()
			}
		}
	},
}

func configure(profile profile.Plugin) (<-chan interface{}, error) {
	var btnsGPIO [4]string

	// prepare GPIO
	for idx := range btnsGPIO {
		if btnGPIO, err := profile.AccessTo("gpio", fmt.Sprintf("pin_%d", idx)); err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("gpio.pin_%d", idx))
		} else {
			btnsGPIO[idx] = fmt.Sprint(btnGPIO)
		}
	}

	// configure engine
	return engine.configure(btnsGPIO)
}
