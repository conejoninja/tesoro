package main

import (
	"encoding/hex"
	"fmt"

	"github.com/conejoninja/tesoro"
	"github.com/conejoninja/tesoro/shell"
	"github.com/conejoninja/tesoro/transport"
	"github.com/trezor/trezord-go/usb/lowlevel"
)

func main() {
	var client tesoro.Client
	numberDevices := 0

	var usbctx lowlevel.Context

	list, err := lowlevel.Get_Device_List(usbctx)

	if err != nil {
		fmt.Println("Get_Device_List error", list, err)
	}

	defer func() {
		lowlevel.Free_Device_List(list, 1) // unlink devices
	}()

	paths := make(map[string]bool)

	for _, dev := range list {

		// MATCH

		c, err := lowlevel.Get_Active_Config_Descriptor(dev)
		if err != nil {
			fmt.Println("webusb - match - error getting config descriptor " + err.Error())
		}
		match := c.BNumInterfaces > 0 &&
			c.Interface[0].Num_altsetting > 0 &&
			c.Interface[0].Altsetting[0].BInterfaceClass == lowlevel.CLASS_VENDOR_SPEC

		// END MATCH

		if match {
			dd, err := lowlevel.Get_Device_Descriptor(dev)
			if err != nil {
				continue
			}

			trezorOne := dd.IdVendor == transport.VendorOne && dd.IdProduct == transport.ProductOne
			trezorT := dd.IdVendor == transport.VendorT && dd.IdProduct == transport.ProductT

			if trezorOne || trezorT {

				path := ""
				var ports [8]byte
				p, err := lowlevel.Get_Port_Numbers(dev, ports[:])
				if err == nil {
					path = "web" + hex.EncodeToString(p)
				}
				inset := paths[path]
				if !inset {
					paths[path] = true
					numberDevices++
					var t transport.TransportWebUSB
					t.SetDevice(dev)
					client.SetTransport(&t)
					break
				}
			}
		}
	}

	if numberDevices == 0 {
		fmt.Println("No TREZOR devices found, make sure your device is connected")
	} else {
		fmt.Printf("Found %d TREZOR devices connected\n", numberDevices)
		shell.NewShell(&client)
		defer client.CloseTransport()
	}

}
