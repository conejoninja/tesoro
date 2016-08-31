package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"encoding/json"

	"encoding/base64"

	"github.com/chzyer/readline"
	"github.com/conejoninja/tesoro"
	"github.com/conejoninja/tesoro/pb/messages"
	"github.com/conejoninja/hid"
)

var client tesoro.Client
var prompt *readline.Instance

func main() {
	numberDevices := 0
	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()
		// TREZOR
		// 0x534c : 21324 vendor
		// 0x0001 : 1     product
		if info.Vendor == 21324 && info.Product == 1 {
			numberDevices++
			_, epOut := device.GetEndpoints()
			if epOut!=1 && epOut!=2 {
				device.SetEpOut(0x01)
			}
			device.SetEpIn(0x81)
			info.Interface = 0x00
			device.SetInfo(info)
			fmt.Println("DEVICE", device)
			client.SetTransport(device)
			return
		}

	})
	if numberDevices == 0 {
		fmt.Println("No TREZOR devices found, make sure your device is connected")
	} else {
		fmt.Printf("Found %d TREZOR devices connected\n", numberDevices)
		shell()
		defer client.CloseTransport()
	}
}

func call(msg []byte) (string, uint16) {
	str, msgType := client.Call(msg)

	if msgType == 18 {
		fmt.Println(str)
		line, err := prompt.Readline()
		if err != nil {
			fmt.Println("ERR", err)
		}
		str, msgType = call(client.PinMatrixAck(line))
	} else if msgType == 26 {
		fmt.Println(str)
		str, msgType = call(client.ButtonAck())
	} else if msgType == 41 {
		fmt.Println(str)
		line, err := prompt.Readline()
		if err != nil {
			fmt.Println("ERR", err)
		}
		str, msgType = call(client.PassphraseAck(line))
	} else if msgType == 46 {
		fmt.Println(str)
		line, err := prompt.Readline()
		if err != nil {
			fmt.Println("ERR", err)
		}
		str, msgType = call(client.WordAck(line))
	}

	return str, msgType
}

