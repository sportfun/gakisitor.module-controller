package main

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/sportfun/gakisitor/config"
	"github.com/sportfun/gakisitor/env"
	"github.com/sportfun/gakisitor/log"
	"github.com/sportfun/gakisitor/module"
)

type plugin struct {
	logger        log.Logger
	notifications *module.NotificationQueue
	state         byte

	shutdown []chan interface{}
	gpio     []string
}

const tick = 50 * time.Millisecond

var (
	debugModuleStarted    = log.NewArgumentBinder("module '%s' started")
	debugModuleConfigured = log.NewArgumentBinder("module '%s' configured")
	debugSessionStarted   = log.NewArgumentBinder("session started")
	debugSessionStopped   = log.NewArgumentBinder("session stopped")
	debugModuleStopped    = log.NewArgumentBinder("module '%s' stopped")

	warnSessionNotStarted = log.NewArgumentBinder("session not started")
)

func init() {
	if _, err := host.Init(); err != nil {
		panic(err)
	}
}

func (p *plugin) Start(q *module.NotificationQueue, l log.Logger) error {
	if q == nil {
		p.state = env.PanicState
		return fmt.Errorf("notification queue is not set")
	}
	if l == nil {
		p.state = env.PanicState
		return fmt.Errorf("logger is not set")
	}

	p.logger = l
	p.notifications = q
	p.state = env.StartedState

	p.shutdown = []chan interface{}{make(chan interface{}), make(chan interface{}), make(chan interface{}), make(chan interface{})}
	p.gpio = make([]string, 4)

	l.Debug(debugModuleStarted.Bind(p.Name()))
	return nil
}

func loadConfigurationItem(items map[string]interface{}, name string) (string, error) {
	_, ok := items[name]
	if !ok {
		return "", fmt.Errorf("invalid value of '%s' in configuration", name)
	}

	v, ok := items[name].(string)
	if !ok {
		return "", fmt.Errorf("invalid value of '%s' in configuration", name)
	}

	return v, nil
}

func (p *plugin) Configure(properties *config.ModuleDefinition) error {
	if properties.Config == nil {
		p.state = env.PanicState
		return fmt.Errorf("configuration needed for this module. RTFM")
	}

	items, ok := properties.Config.(map[string]interface{})
	if !ok {
		p.state = env.PanicState
		return fmt.Errorf("valid configuration needed for this module. RTFM")
	}

	var err error
	for k, v := range map[string]*string{
		"gpio.button1": &p.gpio[0],
		"gpio.button2": &p.gpio[1],
		"gpio.button3": &p.gpio[2],
		"gpio.button4": &p.gpio[3],
	} {
		if *v, err = loadConfigurationItem(items, k); err != nil {
			p.state = env.PanicState
			return err
		}
	}

	p.logger.Debug(debugModuleConfigured.Bind(p.Name()))
	p.state = env.IdleState
	return nil
}

func (p *plugin) Process() error { return nil }

func (p *plugin) Stop() error {
	if p.state == env.WorkingState {
		p.StopSession()
	}

	p.logger.Debug(debugModuleStopped.Bind(p.Name()))
	p.state = env.StoppedState
	return nil
}

func readPin(name string, id int, shutdown <-chan interface{}, notification *module.NotificationQueue) {
	pin := gpioreg.ByName(name)
	if pin != nil {
		pin.In(gpio.PullNoChange, gpio.BothEdges)
		last := gpio.Low

		for {
			select {
			case <-shutdown:
				return
			default:
				pin.WaitForEdge(-1)

				valueReaded := pin.Read()
				if last != valueReaded {
					last = valueReaded
					if last == gpio.High {
						notification.NotifyData("Controller", "%d", id)
					}
				}
			}
		}

	}

}

func (p *plugin) StartSession() error {
	if p.state == env.WorkingState {
		p.StopSession()
		return fmt.Errorf("session already exist")
	}

	p.logger.Debug(debugSessionStarted)

	for id, name := range p.gpio {
		go readPin(name, id, p.shutdown[id], p.notifications)
	}

	p.state = env.WorkingState
	return nil
}

func (p *plugin) StopSession() error {
	if p.state != env.WorkingState {
		p.state = env.IdleState
		return fmt.Errorf("session not started")
	}

	for id := range p.gpio {
		p.shutdown[id] <- "nique ta race"
	}

	p.logger.Debug(debugSessionStopped)
	p.state = env.IdleState
	return nil
}

func (p *plugin) Name() string { return "Controller" }
func (p *plugin) State() byte  { return p.state }

//ExportModule export the controller plugin
func ExportModule() module.Module {
	return &plugin{}
}

// Fix issue #20312 (https://github.com/golang/go/issues/20312)
func main() {}
