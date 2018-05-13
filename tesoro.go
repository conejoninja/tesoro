package tesoro

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/conejoninja/tesoro/pb/messages"
	"github.com/conejoninja/tesoro/pb/types"
	"github.com/conejoninja/tesoro/transport"
	"github.com/golang/protobuf/proto"
	"golang.org/x/text/unicode/norm"
)

const hardkey uint32 = 2147483648

type Client struct {
	t transport.Transport
}

type Storage struct {
	Version string           `json:"version"`
	Config  Config           `json:"config"`
	Tags    map[string]Tag   `json:"tags"`
	Entries map[string]Entry `json:"entries"`
}

type Config struct {
	OrderType string `json:"orderType"`
}

type Tag struct {
	Title  string `json:"title"`
	Icon   string `json:"icon"`
	Active string `json:"active"`
}

type Entry struct {
	Title    string        `json:"title"`
	Username string        `json:"username"`
	Nonce    string        `json:"nonce"`
	Note     string        `json:"note"`
	Password EncryptedData `json:"password"`
	SafeNote EncryptedData `json:"safe_note"`
	Tags     []int         `json:"tags"`
}

type EncryptedData struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

type TxRequest struct {
	Details *types.TxRequestDetailsType `json:"details,omitempty"`
	Type    types.RequestType           `json:"type,omitempty"`
}

func (c *Client) SetTransport(t transport.Transport) {
	c.t = t
}

func (c *Client) CloseTransport() {
	c.t.Close()
}

