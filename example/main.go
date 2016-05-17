package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/chzyer/readline"
	"github.com/conejoninja/trezor"
	"github.com/zserge/hid"
)

func main() {

	var c trezor.TrezorClient

	numberDevices := 0
	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()
		// 0x534c : 21324 vendor
		// 0x0001 : 1     product
		if info.Vendor == 21324 && info.Product == 1 {
			numberDevices++
			c.SetTransport(device)
		}
	})
	if numberDevices == 0 {
		fmt.Println("No TREZOR devices found, make sure your TREZOR device is connected")
	} else {
		fmt.Printf("Found %d TREZOR devices connected\n", numberDevices)
		shell(c)
		defer c.CloseTransport()
	}
}

func shell(c trezor.TrezorClient) {
	var msgType uint16
	rl, err := readline.NewEx(&readline.Config{
		Prompt: ">",
	})
	if err != nil {
		panic(err)
	}

	defer rl.Close()
	log.SetOutput(rl.Stderr())

	inputLine := true
	for {
		if inputLine {
			line, err := rl.Readline()
			if err != nil {
				fmt.Println("ERR", err)
				break
			}
			args := strings.Split(strings.ToLower(line), " ")

			switch args[0] {
			case "ping":
				if len(args) < 2 {
					fmt.Println("Missing parameters")
				} else {
					c.Ping(strings.Join(args[1:], " "))
					inputLine = false
				}
				break
			case "signmessage":
				if len(args) < 2 {
					fmt.Println("Missing parameters")
				} else {
					c.SignMessage([]byte(strings.Join(args[1:], " ")))
					inputLine = false
				}
				break
			case "getaddress":
				c.GetAddress()
				inputLine = false
				break
			default:
				if msgType == 18 { // PIN INPUT
					c.PinMatrixAck(line)
					inputLine = false
				} else {
					fmt.Println("Unknown command")
					msgType = 999
				}
				break
			}
		} else {
			var str string
			for {
				str, msgType = c.Read()
				if msgType != 999 { //timeout
					inputLine = true
					switch msgType {
					case 18: // PIN REQUEST
						//inputLine = false
						break
					case 26: // BUTTON REQUEST
						c.ButtonAck()
						inputLine = false
						break
					default:
						break
					}
					break
				}
			}

			if str != "" {
				fmt.Println(str, msgType)
			}

		}
	}
}
