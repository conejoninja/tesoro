package tests

import (
	"fmt"
	"testing"

	"encoding/json"
	"github.com/conejoninja/tesoro"
	"github.com/conejoninja/tesoro/pb/messages"
	"github.com/conejoninja/tesoro/tests/common"
	"github.com/conejoninja/tesoro/transport"
	"github.com/zserge/hid"
)

//var client tesoro.Client

func init() {
	numberDevices := 0
	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()
		// TREZOR
		// 0x534c : 21324 vendor
		// 0x0001 : 1     product
		// 0x00   : Main Trezor Interface
		if info.Vendor == 21324 && info.Product == 1 && info.Interface == 0 {
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
		//defer client.CloseTransport()
	}
}

func TestPing(t *testing.T) {

	var expectedPing = "PONG"

	t.Log("We need to test the PING.")
	{
		t.Logf("\tChecking PING for response \"%s\"",
			expectedPing)
		{
			str, msgType := common.Call(client, client.Ping(expectedPing, false, false, false))

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

	fmt.Println("[WHAT TO DO] Click on \"Confirm\"")
	t.Log("We need to test the PING.")
	{
		t.Logf("\tChecking PING for response \"%s\"",
			expectedPing)
		{
			str, msgType := common.Call(client, client.Ping(expectedPing, false, false, true))

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

func TestPingButtonCancel(t *testing.T) {

	var expectedPing = "PONG"
	var expectedString = "Ping cancelled"

	fmt.Println("[WHAT TO DO] Click on \"Cancel\"")
	t.Log("We need to test the PING.")
	{
		t.Logf("\tChecking PING for response \"%s\"",
			expectedPing)
		{
			str, msgType := common.Call(client, client.Ping(expectedPing, false, false, true))

			if msgType != 3 {
				t.Errorf("\t\tExpected msgType=3, received %d", msgType)
			}
			if str != expectedString {
				t.Errorf("\t\tExpected str=\"%s\", received\"%s\"", expectedString, str)
			}
			if msgType == 3 && str == expectedString {
				t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
			}
		}
	}
}

func TestInitialize(t *testing.T) {

	t.Log("We need to test the Initialize.")
	{
		t.Log("\tChecking Initialize for response ")
		{
			_, msgType := common.Call(client, client.Initialize())

			if msgType != 17 {
				t.Errorf("\t\tExpected msgType=17, received %d", msgType)
			} else {
				t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
			}
		}
	}
}

func TestGetFeatures(t *testing.T) {

	t.Log("We need to test the GetFeatures.")
	{
		t.Log("\tChecking GetFeatures for response ")
		{
			_, msgType := common.Call(client, client.GetFeatures())

			if msgType != 17 {
				t.Errorf("\t\tExpected msgType=17, received %d", msgType)
			} else {
				t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
			}
		}
	}
}

func TestClearSession(t *testing.T) {

	t.Log("We need to test the ClearSession.")
	{
		t.Log("\tChecking ClearSession for response ")
		{
			_, msgType := common.Call(client, client.ClearSession())

			if msgType != 2 {
				t.Errorf("\t\tExpected msgType=2, received %d", msgType)
			} else {
				t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
			}
		}
	}
}

func TestGetEntropy(t *testing.T) {

	fmt.Println("[WHAT TO DO] Click on \"Confirm\"")
	t.Log("We need to test the GetEntropy.")
	{
		t.Log("\tChecking GetEntropy for response ")
		{
			str, msgType := common.Call(client, client.GetEntropy(8))

			if msgType != 10 {
				t.Errorf("\t\tExpected msgType=10, received %d", msgType)
			}

			if len(str) != 16 {
				t.Errorf("\t\tExpected length=16, received %d", len(str))
			}

			if msgType == 10 && len(str) == 16 {
				t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
			}
		}
	}
}

func aTestLoadDevice12(t *testing.T) {

	t.Log("We need to test the LoadDevice.")
	{
		t.Log("\tWe need to wipe it first")
		{
			fmt.Println("[WHAT TO DO] Click on \"Confirm\"")
			_, msgType := common.Call(client, client.WipeDevice())

			if msgType != 2 {
				t.Errorf("\t\tExpected msgType=2, received %d", msgType)
			} else {

				t.Log("\t\tChecking LoadDevice with 12 words")
				{
					fmt.Println("[WHAT TO DO] Click on \"I take the risk\"")
					_, msgType = common.Call(client, client.LoadDevice(common.Mnemonic12, false, "", "", true, 0))
					if msgType != 2 {
						t.Errorf("\t\tExpected msgType=2, received %d", msgType)
					} else {
						str, msgType := common.Call(client, client.GetAddress(tesoro.StringToBIP32Path(common.DefaultPath), false, common.DefaultCoin))
						if msgType != 30 {
							t.Errorf("\t\tExpected msgType=30, received %d", msgType)
						} else {
							if str != "1PbhxNCa4ZGL8NWdFR4ZXrMeunC1yHVspr" {
								t.Errorf("\t\t\tExpected str=1PbhxNCa4ZGL8NWdFR4ZXrMeunC1yHVspr, received %s", str)
							} else {
								t.Log("\t\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
							}
						}
					}
				}
			}
		}
	}
}

func aTestLoadDevice18(t *testing.T) {

	t.Log("We need to test the LoadDevice.")
	{
		t.Log("\tWe need to wipe it first")
		{
			fmt.Println("[WHAT TO DO] Click on \"Confirm\"")
			_, msgType := common.Call(client, client.WipeDevice())

			if msgType != 2 {
				t.Errorf("\t\tExpected msgType=2, received %d", msgType)
			} else {

				t.Log("\t\tChecking LoadDevice with 18 words")
				{
					fmt.Println("[WHAT TO DO] Click on \"I take the risk\"")
					_, msgType = common.Call(client, client.LoadDevice(common.Mnemonic18, false, "", "", true, 0))
					if msgType != 2 {
						t.Errorf("\t\tExpected msgType=2, received %d", msgType)
					} else {
						str, msgType := common.Call(client, client.GetAddress(tesoro.StringToBIP32Path(common.DefaultPath), false, common.DefaultCoin))
						if msgType != 30 {
							t.Errorf("\t\tExpected msgType=30, received %d", msgType)
						} else {
							if str != "1L3KhknA8NTgDmN7bUXQz8PFNSokGSAe1A" {
								t.Errorf("\t\t\tExpected str=1L3KhknA8NTgDmN7bUXQz8PFNSokGSAe1A, received %s", str)
							} else {
								t.Log("\t\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
							}
						}
					}
				}
			}
		}
	}
}

func TestLoadDevice24(t *testing.T) {

	t.Log("We need to test the LoadDevice.")
	{
		t.Log("\tWe need to wipe it first")
		{
			fmt.Println("[WHAT TO DO] Click on \"Confirm\"")
			_, msgType := common.Call(client, client.WipeDevice())

			if msgType != 2 {
				t.Errorf("\t\tExpected msgType=2, received %d", msgType)
			} else {

				t.Log("\tChecking LoadDevice with 24 words")
				{
					fmt.Println("[WHAT TO DO] Click on \"I take the risk\"")
					_, msgType = common.Call(client, client.LoadDevice(common.Mnemonic24, false, "", "", true, 0))
					if msgType != 2 {
						t.Errorf("\t\tExpected msgType=2, received %d", msgType)
					} else {
						str, msgType := common.Call(client, client.GetAddress(tesoro.StringToBIP32Path(common.DefaultPath), false, common.DefaultCoin))
						if msgType != 30 {
							t.Errorf("\t\tExpected msgType=30, received %d", msgType)
						} else {
							if str != "13v1SDrc2qhXT8cgbYa83Nn6ac2jggYgre" {
								t.Errorf("\t\tExpected str=13v1SDrc2qhXT8cgbYa83Nn6ac2jggYgre, received %s", str)
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

func TestSetLabel(t *testing.T) {

	var expectedLabel = "test.LABEL"
	t.Log("We need to test the SetLabel.")
	{
		fmt.Println("[WHAT TO DO] Click on \"Confirm\"")
		str, msgType := common.Call(client, client.SetLabel(expectedLabel))

		if msgType != 2 {
			t.Errorf("\t\tExpected msgType=2, received %d", msgType)
		} else {

			t.Log("\tChecking SetLabel")
			{
				str, msgType = common.Call(client, client.GetFeatures())
				if msgType != 17 {
					t.Error("\t\tError initializing the device")
				} else {
					var features messages.Features
					err := json.Unmarshal([]byte(str), &features)
					if err == nil {
						if features.GetLabel() != expectedLabel {
							t.Errorf("\t\tExpected label=%s, received %s", expectedLabel, features.GetLabel())
						} else {
							t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
						}
					}
				}
			}
		}
	}
}

func TestSetLabel2(t *testing.T) {

	var expectedLabel = "label.TEST"
	t.Log("We need to test the SetLabel.")
	{
		fmt.Println("[WHAT TO DO] Click on \"Confirm\"")
		str, msgType := common.Call(client, client.SetLabel(expectedLabel))

		if msgType != 2 {
			t.Errorf("\t\tExpected msgType=2, received %d", msgType)
		} else {

			t.Log("\tChecking SetLabel")
			{
				str, msgType = common.Call(client, client.GetFeatures())
				if msgType != 17 {
					t.Error("\t\tError initializing the device")
				} else {
					var features messages.Features
					err := json.Unmarshal([]byte(str), &features)
					if err == nil {
						if features.GetLabel() != expectedLabel {
							t.Errorf("\t\tExpected label=%s, received %s", expectedLabel, features.GetLabel())
						} else {
							t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
						}
					}
				}
			}
		}
	}
}

func TestSetHomeScreen(t *testing.T) {

	t.Log("We need to test the SetHomeScreen.")
	{
		hs, err := tesoro.PNGToString("checked.png")
		if err != nil {
			t.Errorf("\t\tError reading homescreen: %s", err)
		}
		fmt.Println("[WHAT TO DO] Click on \"Confirm\"")
		_, msgType := common.Call(client, client.SetHomescreen(hs))
		if msgType != 2 {
			t.Errorf("\t\tExpected msgType=2, received %d", msgType)
		} else {
			t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
		}
	}
}

func TestSetHomeScreen2(t *testing.T) {

	t.Log("We need to test the SetHomeScreen (again).")
	{
		hs, err := tesoro.PNGToString("bunnyhome.png")
		if err != nil {
			t.Errorf("\t\tError reading homescreen: %s", err)
		}
		fmt.Println("[WHAT TO DO] Click on \"Confirm\"")
		_, msgType := common.Call(client, client.SetHomescreen(hs))
		if msgType != 2 {
			t.Errorf("\t\tExpected msgType=2, received %d", msgType)
		} else {
			t.Log("\t\tEverything went fine, \\ʕ◔ϖ◔ʔ/ YAY!")
		}
	}
}