func shell() {
	var str string
	var msgType uint16
	rl, err := readline.NewEx(&readline.Config{
		Prompt: ">",
	})
	prompt = rl
	if err != nil {
		panic(err)
	}

	defer rl.Close()
	log.SetOutput(rl.Stderr())

	for {
		line, err := rl.Readline()
		if err != nil {
			fmt.Println("ERR", err)
			break
		}
		args := strings.Split(line, " ")

		switch strings.ToLower(args[0]) {
		case "ping":
			if len(args) < 2 {
				fmt.Println("Missing parameters")
			} else {
				pinProtection := false
				buttonProtection := false
				passphraseProtection := false

				if len(args) >= 3 {
					if args[2] == "1" || args[2] == "true" {
						pinProtection = true
					}
				}
				if len(args) >= 4 {
					if args[3] == "1" || args[3] == "true" {
						passphraseProtection = true
					}
				}
				if len(args) >= 5 {
					if args[4] == "1" || args[4] == "true" {
						buttonProtection = true
					}
				}

				str, msgType = call(client.Ping(args[1], pinProtection, passphraseProtection, buttonProtection))
			}
			break
		case "signmessage":
			if len(args) < 2 {
				fmt.Println("Missing parameters")
			} else {
				msg := strings.Join(args[1:], " ")
				str, msgType = call(client.SignMessage([]byte(msg)))
			}
			break
		case "verifymessage":
			if len(args) < 4 {
				fmt.Println("Missing parameters")
			} else {
				str, msgType = call(client.VerifyMessage(args[1], args[2], []byte(args[3])))
			}
			break
		case "getaddress":
			var path string
			showDisplay := false
			coinName := "Bitcoin"
			if len(args) < 2 {
				path = "m/44'/0'/0'"
			} else {
				path = args[1]
			}
			if len(args) >= 3 {
				if args[2] == "1" || args[2] == "true" {
					showDisplay = true
				}
			}
			if len(args) >= 4 {
				coinName = args[3]
			}

			str, msgType = call(client.GetAddress(tesoro.StringToBIP32Path(path), showDisplay, coinName))
			break
		case "encryptmessage":
			if len(args) < 3 {
				fmt.Println("Missing parameters")
			} else {
				path := "m/44'/0'/0'"
				displayOnly := false
				coinName := "Bitcoin"
				pubkey, errHex := hex.DecodeString(args[1])
				if errHex == nil {
					message := args[2]
					if len(args) >= 4 {
						if args[3] == "1" || args[3] == "true" {
							displayOnly = true
						}
					}
					if len(args) >= 5 {
						path = args[4]
					}
					if len(args) >= 6 {
						coinName = args[5]
					}
					str, msgType = call(client.EncryptMessage(string(pubkey), message, displayOnly, path, coinName))
					var encrypted messages.EncryptedMessage
					err := json.Unmarshal([]byte(str), &encrypted)
					if err == nil {
						str = base64.StdEncoding.EncodeToString([]byte(string(encrypted.Nonce) + string(encrypted.Message) + string(encrypted.Hmac)))
					} else {
						str = "Error in data"
					}

				} else {
					fmt.Println("Public key has to be hexadecimal", string(pubkey), errHex)
				}
			}
			break
		case "decryptmessage":
			if len(args) < 3 {
				fmt.Println("Missing parameters")
			} else {
				decoded, err := base64.StdEncoding.DecodeString(args[2])
				if err == nil {
					l := len(decoded)
					str, msgType = call(client.DecryptMessage(args[1], decoded[:33], decoded[33:l-8], decoded[l-8:]))
				} else {
					fmt.Println("Not a valid payload")
				}
			}
			break
		case "getentropy":
			if len(args) < 2 {
				fmt.Println("Missing parameters")
			} else {
				size, _ := strconv.Atoi(args[1])
				str, msgType = call(client.GetEntropy(uint32(size)))
			}
			break
		case "setlabel":
			if len(args) < 2 {
				fmt.Println("Missing parameters")
			} else {
				str, msgType = call(client.SetLabel(strings.Join(args[1:], " ")))
			}
			break
		case "initialize":
		case "init":
			str, msgType = call(client.Initialize())
			break
		case "firmwareerase":
			str, msgType = call(client.FirmwareErase())
			break
		case "wipedevice":
			str, msgType = call(client.WipeDevice())
			break
		case "resetdevice":
			//displayRandom bool, strength uint32, passphraseProtection, pinProtection bool, label string
			displayRandom := false
			if len(args) > 1 {
				if args[1] == "1" || args[1] == "true" {
					displayRandom = true
				}
			}
			var strength uint32
			strength = 256
			if len(args) > 2 {
				if args[2] == "128" || args[2] == "196" {
					i, ierr := strconv.Atoi(args[2])
					if ierr == nil {
						strength = uint32(i)
					}
				}
			}
			passphraseProtection := false
			if len(args) > 3 {
				if args[3] == "1" || args[3] == "true" {
					passphraseProtection = true
				}
			}
			pinProtection := false
			if len(args) > 4 {
				if args[4] == "1" || args[4] == "true" {
					pinProtection = true
				}
			}
			label := ""
			if len(args) > 5 {
				label = args[5]
			}
			str, msgType = call(client.ResetDevice(displayRandom, strength, passphraseProtection, pinProtection, label))
			break
		case "loaddevice":
			l := len(args)
			wordCount := 0
			if l >= 13 && l <= 16 {
				wordCount = 13
			} else if l >= 19 && l <= 22 {
				wordCount = 19
			} else if l >= 25 {
				wordCount = 25
			}
			if wordCount == 0 {
				fmt.Println("Wrong number of parameters")
			} else {
				mnemonic := strings.Join(args[1:wordCount], " ")
				passphraseProtection := false
				if l >= wordCount+1 && (args[wordCount] == "1" || args[wordCount] == "true") {
					passphraseProtection = true
				}
				var label string
				if l >= wordCount+2 {
					label = args[wordCount+1]
				}
				var pin string
				if l >= wordCount+3 {
					pin = args[wordCount+2]
				}
				str, msgType = call(client.LoadDevice(mnemonic, passphraseProtection, label, pin))
			}
			break
		case "recoverydevice":
			l := len(args)
			if l < 2 {
				fmt.Println("Wrong number of parameters")
			} else {
				var wordCount uint32
				i, _ := strconv.Atoi(args[1])
				wordCount = uint32(i)
				if wordCount == 12 || wordCount == 18 || wordCount == 24 {
					passphraseProtection := false
					if l >= 3 {
						if args[2] == "1" || args[2] == "true" {
							passphraseProtection = true
						}
					}
					pinProtection := false
					if l >= 4 {
						if args[3] == "1" || args[3] == "true" {
							pinProtection = true
						}
					}
					var label string
					if l == 5 {
						label = args[4]
					}
					str, msgType = call(client.RecoveryDevice(wordCount, passphraseProtection, pinProtection, label))
				} else {
					fmt.Println("Invalid word count. Use 12/18/24")
				}
			}
			break
		case "sethomescreen":
			if len(args) < 2 {
				fmt.Println("Missing parameters")
			} else {
				homescreen, err := tesoro.PNGToString(args[1])
				if err != nil {
					fmt.Println("Error reading image")
				} else {
					str, msgType = call(client.SetHomescreen(homescreen))
				}
			}
			break
		case "fu":
		//case "firmwareupload":
			if len(args) < 2 {
				fmt.Println("Missing parameters")
			} else {
				fw, err := readFile(args[1])
				if err != nil {
					fmt.Println("Error reading firmware:", err)
				} else {
					if string(fw[:4]) != "TRZR" {
						fmt.Println("Not a TREZOR firmware")
					} else {
						var features messages.Features
						str, msgType = call(client.Initialize())
						if msgType != 17 {
							fmt.Println("Error initializing the device")
						} else {
							err := json.Unmarshal([]byte(str), &features)
							if err == nil {
								if features.GetBootloaderMode() != true {
									fmt.Println("Device must be in bootloader mode")
								} else {
									str, msgType = call(client.FirmwareErase())
									if msgType != 2 {
										fmt.Println("Error erasing previous firmware")
									} else {
										h := sha256.New()
										h.Write([]byte(fw[256:]))
										hash := h.Sum(nil)
										fingerPrint := hex.EncodeToString(hash)
										fmt.Println("Fingerprint:", fingerPrint)
										str, msgType = call(client.FirmwareUpload(fw))
									}
								}
							}
						}
					}

				}
			}
			break
		case "getpublickey":
			var path string
			if len(args) < 2 {
				path = "m/44'/0'/0'"
			} else {
				path = args[1]
			}

			if !tesoro.ValidBIP32(path) {
				fmt.Println("Invalid BIP32 path. Example: m/44'/0'/0'/0/27 ")
			} else {
				str, msgType = call(client.GetPublicKey(tesoro.StringToBIP32Path(path)))
				var node messages.PublicKey
				err := json.Unmarshal([]byte(str), &node)
				if err == nil {
					str = node.GetXpub()
				}
			}
			break
		case "getnode":
			var path string
			if len(args) < 2 {
				path = "m/44'/0'/0'"
			} else {
				path = args[1]
			}

			if !tesoro.ValidBIP32(path) {
				fmt.Println("Invalid BIP32 path. Example: m/44'/0'/0'/0/27 ")
			} else {
				str, msgType = call(client.GetPublicKey(tesoro.StringToBIP32Path(path)))
				var node messages.PublicKey
				err := json.Unmarshal([]byte(str), &node)
				if err == nil {
					smJSON, _ := json.Marshal(node.GetNode())
					str = string(smJSON)
				}
			}
			break
		case "signidentity":
			var index uint32
			if len(args) < 4 {
				fmt.Println("Missing parameters")
			} else {
				if len(args) >= 5 {
					i, _ := strconv.Atoi(args[4])
					index = uint32(i)
				}
				str, msgType = call(client.SignIdentity(args[1], []byte(args[2]), args[3], index))
			}
			break
		case "getfeatures":
			str, msgType = call(client.GetFeatures())
			break
		case "clearsession":
			str, msgType = call(client.ClearSession())
			break
		case "changepin":
			str, msgType = call(client.ChangePin())
			break
		case "cipherkeyvalue":
			var path string
			var iv []byte
			encrypt := true
			askOnEncode := true
			askOnDecode := true
			if len(args) < 4 {
				fmt.Println("Missing parameters")
			} else {
				if args[1] == "0" || args[1] == "false" {
					encrypt = false
				}
				if len(args) < 5 {
					path = "m/44'/0'/0'"
				} else {
					path = args[4]
				}
				if len(args) >= 6 {
					iv = []byte(args[5])
				}
				if len(args) >= 7 && (args[6] == "0" || args[6] == "false") {
					askOnEncode = false
				}
				if len(args) >= 8 && (args[7] == "0" || args[7] == "false") {
					askOnDecode = false
				}
				if !tesoro.ValidBIP32(path) {
					fmt.Println("Invalid BIP32 path. Example: m/44'/0'/0'/0/27 ")
				} else {
					str, msgType = call(client.CipherKeyValue(encrypt, args[2], []byte(args[3]), tesoro.StringToBIP32Path(path), iv, askOnEncode, askOnDecode))
				}
			}
			break
		case "pswdmanager":
		case "pm":
			// GET MASTER KEY
			str, msgType = call(client.GetMasterKey())
			if msgType == 48 {
				masterKey := hex.EncodeToString([]byte(str))
				filename, _, encKey := tesoro.GetFileEncKey(masterKey)

				// OPEN FILE
				contentByte, err := readFile("./" + filename)
				content := string(contentByte)
				if err == nil {
					// DECRYPT STORAGE
					data, err := tesoro.DecryptStorage(content, encKey)
					printStorage(data)

					// Read entry to decrypt
					line, err := rl.Readline()
					if err != nil {
						fmt.Println("ERR", err)
						break
					}
					args = strings.Split(line, " ")
					if _, ok := data.Entries[args[0]]; ok {
						str, msgType = call(client.GetEntryNonce(data.Entries[args[0]].Title, data.Entries[args[0]].Username, data.Entries[args[0]].Nonce))
						pswd, _ := tesoro.DecryptEntry(string(data.Entries[args[0]].Password.Data), str)
						note, _ := tesoro.DecryptEntry(string(data.Entries[args[0]].SafeNote.Data), str)
						if len(pswd) > 2 {
							fmt.Println("Password:", pswd[1:len(pswd)-1])
						} else {
							fmt.Println("Password:")
						}
						if len(note) > 2 {
							fmt.Println("Safe note:", note[1:len(note)-1])
						} else {
							fmt.Println("Safe note:")
						}
					} else {
						fmt.Println("Selected entry does not exists")
					}
					str = ""
				} else {
					str = "Error opening file " + filename
				}

			}
			break
		case "pswdexample": // Insert random entry as an example
		case "pe":
			// GET MASTER KEY
			str, msgType = call(client.GetMasterKey())
			if msgType == 48 {
				masterKey := hex.EncodeToString([]byte(str))
				filename, _, encKey := tesoro.GetFileEncKey(masterKey)

				// OPEN FILE
				contentByte, err := readFile("./" + filename)
				content := string(contentByte)
				if err == nil {
					// DECRYPT STORAGE
					data, _ := tesoro.DecryptStorage(content, encKey)

					var entry tesoro.Entry

					rndByte, _ := tesoro.GenerateRandomBytes(3)
					rnd := hex.EncodeToString(rndByte)

					entry.Title = "Some Service " + rnd
					entry.Username = "MyUsername" + rnd
					entry.Note = "My normal note " + rnd
					nonceByte, _ := tesoro.GenerateRandomBytes(32)
					nonce := string(nonceByte)
					entry.Tags = []int{1}
					var eNonce string
					eNonce, msgType = call(client.SetEntryNonce(entry.Title, entry.Username, nonce))
					entry.Nonce = hex.EncodeToString([]byte(eNonce))
					entry.Password = tesoro.EncryptedData{"Buffer", tesoro.EncryptEntry("\"MySecretPassword"+rnd+"\"", nonce)}
					entry.SafeNote = tesoro.EncryptedData{"Buffer", tesoro.EncryptEntry("\"My Safe Note is safe "+rnd+"\"", nonce)}

					max := 0
					for k, _ := range data.Entries {
						i, e := strconv.Atoi(k)
						if e == nil && i > max {
							max = i
						}
					}

					lastEntry := strconv.Itoa(max + 1)

					data.Entries[lastEntry] = entry
					fmt.Printf("Added entry #%s\n", lastEntry)
					efile := tesoro.EncryptStorage(data, encKey)
					ioutil.WriteFile("./"+filename, efile, 0644)
					str = ""
				} else {
					str = "Error opening file " + filename
				}

			}
			break
		case "pswdremove": // Remove entry from the list
		case "pr":
			// GET MASTER KEY
			str, msgType = call(client.GetMasterKey())
			if msgType == 48 {
				masterKey := hex.EncodeToString([]byte(str))
				filename, _, encKey := tesoro.GetFileEncKey(masterKey)

				// OPEN FILE
				contentByte, err := readFile("./" + filename)
				content := string(contentByte)
				if err == nil {
					// DECRYPT STORAGE
					data, err := tesoro.DecryptStorage(content, encKey)
					printStorage(data)

					// Read entry to decrypt
					line, err := rl.Readline()
					if err != nil {
						fmt.Println("ERR", err)
						break
					}
					args = strings.Split(line, " ")
					if _, ok := data.Entries[args[0]]; ok {
						delete(data.Entries, args[0])
						efile := tesoro.EncryptStorage(data, encKey)
						ioutil.WriteFile("./"+filename, efile, 0644)
						fmt.Printf("Deleted entry #%s\n", args[0])
					} else {
						fmt.Println("Selected entry does not exists")
					}
					str = ""
				} else {
					str = "Error opening file " + filename
				}

			}
			break
		default:
			fmt.Println("Unknown command")
			str = line
			msgType = 999
			break
		}
		if str != "" {
			fmt.Println(str, msgType)
		}
	}
}

func printStorage(s tesoro.Storage) {
	fmt.Println("Password Entries")
	fmt.Println("================")
	fmt.Println("")

	for id, e := range s.Entries {
		printEntry(id, e)
	}

	fmt.Println("")
	fmt.Println("Select entry number to decrypt: ")

}

func printEntry(id string, e tesoro.Entry) {
	fmt.Printf("Entry id: #%s\n", id)
	for i := 0; i < (11 + len(id)); i++ {
		fmt.Print("-")
	}
	fmt.Println("")
	fmt.Println("* username : ", e.Username)
	fmt.Println("* tags : ", e.Tags)
	fmt.Println("* title : ", e.Title)
	fmt.Println("* note : ", e.Note)
	fmt.Println("")
}

func readFile(filename string) ([]byte, error) {
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