func (c *Client) Header(msgType messages.MessageType, msg []byte) []byte {

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

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_Initialize, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) Ping(str string, pinProtection, passphraseProtection, buttonProtection bool) []byte {
	var m messages.Ping
	m.Message = &str
	m.ButtonProtection = &buttonProtection
	m.PinProtection = &pinProtection
	m.PassphraseProtection = &passphraseProtection
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_Ping, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) ChangePin() []byte {
	var m messages.ChangePin
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_ChangePin, marshalled)...)
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

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_GetEntropy, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) GetFeatures() []byte {
	var m messages.GetFeatures
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_GetFeatures, marshalled)...)
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

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_PinMatrixAck, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) PassphraseAck(str string) []byte {
	var m messages.PassphraseAck
	m.Passphrase = &str
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_PassphraseAck, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}
func (c *Client) WordAck(str string) []byte {
	var m messages.WordAck
	m.Word = &str
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_WordAck, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) GetAddress(addressN []uint32, showDisplay bool, coinName string) []byte {
	var m messages.GetAddress
	m.AddressN = addressN
	m.CoinName = &coinName
	m.ShowDisplay = &showDisplay
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_GetAddress, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) GetPublicKey(address []uint32) []byte {
	var m messages.GetPublicKey
	m.AddressN = address
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_GetPublicKey, marshalled)...)
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

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_SignMessage, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) SignIdentity(uri string, challengeHidden []byte, challengeVisual string, index uint32) []byte {
	var m messages.SignIdentity
	identity := URIToIdentity(uri)
	identity.Index = &index
	m.Identity = &identity
	m.ChallengeHidden = challengeHidden
	m.ChallengeVisual = &challengeVisual
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_SignIdentity, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) SetLabel(label string) []byte {
	var m messages.ApplySettings
	m.Label = &label
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_ApplySettings, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) WipeDevice() []byte {
	var m messages.WipeDevice
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_WipeDevice, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) EntropyAck(entropy []byte) []byte {
	var m messages.EntropyAck
	m.Entropy = entropy
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_EntropyAck, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) ResetDevice(displayRandom bool, strength uint32, passphraseProtection, pinProtection bool, label string, U2FCounter uint32) []byte {
	var m messages.ResetDevice
	m.DisplayRandom = &displayRandom
	m.Strength = &strength
	m.PassphraseProtection = &passphraseProtection
	m.PinProtection = &pinProtection
	m.U2FCounter = &U2FCounter
	if label != "" {
		m.Label = &label
	}
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_ResetDevice, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) LoadDevice(mnemonic string, passphraseProtection bool, label, pin string, SkipChecksum bool, U2FCounter uint32) []byte {
	var m messages.LoadDevice
	m.Mnemonic = &mnemonic
	m.PassphraseProtection = &passphraseProtection
	if label != "" {
		m.Label = &label
	}
	if pin != "" {
		m.Pin = &pin
	}
	m.SkipChecksum = &SkipChecksum
	m.U2FCounter = &U2FCounter
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_LoadDevice, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) EncryptMessage(pubkey, message string, displayOnly bool, path, coinName string) []byte {
	var m messages.EncryptMessage
	m.Pubkey = []byte(pubkey)
	m.Message = []byte(message)
	m.DisplayOnly = &displayOnly
	m.AddressN = StringToBIP32Path(path)
	m.CoinName = &coinName
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_EncryptMessage, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) DecryptMessage(path string, nonce, message, hmac []byte) []byte {
	var m messages.DecryptMessage
	m.AddressN = StringToBIP32Path(path)
	m.Nonce = nonce
	m.Message = message
	m.Hmac = hmac
	marshalled, err := proto.Marshal(&m)
	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_DecryptMessage, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) RecoveryDevice(wordCount uint32, passphraseProtection, pinProtection bool, label string, EnforceWordList bool, U2FCounter uint32) []byte {
	var m messages.RecoveryDevice
	m.WordCount = &wordCount
	m.PassphraseProtection = &passphraseProtection
	m.PinProtection = &pinProtection
	m.Label = &label
	m.EnforceWordlist = &EnforceWordList
	m.U2FCounter = &U2FCounter

	if label != "" {
		m.Label = &label
	}
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_RecoveryDevice, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) SetHomescreen(homescreen []byte) []byte {
	var m messages.ApplySettings
	m.Homescreen = homescreen
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_ApplySettings, marshalled)...)
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

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_VerifyMessage, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) EstimateTxSize(outputsCount, inputsCount uint32, coinName string) []byte {
	var m messages.EstimateTxSize
	m.OutputsCount = &outputsCount
	m.InputsCount = &inputsCount
	m.CoinName = &coinName
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_EstimateTxSize, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) ButtonAck() []byte {
	var m messages.ButtonAck
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_ButtonAck, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) GetMasterKey() []byte {
	masterKey, _ := hex.DecodeString("2d650551248d792eabf628f451200d7f51cb63e46aadcbb1038aacb05e8c8aee2d650551248d792eabf628f451200d7f51cb63e46aadcbb1038aacb05e8c8aee")
	return c.CipherKeyValue(
		true,
		"Activate TREZOR Password Manager?",
		masterKey,
		StringToBIP32Path("m/10016'/0"),
		[]byte{},
		true,
		true,
	)
}

func (c *Client) GetEntryNonce(title, username, nonce string) []byte {
	return c.CipherKeyValue(
		false,
		"Unlock "+title+" for user "+username+"?",
		[]byte(nonce),
		StringToBIP32Path("m/10016'/0"),
		[]byte{},
		false,
		true,
	)
}

func (c *Client) SetEntryNonce(title, username, nonce string) []byte {
	return c.CipherKeyValue(
		true,
		"Unlock "+title+" for user "+username+"?",
		[]byte(nonce),
		StringToBIP32Path("m/10016'/0"),
		[]byte{},
		false,
		true,
	)
}

