package accounts

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"

	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/scrypt"
	"jumbochain.org/enum"
	"jumbochain.org/filemanagement"
	DB "jumbochain.org/ldb"
	"jumbochain.org/types"

	"github.com/google/uuid"
)

const (
	N      = 262144
	R      = 8
	P      = 1
	keyLen = 32
)

type EncryptedKeyJSONV1 struct {
	Address string     `json:"address"`
	Crypto  CryptoJSON `json:"crypto"`
	Id      string     `json:"id"`
	Version int        `json:"version"`
}

type CryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherparamsJSON       `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type cipherparamsJSON struct {
	IV string `json:"iv"`
}

func keyFileName(keyAddr types.Address) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), hex.EncodeToString(keyAddr[:]))
}

func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}

func EncodePrivateKey(privKey *ecdsa.PrivateKey) string {
	privKeyBytes := privKey.D.Bytes()
	return hex.EncodeToString(privKeyBytes)
}

var mnemonic string

var seed []byte

// Note :- This code can be used to generate mutiple addresses from same memonic
func NewAccount() {

	var auth1 string
	fmt.Println("Enter a auth:") // password for the address/account
	fmt.Scanln(&auth1)
	auth := []byte(auth1)
	id, err := uuid.NewRandom() // create a random id for the account
	if err != nil {
		panic(err)
	}
	salt := make([]byte, 32)
	salt = []byte(hex.EncodeToString(salt))
	derivedKey, err := scrypt.Key(auth, salt, N, R, P, keyLen)
	if err != nil {
		panic(err)
	}
	privateKey := GeneratePrivateKey() // generating private key for user
	privKey := privateKey.key
	ciphertext := EncodePrivateKey(privKey)
	ciphertext = hex.EncodeToString([]byte(ciphertext))
	iv := make([]byte, aes.BlockSize)
	iv = []byte(hex.EncodeToString(iv))
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	publicKey := privateKey.PublicKey() // generating public key from private key
	address := publicKey.Address()      // generating address from private key
	if err != nil {
		panic(err)
	}
	finalAddress := "DH" + address.String()
	finalAddressInByte := []byte(finalAddress)
	DB.AddDataToDatabase1(finalAddressInByte, auth) // adding auth to database
	mac := append(derivedKey[16:32], ciphertext...)
	mac = []byte(hex.EncodeToString(mac))

	sample := map[string]interface{}{
		"address": finalAddress,
		"id":      id,
		"version": 1,
		"crypto": map[string]interface{}{
			"cipher":     "aes-128-ctr",
			"ciphertext": ciphertext,
			"cipherparams": map[string]interface{}{
				"iv": iv,
			},
			"kdf": "scrypt",
			"kdfparams": map[string]interface{}{
				"dklen": keyLen, "n": N,
				"p":    P,
				"r":    R,
				"salt": salt,
			},
			"mac": mac,
		},
	}
	folderPath := "./keystore" // place where keystore file is located
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err = os.Mkdir(folderPath, 0755)
		if err != nil {
			//fmt.Println("Error creating folder:", err)
			return
		}
	}

	path := filepath.Join(folderPath, keyFileName(address))
	forKeystore, err := os.Create(path)
	if err != nil {
		//fmt.Println(err)
	}
	defer forKeystore.Close()

	dataBytes, err := json.MarshalIndent(sample, "", "")
	if err != nil {
		//fmt.Println("marsh error ", err)
	}
	_, err2 := forKeystore.Write(dataBytes)
	if err2 != nil {
		//fmt.Println(err2)
	}
	filemanagement.AppendTofile(string(enum.Keystore), finalAddress)
	fmt.Println("Your Address is: ", finalAddress, "& is stored in keystore folder")

}

func ConvertAddressToBytes(address string) ([]byte, error) {
	addressInBytes := []byte(address)
	return addressInBytes, nil
}

func ConvertBytesToAddress(address []byte) (string, error) {
	addressInString := string(address)
	return addressInString, nil
}

func DecodePrivateKey(privKeyStr string) (*ecdsa.PrivateKey, error) {
	privKeyBytes, err := hex.DecodeString(privKeyStr)
	if err != nil {
		////fmt.Println("error decoding your private key: ", err)
		return nil, err
	}
	curve := elliptic.P256()
	privKey := new(ecdsa.PrivateKey)
	privKey.D = new(big.Int).SetBytes(privKeyBytes)
	privKey.PublicKey.Curve = curve
	privKey.PublicKey.X, privKey.PublicKey.Y = curve.ScalarBaseMult(privKeyBytes)
	return privKey, nil

}

func GetPrivateKeyFromKeystore(addr string, auth string) *ecdsa.PrivateKey {
	add, err := ReadAllFiles(addr)
	if err != nil {
		////fmt.Println("Address not found in folder: ", add)
	}
	file, err := os.Open("./keystore/" + add)
	if err != nil {
		////fmt.Println("File not found: ", file)
	}
	f, err := ioutil.ReadAll(file)
	if err != nil {
		////fmt.Println("Error reading private key")
	}
	defer file.Close()
	var value EncryptedKeyJSONV1
	err = json.Unmarshal(f, &value)
	if err != nil {
		////fmt.Println("Error reading private key")
	}
	address := value.Address
	addressInBytes := []byte(address)
	password, err := DB.FeatchFromDatabase1(addressInBytes)
	if err != nil {
		////fmt.Println(err)
	}
	passwordInstring := string(password)
	var privatekey *ecdsa.PrivateKey
	if passwordInstring == auth {
		ciphertext := value.Crypto.CipherText
		privatekey, err = DecodePrivateKey(ciphertext)
		if err != nil {
			////fmt.Println("Error decoding private key", err)
		}
	}
	return privatekey
}

// func CompareInputWithMac(addr string, userInput string) bool {
// 	add, err := ReadAllFiles(addr)
// 	if err != nil {
// 		////fmt.Println("Address is not valid", err)
// 	}
// 	macFromKeystore := GetPrivateKeyFromKeystore(add)
// 	if addr == macFromKeystore {
// 		////fmt.Println("Account unlocked :", macFromKeystore)
// 	}
// 	////fmt.Println("unlocked :", macFromKeystore)
// 	return true
// }

func ReadAllFiles(addr string) (string, error) {
	var fileName string
	files, err := os.ReadDir("keystore")
	if err != nil {
		////fmt.Println(err)
	}
	for _, f := range files {
		fileName = f.Name()
		if strings.Contains(fileName, addr) {
		}
	}
	return fileName, nil
}

// func GenerateRandomSaltForKeystore() string {
// 	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
// 	b := make([]byte, 12)
// 	var c string
// 	for i := range b {
// 		b[i] = letters[rand.Intn(len(letters))]
// 		c = hex.EncodeToString(b)
// 	}
// 	return string(c)
// }

func Encrypt(text, MySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	buf := make([]byte, 16)
	cfb := cipher.NewCFBEncrypter(block, buf)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText), nil
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Decrypt(text, MySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	cipherText := Decode(text)
	buf := make([]byte, 16)
	cfb := cipher.NewCFBDecrypter(block, buf)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}
