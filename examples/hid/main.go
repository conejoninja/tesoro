package main

import (
	"fmt"

	"github.com/conejoninja/tesoro"
	"github.com/conejoninja/tesoro/shell"
	"github.com/conejoninja/tesoro/transport"
	"github.com/zserge/hid"
)

func main() {
	var client tesoro.Client
	numberDevices := 0

	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()
		if info.Vendor == transport.VendorOne && info.Product == transport.ProductOne && info.Interface == 0 {
			numberDevices++
			var t transport.TransportHID
			t.SetDevice(device)
			client.SetTransport(&t)
			return
		}

	})
	if numberDevices == 0 {
		fmt.Println("No TREZOR devices found, make sure your device is connected")
	} else {
		fmt.Printf("Found %d TREZOR devices connected\n", numberDevices)
		shell.NewShell(&client)
		defer client.CloseTransport()
	}
}