func (c *Client) ClearSession() []byte {
	var m messages.ClearSession
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_ClearSession, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) SetU2FCounter(U2FCounter uint32) []byte {
	var m messages.SetU2FCounter
	m.U2FCounter = &U2FCounter
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_SetU2FCounter, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) GetECDHSessionKey(uri string, index uint32, peerPublicKey []byte, ecdsaCurveName string) []byte {
	var m messages.GetECDHSessionKey
	identity := URIToIdentity(uri)
	identity.Index = &index
	m.Identity = &identity
	m.PeerPublicKey = peerPublicKey
	m.EcdsaCurveName = &ecdsaCurveName
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_GetECDHSessionKey, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) FirmwareErase() []byte {
	var m messages.FirmwareErase
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_FirmwareErase, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) FirmwareUpload(payload []byte) []byte {
	var m messages.FirmwareUpload
	m.Payload = payload
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_FirmwareUpload, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) SignTx(outputsCount, inputsCount uint32, coinName string, version, lockTime uint32) []byte {
	var m messages.SignTx
	m.OutputsCount = &outputsCount
	m.InputsCount = &inputsCount
	m.CoinName = &coinName
	if version != 0 {
		m.Version = &version
	}
	if lockTime != 0 {
		m.LockTime = &lockTime
	}
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_SignTx, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) TxAck(tx types.TransactionType) []byte {
	var m messages.TxAck
	m.Tx = &tx
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_TxAck, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) CipherKeyValue(encrypt bool, key string, value []byte, address []uint32, iv []byte, askOnEncrypt, askOnDecrypt bool) []byte {
	var m messages.CipherKeyValue
	m.Key = &key
	if encrypt {
		paddedValue := make([]byte, 16*int(math.Ceil(float64(len(value))/16)))
		copy(paddedValue, value)
		m.Value = paddedValue
	} else {
		var err error
		m.Value, err = hex.DecodeString(string(value))
		if err != nil {
			fmt.Println("ERROR Decoding string")
		}
	}
	m.AddressN = address
	if len(iv) > 0 {
		m.Iv = iv
	}
	m.Encrypt = &encrypt
	m.AskOnEncrypt = &askOnEncrypt
	m.AskOnDecrypt = &askOnDecrypt
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_CipherKeyValue, marshalled)...)
	msg := append(magicHeader, marshalled...)

	return msg
}

