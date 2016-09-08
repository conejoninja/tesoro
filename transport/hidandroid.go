package transport

import (
	"encoding/binary"
	"log"
	"math"
	"time"

	"github.com/conejoninja/hid"
)

type TransportHIDAndroid struct {
	device hid.Device
}

func (t *TransportHIDAndroid) SetDevice(device hid.Device) {
	t.device = device
	if err := t.device.Open(); err != nil {
		log.Println("Open error: ", err)
	}
}

func (t *TransportHIDAndroid) Close() {
	t.device.Close()
}

func (t *TransportHIDAndroid) Write(msg []byte) {
	for len(msg) > 0 && t.device != nil {
		blank := make([]byte, 64)
		l := int(math.Min(63, float64(len(msg))))
		tmp := append([]byte{63}, msg[:l]...)
		copy(blank, tmp)
		n, err := t.device.Write(blank, 1*time.Second)

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

func (t *TransportHIDAndroid) Read() ([]byte, uint16, int, error) {
	buf, err := t.device.Read(-1, 100*time.Millisecond)
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
					buf, err = t.device.Read(-1, 100*time.Millisecond)
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
