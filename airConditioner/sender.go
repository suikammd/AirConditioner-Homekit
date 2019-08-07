package airConditioner

import (
	"fmt"
	"github.com/brutella/hc/characteristic"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	OFF  = characteristic.CurrentHeatingCoolingStateOff
	HEAT = characteristic.CurrentHeatingCoolingStateHeat
	COOL = characteristic.CurrentHeatingCoolingStateCool
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

const autoPowerOnDuration = 10

type ACRemoteController struct {
	mode                int
	temperature         int
	fanSpeed            int
	temperatureCallBack func(float64)
	modeCallback        func(int)
	lastPowerOff        time.Time

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

func sendCommand(command string) (isSuccessful bool) {
	const Profile = "air"
	const LircdUnixFile = "/var/run/lirc/lircd"

	logger := log.New(os.Stdout, "[Lircd-"+Profile+"]", log.Ldate|log.Ltime)
	logger.Println("send command ", command)
	addr, err := net.ResolveUnixAddr("unix", LircdUnixFile)
	if err != nil {
		logger.Println("cannot resolve address "+LircdUnixFile, err)
		return
	}
	c, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		logger.Println("cannot dial unix socket "+LircdUnixFile, err)
		return
	}
	defer c.Close()
	_, err = c.Write([]byte("SEND_ONCE " + Profile + " " + command + "\n"))
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
	isSuccessful = true
	return
}

func initDefaultACRemoteController() *ACRemoteController {
	return &ACRemoteController{
		mode:         OFF,
		temperature:  0,
		fanSpeed:     0,
		lastPowerOff: time.Now(),
		logger:       log.New(os.Stdout, "[ACRemoteController]", log.Ldate|log.Ltime),
	}
}

func (ac *ACRemoteController) updateTargetTemperature(temp float64) {
	if temp < MinTemperature || temp > MaxTemperature {
		ac.logger.Println("invalid temperature")
		return
	}

	ac.lock.Lock()
	defer ac.lock.Unlock()

	// Change temperature in power-off mode and just powered-off
	if ac.mode == OFF && time.Now().Sub(ac.lastPowerOff) < autoPowerOnDuration*time.Second {
		return
	}

	// Auto power-on when change temperature
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

	if ac.mode == OFF {
		ac.lastPowerOff = time.Now()
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
