package trezor

import (
	"encoding/binary"

	"fmt"

	"github.com/conejoninja/trezor/pb/messages"
	"github.com/conejoninja/trezor/transport"
	"github.com/golang/protobuf/proto"
	"github.com/zserge/hid"
)

type TrezorClient struct {
	t transport.TransportHID
}

func (c *TrezorClient) SetTransport(device hid.Device) {
	c.t.SetDevice(device)
}

func (c *TrezorClient) CloseTransport() {
	c.t.Close()
}

func (c *TrezorClient) Header(msgType int, msg []byte) []byte {

	typebuf := make([]byte, 2)
	binary.BigEndian.PutUint16(typebuf, uint16(msgType))

	msgbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(msgbuf, uint32(len(msg)))

	return append(typebuf, msgbuf...)
}

func (c *TrezorClient) Initialize() {
	var m messages.Initialize
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling", err)
	} else {
		magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_Initialize"]), marshalled)...)
		msg := append(magicHeader, marshalled...)

		c.t.Write(msg)
		fmt.Println(c.Read())
	}
}

func (c *TrezorClient) Ping(str string) {
	var m messages.Ping
	ffalse := false
	m.Message = &str
	m.ButtonProtection = &ffalse
	m.PinProtection = &ffalse
	m.PassphraseProtection = &ffalse
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling", err)
	} else {
		magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_Ping"]), marshalled)...)
		msg := append(magicHeader, marshalled...)

		c.t.Write(msg)
		fmt.Println(c.Read())
	}
}

func (c *TrezorClient) PinMatrixAck(str string) {
	var m messages.PinMatrixAck
	m.Pin = &str
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling", err)
	} else {
		magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_PinMatrixAck"]), marshalled)...)
		msg := append(magicHeader, marshalled...)

		c.t.Write(msg)
		fmt.Println(c.Read())
	}
}

func (c *TrezorClient) GetAddress() {
	ttrue := false
	bitcoin := "Bitcoin"
	var m messages.GetAddress
	//m.AddressN = []uint32{}
	m.CoinName = &bitcoin
	m.ShowDisplay = &ttrue
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling", err)
	} else {
		magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_GetAddress"]), marshalled)...)
		msg := append(magicHeader, marshalled...)

		c.t.Write(msg)
		fmt.Println(c.Read())
	}
}

func (c *TrezorClient) Read() string {
	marshalled, msgType, msgLength, err := c.t.Read()
	if err != nil {
		return "Error reading"
	}
	if msgLength <= 0 {
		return ""
	}

	switch msgType {
	case 2:
		break
	default:
		break
	}

	if msgType == 2 {
		var msg messages.Success
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			return "Error unmarshalling (2)"
		}
		return msg.GetMessage()
	} else if msgType == 3 {
		var msg messages.Failure
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			return "Error unmarshalling (3)"
		}
		return msg.GetMessage()
	} else if msgType == 18 {
		var msg messages.PinMatrixRequest
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			return "Error unmarshalling (18)"
		}
		return "Please enter current PIN:"
	} else if msgType == 30 {
		var msg messages.Address
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			return "Error unmarshalling (18)"
		}
		return msg.GetAddress()
	}
	return fmt.Sprint("Uncaught message type ", msgType)
}
