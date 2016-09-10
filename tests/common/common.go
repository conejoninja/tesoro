package common

import (
	"bufio"
	"os"

	"github.com/conejoninja/tesoro"
)

const Mnemonic12 = "alcohol woman abuse must during monitor noble actual mixed trade anger aisle"
const Mnemonic18 = "owner little vague addict embark decide pink prosper true fork panda embody mixture exchange choose canoe electric jewel"
const Mnemonic24 = "dignity pass list indicate nasty swamp pool script soccer toe leaf photo multiply desk host tomato cradle drill spread actor shine dismiss champion exotic"

const Pin4 = "1234"
const Pin6 = "789456"
const Pin8 = "45678978"

const DefaultPath = "m/44'/0'/0'"
const DefaultCoin = "Bitcoin"

func Call(client tesoro.Client, msg []byte) (string, uint16) {
	str, msgType := client.Call(msg)

	if msgType == 18 {
		/*fmt.Println(str)
		line, err := prompt.Readline()
		if err != nil {
			fmt.Println("ERR", err)
		}
		str, msgType = Call(client, client.PinMatrixAck(line))*/
	} else if msgType == 26 {
		//fmt.Println(str)
		str, msgType = Call(client, client.ButtonAck())
		/*} else if msgType == 41 {
			fmt.Println(str)
			line, err := prompt.Readline()
			if err != nil {
				fmt.Println("ERR", err)
			}
			str, msgType = Call(client, client.PassphraseAck(line))
		} else if msgType == 46 {
			fmt.Println(str)
			line, err := prompt.Readline()
			if err != nil {
				fmt.Println("ERR", err)
			}
			str, msgType = Call(client, client.WordAck(line))*/
	}

	return str, msgType
}

func ReadFile(filename string) ([]byte, error) {
	var empty []byte

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return empty, err
	}

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return empty, statsErr
	}
	var size int64 = stats.Size()
	fw := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(fw)
	return fw, err
}