func (c *Client) EthereumGetAddress(addressN []uint32, showDisplay bool) []byte {
	var m messages.EthereumGetAddress
	m.AddressN = addressN
	m.ShowDisplay = &showDisplay
	marshalled, err := proto.Marshal(&m)

	if err != nil {
		fmt.Println("ERROR Marshalling")
	}

	magicHeader := append([]byte{35, 35}, c.Header(messages.MessageType_MessageType_EthereumGetAddress, marshalled)...)
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
	marshalled, msgType, _, err := c.t.Read()
	if err != nil {
		return "Error reading", 999
	}

	str := "Uncaught message type " + strconv.Itoa(int(msgType))
	switch messages.MessageType(msgType) {
	case messages.MessageType_MessageType_Success:
		var msg messages.Success
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (2)"
		} else {
			str = msg.GetMessage()
		}
		break
	case messages.MessageType_MessageType_Failure:
		var msg messages.Failure
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (3)"
		} else {
			str = msg.GetMessage()
		}
		break
	case messages.MessageType_MessageType_Entropy:
		var msg messages.Entropy
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (10)"
		} else {
			str = hex.EncodeToString(msg.GetEntropy())
		}
		break
	case messages.MessageType_MessageType_PublicKey:
		var msg messages.PublicKey
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (12)"
		} else {
			smJSON, _ := json.Marshal(&msg)
			str = string(smJSON)
		}
		break
	case messages.MessageType_MessageType_Features:
		var msg messages.Features
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (17)"
		} else {
			ftsJSON, _ := json.Marshal(&msg)
			str = string(ftsJSON)
		}
		break
	case messages.MessageType_MessageType_PinMatrixRequest:
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
		break
	case messages.MessageType_MessageType_TxRequest:
		var msg messages.TxRequest
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (21)"
		} else {
			var txreq TxRequest
			txreq.Details = msg.GetDetails()
			txreq.Type = msg.GetRequestType()
			smJSON, _ := json.Marshal(&msg)
			str = string(smJSON)
		}
		break
	case messages.MessageType_MessageType_ButtonRequest:
		var msg messages.ButtonRequest
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (26)"
		} else {
			str = "Confirm action on TREZOR device"
		}
		break
	case messages.MessageType_MessageType_Address:
		var msg messages.Address
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (30)"
		} else {
			str = msg.GetAddress()
		}
		break
	case messages.MessageType_MessageType_EntropyRequest:
		externalEntropy, _ := GenerateRandomBytes(32)
		str, msgType = c.Call(c.EntropyAck(externalEntropy))
		break
	case messages.MessageType_MessageType_MessageSignature:
		var msg messages.MessageSignature
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (40)"
		} else {
			smJSON, _ := json.Marshal(&msg)
			str = string(smJSON)
		}
		break
	case messages.MessageType_MessageType_PassphraseRequest:
		var msg messages.PassphraseRequest
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (41)"
		} else {
			str = "Enter your passphrase"
		}
		break
	case messages.MessageType_MessageType_TxSize:
		var msg messages.TxSize
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (44)"
		} else {
			str = strconv.Itoa(int(msg.GetTxSize()))
		}
		break
	case messages.MessageType_MessageType_WordRequest:
		var msg messages.WordRequest
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (46)"
		} else {
			str = "Enter the word"
		}
		break
	case messages.MessageType_MessageType_CipheredKeyValue:
		var msg messages.CipheredKeyValue
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (48)"
		} else {
			str = string(msg.GetValue())
		}
		break
	case messages.MessageType_MessageType_EncryptedMessage:
		var msg messages.EncryptedMessage
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (50)"
		} else {
			smJSON, _ := json.Marshal(&msg)
			str = string(smJSON)
		}
		break
	case messages.MessageType_MessageType_DecryptedMessage:
		var msg messages.DecryptedMessage
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (52)"
		} else {
			str = string(msg.GetMessage())
		}
		break
	case messages.MessageType_MessageType_SignedIdentity:
		var msg messages.SignedIdentity
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (54)"
		} else {
			smJSON, _ := json.Marshal(&msg)
			str = string(smJSON)
		}
		break
	case messages.MessageType_MessageType_EthereumAddress:
		var msg messages.EthereumAddress
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (57)"
		} else {
			str = hex.EncodeToString(msg.GetAddress())
		}
		break
	case messages.MessageType_MessageType_ECDHSessionKey:
		var msg messages.ECDHSessionKey
		err = proto.Unmarshal(marshalled, &msg)
		if err != nil {
			str = "Error unmarshalling (62)"
		} else {
			str = string(msg.SessionKey)
		}
		break
	default:
		break
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

	if !ValidBIP32(str) {
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

func ValidBIP32(path string) bool {
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

func PNGToString(filename string) ([]byte, error) {
	img := make([]byte, 1024)
	infile, err := os.Open(filename)
	if err != nil {
		return img, err
	}
	defer infile.Close()

	src, _, err := image.Decode(infile)
	if err != nil {
		return img, err
	}

	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	if w != 128 || h != 64 {
		err = errors.New("Wrong homescreen size")
		return img, err
	}

	imgBin := ""
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			color := src.At(i, j)
			r, g, b, _ := color.RGBA()
			if (r + g + b) > 0 {
				imgBin += "1"
			} else {
				imgBin += "0"
			}
		}
	}
	k := 0
	for i := 0; i < len(imgBin); i += 8 {
		if s, err := strconv.ParseUint(imgBin[i:i+8], 2, 32); err == nil {
			img[k] = byte(s)
			k++
		}
	}
	return img, nil
}

func URIToIdentity(uri string) types.IdentityType {
	var identity types.IdentityType
	u, err := url.Parse(uri)
	if err != nil {
		return identity
	}

	defaultPort := ""
	identity.Proto = &u.Scheme
	user := ""
	if u.User != nil {
		user = u.User.String()
	}
	identity.User = &user
	tmp := strings.Split(u.Host, ":")
	identity.Host = &tmp[0]
	if len(tmp) > 1 {
		identity.Port = &tmp[1]
	} else {
		identity.Port = &defaultPort
	}
	identity.Path = &u.Path
	return identity
}

