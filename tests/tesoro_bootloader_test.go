package tests

import (
	"fmt"
	"testing"

	"encoding/json"

	"crypto/sha256"
	"encoding/hex"
	"github.com/conejoninja/tesoro"
	"github.com/conejoninja/tesoro/pb/messages"
	"github.com/conejoninja/tesoro/tests/common"
	"github.com/conejoninja/tesoro/transport"
	"github.com/zserge/hid"
	"time"
)

var testBLClient tesoro.Client

func init() {
	numberDevices := 0

	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()
		if info.Vendor == transport.VendorOne && info.Product == transport.ProductOne && info.Interface == 0 {
			numberDevices++
			var t transport.TransportHID
			t.SetDevice(device)
			testBLClient.SetTransport(&t)
			return
		}

	})
	if numberDevices == 0 {
		fmt.Println("No TREZOR devices found, make sure your device is connected")
	} else {
		fmt.Printf("Found %d TREZOR devices connected\n", numberDevices)
	}
	// Introduce delay, or it's too fast and it will fail the tests
	time.Sleep(1 * time.Second)
}

func TestBLInitialize(t *testing.T) {

	t.Log("We need to check if device is in bootloader mode.")
	{
		str, msgType := common.Call(testBLClient, testBLClient.Initialize())

		if msgType != 17 {
			t.Errorf("\t\tExpected msgType=17, received %d", msgType)
		} else {
			var ft messages.Features
			err := json.Unmarshal([]byte(str), &ft)
			if err != nil {
				t.Errorf("\t\tError unmarshalling features message: %s", err)
			} else {
				if !ft.GetBootloaderMode() {
					t.Error("\t\tDevice is not in bootloadermode")
				} else {
					t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
				}
			}
		}
	}
}

func TestBLFirmwareUpload(t *testing.T) {

	t.Log("We'll try to upload a new firmware.")
	{
		fw, err := common.ReadFile("firmware.bin")
		if err != nil {
			t.Error("\t\tError reading firmware:", err)
		} else {
			if string(fw[:4]) != "TRZR" {
				t.Error("\t\tNot a TREZOR firmware")
			} else {
				var features messages.Features
				str, msgType := common.Call(testBLClient, testBLClient.Initialize())
				if msgType != 17 {
					t.Error("\t\tError initializing the device")
				} else {
					err := json.Unmarshal([]byte(str), &features)
					if err == nil {
						if features.GetBootloaderMode() != true {
							t.Error("\t\tDevice must be in bootloader mode")
						} else {
							fmt.Println("[WHAT TO DO] Erase firmware, click \"Continue\"")
							str, msgType = common.Call(testBLClient, testBLClient.FirmwareErase())
							if msgType != 2 {
								t.Error("\t\tError erasing previous firmware")
							} else {
								h := sha256.New()
								h.Write([]byte(fw[256:]))
								hash := h.Sum(nil)
								fingerPrint := hex.EncodeToString(hash)
								fmt.Printf("[WHAT TO DO] Check fingerprint match: %s and click \"Continue\" \n", fingerPrint)
								_, msgType = common.Call(testBLClient, testBLClient.FirmwareUpload(fw))
								if msgType != 2 {
									t.Errorf("\t\tExpected msgType=2, received %d", msgType)
								} else {
									t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
								}
							}
						}
					}
				}
			}

		}
	}
}
