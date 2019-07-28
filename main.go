package main

import (
	"./airConditioner"
	"flag"
)

var storage string
var pin string

func main() {
	flag.StringVar(&storage, "storage", "config", "config path")
	flag.StringVar(&pin, "pin", "23333333", "accessory pin")
	flag.Parse()
	airConditioner.AirConditionerMain(storage, pin)
}
