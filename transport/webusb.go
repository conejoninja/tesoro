package transport

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"

	"github.com/trezor/trezord-go/usb/lowlevel"
)

const (
	webusbPrefix  = "web"
	webConfigNum  = 1
	webIfaceNum   = 0
	webAltSetting = 0
	webEpIn       = 0x81
	webEpOut      = 0x01
	usbTimeout    = 5000
)

type TransportWebUSB struct {
	device        lowlevel.Device_Handle
	closed        int32
	transferMutex sync.Mutex
}

func (t *TransportWebUSB) SetDevice(device lowlevel.Device) {
	d, err := lowlevel.Open(device)
	if err != nil {
		log.Fatal("Could not open WebUSB", err)
	}

	err = lowlevel.Reset_Device(d)
	if err != nil {
		// don't abort if reset fails
		// lowlevel.Close(d)
		// return nil, err
	}

	currConf, err := lowlevel.Get_Configuration(d)
	if err != nil {
		log.Fatalf("webusb - connect - current configuration err %s", err.Error())
	} else {
		fmt.Printf("webusb - connect - current configuration %d\n", currConf)
	}

	err = lowlevel.Set_Configuration(d, webConfigNum)
	if err != nil {
		// don't abort if set configuration fails
		// lowlevel.Close(d)
		// return nil, err
		fmt.Printf("Warning: error at configuration set: %s\n", err)
	}

	currConf, err = lowlevel.Get_Configuration(d)
	if err != nil {
		fmt.Printf("webusb - connect - current configuration err %s\n", err.Error())
	} else {
		fmt.Printf("webusb - connect - current configuration %d\n", currConf)
	}

	err = lowlevel.Claim_Interface(d, webIfaceNum)
	if err != nil {
		lowlevel.Close(d)
		log.Fatal("webusb - connect - claiming interface failed", err)
	}

	t.device = d
}

func (t *TransportWebUSB) Close() {
	atomic.StoreInt32(&t.closed, 1)
	t.finishReadQueue()
	t.transferMutex.Lock()
	lowlevel.Close(t.device)
	t.transferMutex.Unlock()
}

func (t *TransportWebUSB) finishReadQueue() {
	t.transferMutex.Lock()
	var err error
	var buf [64]byte

	for err == nil {
		_, err = lowlevel.Interrupt_Transfer(t.device, webEpIn, buf[:], 50)
	}
	t.transferMutex.Unlock()
}

func (t *TransportWebUSB) readWrite(buf []byte, endpoint uint8) (int, error) {
	for {
		closed := (atomic.LoadInt32(&t.closed)) == 1
		if closed {
			return 0, errors.New("closed device")
		}

		t.transferMutex.Lock()
		p, err := lowlevel.Interrupt_Transfer(t.device, endpoint, buf, usbTimeout)
		t.transferMutex.Unlock()

		if err == nil {
			if len(p) > 0 {
				return len(p), err
			}
		}

		if err != nil {
			if err.Error() == lowlevel.Error_Name(int(lowlevel.ERROR_IO)) ||
				err.Error() == lowlevel.Error_Name(int(lowlevel.ERROR_NO_DEVICE)) {
				return 0, errors.New("device disconnected during action")
			}

			if err.Error() != lowlevel.Error_Name(int(lowlevel.ERROR_TIMEOUT)) {
				return 0, err
			}
		}
	}
}

func (t *TransportWebUSB) Write(msg []byte) {
	for len(msg) > 0 && t.device != nil {
		blank := make([]byte, 64)
		l := int(math.Min(63, float64(len(msg))))
		tmp := append([]byte{63}, msg[:l]...)
		copy(blank, tmp)
		n, err := t.readWrite(blank, webEpOut)

		if err == nil && n > 0 {
			if len(msg) < 64 {
				break
			} else {
				msg = msg[63:]
			}
		} else {
			break
		}
	}
}

func (t *TransportWebUSB) Read() ([]byte, uint16, int, error) {
	buf := make([]byte, 64)
	_, err := t.readWrite(buf, webEpIn)
	var marshalled []byte

	bufLength := len(buf)
	for i := 0; i < bufLength; i++ {
		// 35 : '#' magic header
		if buf[i] == 35 && buf[i+1] == 35 {
			msgType := binary.BigEndian.Uint16(buf[i+2 : i+4])
			msgLength := int(binary.BigEndian.Uint32(buf[i+4 : i+8]))
			originalMsgLength := msgLength

			if (bufLength - i - 8) < msgLength {
				marshalled = buf[i+8:]
				msgLength = msgLength - (len(buf) - i - 8)
				for msgLength > 0 {
					buf = make([]byte, 64)
					_, err = t.readWrite(buf, webEpIn)
					bufLength = len(buf)
					if bufLength > 0 {
						l := int(math.Min(float64(bufLength-1), float64(msgLength)))
						marshalled = append(marshalled, buf[1:l+1]...)
						msgLength = msgLength - l
					}
				}
			} else {
				marshalled = buf[i+8 : i+8+msgLength]
			}

			return marshalled, msgType, originalMsgLength, nil
		}
	}
	return marshalled, 999, 0, err
}
