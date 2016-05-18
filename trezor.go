package trezor

import (
	"encoding/binary"

	"strconv"

	"fmt"

	"encoding/json"

	"github.com/conejoninja/trezor/pb/messages"
	"github.com/conejoninja/trezor/transport"
	"github.com/golang/protobuf/proto"
	"github.com/zserge/hid"
)

type SignMessage struct {
	Message   string `json:"message"`
	Address   string `json:"address"`
	Signature []byte `json:"signature"`
}

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

func (c *TrezorClient) Initialize() []byte {
	var m messages.Initialize
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_Initialize"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *TrezorClient) Ping(str string) []byte {
	var m messages.Ping
	ffalse := false
	m.Message = &str
	m.ButtonProtection = &ffalse
	m.PinProtection = &ffalse
	m.PassphraseProtection = &ffalse
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_Ping"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *TrezorClient) PinMatrixAck(str string) []byte {
	var m messages.PinMatrixAck
	m.Pin = &str
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_PinMatrixAck"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *TrezorClient) GetAddress() []byte {
	ttrue := false
	bitcoin := "Bitcoin"
	var m messages.GetAddress
	//m.AddressN = []uint32{}
	m.CoinName = &bitcoin
	m.ShowDisplay = &ttrue
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_GetAddress"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *TrezorClient) SignMessage(message []byte) []byte {
	bitcoin := "Bitcoin"
	var m messages.SignMessage
	//m.AddressN = []uint32{}
	m.CoinName = &bitcoin
	m.Message = message
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_SignMessage"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *TrezorClient) ButtonAck() []byte {
	var m messages.ButtonAck
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_ButtonAck"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *TrezorClient) Call(msg []byte) (string, uint16) {
	c.t.Write(msg)
	return c.ReadUntil()
}

func (c *TrezorClient) ReadUntil() (string, uint16) {
	var str string
	var msgType uint16
	for {
		str, msgType = c.Read()
		if msgType != 999 { //timeout
			fmt.Println("READUNTIL", msgType)
			break
		}
	}

	return str, msgType
}

func (c *TrezorClient) Read() (string, uint16) {
	marshalled, msgType, msgLength, err := c.t.Read()
	if err != nil {
		//fmt.Println(err)
		return "Error reading", 999
	}
	if msgLength <= 0 {
		fmt.Println("Empty message", msgType)
		return "", msgType
	}

	str := "Uncaught message type " + strconv.Itoa(int(msgType))
	if msgType == 2 {
		var msg messages.Success
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (2)"
		} else {
			str = msg.GetMessage()
		}
	} else if msgType == 3 {
		var msg messages.Failure
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (3)"
		} else {
			str = msg.GetMessage()
		}
	} else if msgType == 18 {
		var msg messages.PinMatrixRequest
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (18)"
		} else {
			str = "Please enter current PIN:"
		}
	} else if msgType == 26 {
		var msg messages.ButtonRequest
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (26)"
		} else {
			str = "Action required on TREZOR device"
		}
	} else if msgType == 30 {
		var msg messages.Address
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (30)"
		} else {
			str = msg.GetAddress()
		}
	} else if msgType == 40 {
		var msg messages.MessageSignature
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (40)"
		} else {
			sm := &SignMessage{
				Address:   msg.GetAddress(),
				Signature: msg.GetSignature()}
			smJSON, _ := json.Marshal(sm)
			str = string(smJSON)
		}
	}
	return str, msgType
}
