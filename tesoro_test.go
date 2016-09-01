package tesoro

import (
	"fmt"
	"testing"

	"github.com/conejoninja/hid"
)

var client Client

func init() {
	numberDevices := 0
	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()
		// TREZOR
		// 0x534c : 21324 vendor
		// 0x0001 : 1     product
		if info.Vendor == 21324 && info.Product == 1 {
			numberDevices++
			_, epOut := device.GetEndpoints()
			if epOut != 1 && epOut != 2 {
				device.SetEpOut(0x01)
			}
			device.SetEpIn(0x81)
			info.Interface = 0x00
			device.SetInfo(info)
			fmt.Println("DEVICE", device)
			client.SetTransport(device)
			return
		}

	})
	if numberDevices == 0 {
		fmt.Println("No TREZOR devices found, make sure your device is connected")
	} else {
		fmt.Printf("Found %d TREZOR devices connected\n", numberDevices)
		//defer client.CloseTransport()
	}
}

func call(msg []byte) (string, uint16) {
	str, msgType := client.Call(msg)

	if msgType == 18 {
		/*fmt.Println(str)
		line, err := prompt.Readline()
		if err != nil {
			fmt.Println("ERR", err)
		}
		str, msgType = call(client.PinMatrixAck(line))*/
	} else if msgType == 26 {
		fmt.Println(str)
		str, msgType = call(client.ButtonAck())
		/*} else if msgType == 41 {
			fmt.Println(str)
			line, err := prompt.Readline()
			if err != nil {
				fmt.Println("ERR", err)
			}
			str, msgType = call(client.PassphraseAck(line))
		} else if msgType == 46 {
			fmt.Println(str)
			line, err := prompt.Readline()
			if err != nil {
				fmt.Println("ERR", err)
			}
			str, msgType = call(client.WordAck(line))*/
	}

	return str, msgType
}

func TestPing(t *testing.T) {

	var expectedPing = "PONG"

	t.Log("We need to test the PING.")
	{
		t.Logf("\tChecking PING for response \"%s\"",
			expectedPing)
		{
			str, msgType := call(client.Ping(expectedPing, false, false, false))

			if msgType != 2 {
				t.Errorf("\t\tExpected msgType=2, received %d", msgType)
			}
			if str != expectedPing {
				t.Errorf("\t\tExpected str=\"%s\", received\"%s\"", expectedPing, str)
			}
			if msgType == 2 && str == expectedPing {
				t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
			}
		}
	}
}

func TestPingButton(t *testing.T) {

	var expectedPing = "PONG"

	t.Log("We need to test the PING.")
	{
		t.Logf("\tChecking PING for response \"%s\"",
			expectedPing)
		{
			str, msgType := call(client.Ping(expectedPing, false, false, true))

			if msgType != 2 {
				t.Errorf("\t\tExpected msgType=2, received %d", msgType)
			}
			if str != expectedPing {
				t.Errorf("\t\tExpected str=\"%s\", received\"%s\"", expectedPing, str)
			}
			if msgType == 2 && str == expectedPing {
				t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
			}
		}
	}
}
