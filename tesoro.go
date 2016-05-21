package tesoro

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"golang.org/x/text/unicode/norm"

	"encoding/base64"

	"strings"

	"regexp"

	"encoding/hex"

	"github.com/conejoninja/tesoro/pb/messages"
	"github.com/conejoninja/tesoro/transport"
	"github.com/golang/protobuf/proto"
	"github.com/zserge/hid"
)

const hardkey uint32 = 2147483648

type Client struct {
	t transport.TransportHID
}

func (c *Client) SetTransport(device hid.Device) {
	c.t.SetDevice(device)
}

func (c *Client) CloseTransport() {
	c.t.Close()
}

func (c *Client) Header(msgType int, msg []byte) []byte {

	typebuf := make([]byte, 2)
	binary.BigEndian.PutUint16(typebuf, uint16(msgType))

	msgbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(msgbuf, uint32(len(msg)))

	return append(typebuf, msgbuf...)
}

func (c *Client) Initialize() []byte {
	var m messages.Initialize
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_Initialize"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) Ping(str string) []byte {
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

func (c *Client) ChangePin() []byte {
	var m messages.ChangePin
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_ChangePin"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) GetEntropy(size uint32) []byte {
	var m messages.GetEntropy
	m.Size = &size
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_GetEntropy"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) GetFeatures() []byte {
	var m messages.GetFeatures
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_GetFeatures"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) PinMatrixAck(str string) []byte {
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

func (c *Client) GetAddress() []byte {
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

func (c *Client) GetPublicKey() []byte {
	var m messages.GetPublicKey
	m.AddressN = StringToBIP32Path("m/44'/0'/0'")
	//m.AddressN = []uint32{hardened(44), hardened(0), hardened(0)} //default key for account #1
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_GetPublicKey"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) SignMessage(message []byte) []byte {
	var m messages.SignMessage
	m.Message = norm.NFC.Bytes(message)
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_SignMessage"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) VerifyMessage(address, signature string, message []byte) []byte {

	sign, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return []byte("Wrong signature")
	}

	var m messages.VerifyMessage
	m.Address = &address
	m.Signature = sign
	m.Message = norm.NFC.Bytes(message)
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_VerifyMessage"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) ButtonAck() []byte {
	var m messages.ButtonAck
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(int(messages.MessageType_value["MessageType_ButtonAck"]), marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) Call(msg []byte) (string, uint16) {
	c.t.Write(msg)
	return c.ReadUntil()
}

func (c *Client) ReadUntil() (string, uint16) {
	var str string
	var msgType uint16
	for {
		str, msgType = c.Read()
		if msgType != 999 { //timeout
			break
		}
	}

	return str, msgType
}

func (c *Client) Read() (string, uint16) {
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
	} else if msgType == 10 {
		var msg messages.Entropy
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (10)"
		} else {
			str = hex.EncodeToString(msg.GetEntropy())
		}
	} else if msgType == 12 {
		var msg messages.PublicKey
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (12)"
		} else {
			str = msg.GetXpub()
		}
	} else if msgType == 17 {
		var msg messages.Features
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (17)"
		} else {
			ftsJSON, _ := json.Marshal(&msg)
			str = string(ftsJSON)
		}
	} else if msgType == 18 {
		var msg messages.PinMatrixRequest
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (18)"
		} else {
			msgSubType := msg.GetType()
			if msgSubType == 1 {
				str = "Please enter current PIN:"
			} else if msgSubType == 2 {
				str = "Please enter new PIN:"
			} else {
				str = "Please re-enter new PIN:"
			}
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
			smJSON, _ := json.Marshal(&msg)
			str = string(smJSON)
		}
	}
	return str, msgType
}

func BIP32Path(keys []uint32) string {
	path := "m"
	for _, key := range keys {
		path += "/"
		if key < hardkey {
			path += string(key)
		} else {

			path += string(key-hardkey) + "'"
		}
	}
	return path
}

func StringToBIP32Path(str string) []uint32 {

	if !ValidBip32(str) {
		return []uint32{}
	}

	re := regexp.MustCompile("([/]+)")
	str = re.ReplaceAllString(str, "/")

	keys := strings.Split(str, "/")
	path := make([]uint32, len(keys)-1)
	for k := 1; k < len(keys); k++ {
		i, _ := strconv.Atoi(strings.Replace(keys[k], "'", "", -1))
		if strings.Contains(keys[k], "'") {
			path[k-1] = hardened(uint32(i))
		} else {
			path[k-1] = uint32(i)
		}
	}
	return path
}

func ValidBip32(path string) bool {
	re := regexp.MustCompile("([/]+)")
	path = re.ReplaceAllString(path, "/")

	re = regexp.MustCompile("^m/")
	path = re.ReplaceAllString(path, "")

	re = regexp.MustCompile("'/")
	path = re.ReplaceAllString(path+"/", "")

	re = regexp.MustCompile("[0-9/]+")
	path = re.ReplaceAllString(path, "")

	return path == ""
}

func hardened(key uint32) uint32 {
	return hardkey + key
}
