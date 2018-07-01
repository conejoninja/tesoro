package transport

import (
	"encoding/binary"
	"math"

	"fmt"

	"github.com/gotmc/libusb"
)

type TransportLibUSB struct {
	device       *libusb.Device
	deviceHandle *libusb.DeviceHandle
}

func (t *TransportLibUSB) SetDevice(deviceHandle *libusb.DeviceHandle) {
	t.deviceHandle = deviceHandle
	t.deviceHandle.DetachKernelDriver(0)
	err := t.deviceHandle.ClaimInterface(0) // TODO change it
	if err != nil {
		t.deviceHandle = nil
	}

}

func (t *TransportLibUSB) Close() {
	//t.deviceHandle.Close()
}

func (t *TransportLibUSB) Write(msg []byte) {
	for len(msg) > 0 && t.deviceHandle != nil {
		blank := make([]byte, 64)
		l := int(math.Min(63, float64(len(msg))))
		tmp := append([]byte{63}, msg[:l]...)
		copy(blank, tmp)
		n, err := t.deviceHandle.BulkTransferOut(1, blank, 1000)

		if err == nil && n > 0 {
			if len(msg) < 64 {
				break
			} else {
				msg = msg[63:]
			}
		} else {
			fmt.Println("ERROR WRITTING", n, err)
			break
		}
	}
}

func (t *TransportLibUSB) Read() ([]byte, uint16, int, error) {
	buf, _, err := t.deviceHandle.BulkTransferIn(129, 64, 1000)
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
					buf, _, err = t.deviceHandle.BulkTransferIn(129, 64, 10)
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
