package airConditioner

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

const (
	OFF  = 0
	HEAT = 1
	COOL = 2
)

const (
	MinTemperature     = 22
	MaxTemperature     = 27
	DefaultTemperature = 25
)

const (
	MinSpeed     = 0
	MaxSpeed     = 3
	DefaultSpeed = 0
)

type ACRemoteController struct {
	mode                int
	temperature         int
	fanSpeed            int
	temperatureCallBack func(float64)
	modeCallback        func(int)

	lock   sync.Mutex
	logger *log.Logger
}

func (ac *ACRemoteController) generateCommand() (command string) {
	var m string
	switch ac.mode {
	case OFF:
		command = "OFF"
		return
	case HEAT:
		m = "H"
	case COOL:
		m = "C"
	default:
		m = "C"
	}
	command = fmt.Sprintf("ON_%s_%d_%d", m, ac.temperature, ac.fanSpeed)
	return
}

func sendCommand(command string) (isSucc bool) {
	const PROFILE = "air"
	const UNIXFILE = "/var/run/lirc/lircd"

	logger := log.New(os.Stdout, "LIRC-"+PROFILE, log.Ldate|log.Ltime)
	logger.Println("send command ", command)
	addr, err := net.ResolveUnixAddr("unix", UNIXFILE)
	if err != nil {
		logger.Println("cannot resolve address "+UNIXFILE, err)
		return
	}
	c, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		logger.Println("cannot dial unix socket "+UNIXFILE, err)
		return
	}
	defer c.Close()
	_, err = c.Write([]byte("SEND_ONCE " + PROFILE + " " + command + "\n"))
	if err != nil {
		logger.Println("cannot write command to socket", err)
		return
	}
	buf := make([]byte, 1024)
	nr, err := c.Read(buf)
	if err != nil {
		logger.Println("cannot read the result back", err)
		return
	}
	logger.Println(string(buf[0:nr]))
	isSucc = true
	return
}

func initDefaultACRemoteController() *ACRemoteController {
	return &ACRemoteController{
		mode:        OFF,
		temperature: 0,
		fanSpeed:    0,
		logger:      log.New(os.Stdout, "ACRemoteController", log.Ldate|log.Ltime),
	}
}

func (ac *ACRemoteController) updateTargetTemperature(temp float64) {
	if temp < MinTemperature || temp > MaxTemperature {
		ac.logger.Println("Invalid temperature")
		return
	}

	ac.lock.Lock()
	defer ac.lock.Unlock()

	if ac.mode == OFF {
		ac.mode = COOL
	}

	if ac.temperature == int(temp) {
		return
	}
	ac.temperature = int(temp)
	command := ac.generateCommand()
	if sendCommand(command) == true && ac.modeCallback != nil && ac.temperatureCallBack != nil {
		ac.modeCallback(ac.mode)
		ac.temperatureCallBack(temp)
	}
}

func (ac *ACRemoteController) updateTargetHeatingCoolingState(mode int) {
	if mode < OFF || mode > COOL {
		ac.logger.Println("Invalid mode")
	}

	ac.lock.Lock()
	defer ac.lock.Unlock()

	if ac.mode == mode {
		return
	}

	if ac.temperature < MinTemperature || ac.temperature > MaxTemperature {
		ac.temperature = DefaultTemperature
	}

	ac.mode = mode
	command := ac.generateCommand()
	sendCommand(command)
}

func (ac *ACRemoteController) updateFanSpeed(speed float64) {
	if speed < MinSpeed || speed > MaxSpeed {
		ac.logger.Println("Invalid speed")
	}
	ac.lock.Lock()
	defer ac.lock.Unlock()

	if ac.mode == OFF {
		return
	}

	if ac.fanSpeed == int(speed) {
		return
	}

	ac.fanSpeed = int(speed)
	command := ac.generateCommand()
	sendCommand(command)
}