func hardened(key uint32) uint32 {
	return hardkey + key
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func AES256GCMMEncrypt(plainText, key []byte) ([]byte, []byte) {

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	cipheredText := aesgcm.Seal(nil, nonce, plainText, nil)
	return cipheredText, nonce
}

func AES256GCMDecrypt(cipheredText, key, nonce, tag []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}

	plainText, err := aesgcm.Open(nil, nonce, cipheredText, nil)
	if err != nil {
		return []byte{}, err
	}

	return plainText, nil
}

func GetFileEncKey(masterKey string) (string, string, string) {
	fileKey := masterKey[:len(masterKey)/2]
	encKey := masterKey[len(masterKey)/2:]
	filename_mess := []byte("5f91add3fa1c3c76e90c90a3bd0999e2bd7833d06a483fe884ee60397aca277a")
	mac := hmac.New(sha256.New, []byte(fileKey))
	mac.Write(filename_mess)
	tmpMac := mac.Sum(nil)
	digest := hex.EncodeToString(tmpMac)
	filename := digest + ".pswd"
	return filename, fileKey, encKey
}

func DecryptStorage(content, key string) (Storage, error) {
	cipherKey, _ := hex.DecodeString(key)
	plainText, err := AES256GCMDecrypt([]byte(content[28:]+content[12:28]), cipherKey, []byte(content[:12]), []byte(content[12:28]))

	if err != nil {
		log.Panic("Error decrypting")
	}

	var pc Storage
	fmt.Println(string(plainText))
	err = json.Unmarshal(plainText, &pc)
	return pc, err
}

func DecryptEntry(content, key string) (string, error) {
	cipherKey := []byte(key)
	value, err := AES256GCMDecrypt([]byte(content[28:]+content[12:28]), cipherKey, []byte(content[:12]), []byte(content[12:28]))
	return string(value), err
}

func EncryptEntry(content, key string) []byte {
	ciphered, nonce := AES256GCMMEncrypt([]byte(content), []byte(key))
	cipheredText := string(ciphered)
	l := len(ciphered)
	return []byte(string(nonce) + cipheredText[l-16:] + cipheredText[:l-16])
}

func EncryptStorage(s Storage, key string) []byte {
	cipherKey, _ := hex.DecodeString(key)
	content, err := json.Marshal(s)
	if err != nil {
		log.Panic("Error encrypting")
	}

	ciphered, nonce := AES256GCMMEncrypt(content, cipherKey)
	cipheredText := string(ciphered)
	l := len(ciphered)
	return []byte(string(nonce) + cipheredText[l-16:] + cipheredText[:l-16])
}

// TODO : Work on this
func (e *Entry) Equal(entry Entry) bool {
	if e.Title == entry.Title &&
		e.Username == entry.Username &&
		e.Nonce == entry.Nonce &&
		e.Note == entry.Note &&
		reflect.DeepEqual(e.Password.Data, entry.Password.Data) &&
		e.Password.Type == entry.Password.Type &&
		reflect.DeepEqual(e.SafeNote.Data, entry.SafeNote.Data) &&
		e.SafeNote.Type == entry.SafeNote.Type &&
		reflect.DeepEqual(e.Tags, entry.Tags) {
		return true
	}

	return false
}

// TPM uses []int instead of []byte
func (e EncryptedData) MarshalJSON() ([]byte, error) {

	l := len(e.Data)
	dataInt := make([]int, l)
	for i := 0; i < l; i++ {
		dataInt[i] = int(e.Data[i])
	}
	return json.Marshal(&struct {
		Type string `json:"type"`
		Data []int  `json:"data"`
	}{
		Type: e.Type,
		Data: dataInt,
	})
}
