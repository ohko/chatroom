package common

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/ohko/chatroom/config"
)

func H_JSON(no int, data any) *config.ResultData {
	result := config.ResultData{Type: "JSON"}

	bs, err := json.Marshal(&config.JSONData{No: no, Data: data})
	if err != nil {
		result.Error = err
		return &result
	}

	result.Data = bs
	return &result
}

func H_Template(data any, files ...string) *config.ResultData {
	return &config.ResultData{Type: "TEMPLATE", Template: files, Data: data}
}

func H_HTML(data string) *config.ResultData {
	return &config.ResultData{Type: "HTML", Data: data}
}

func H_Redirect(url string) *config.ResultData {
	return &config.ResultData{Type: "Redirect", Data: url}
}

func H_404() *config.ResultData {
	return &config.ResultData{Type: "404"}
}

func GetRealIP(r *http.Request) string {
	if r.Header.Get("Ali-Cdn-Real-Ip") != "" {
		return r.Header.Get("Ali-Cdn-Real-Ip")
	}
	if r.Header.Get("X-Forwarded-For") != "" {
		return r.Header.Get("X-Forwarded-For")
	}
	// ipv6
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	// ipv4
	return strings.Split(r.RemoteAddr, ":")[0]
}

func NewContext() *config.Context {
	return &config.Context{Ctx: context.Background()}
}

func Middleware(f func(ctx *config.Context, w http.ResponseWriter, r *http.Request) *config.ResultData) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			return
		}

		ctx := NewContext()

		var result *config.ResultData
		defer func() {
			if perr := recover(); perr != nil {
				log.Println(perr)
				dep := 0
				for i := 1; i < 10; i++ {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					log.Printf("%s∟ %s:%d\n", strings.Repeat(" ", dep), file, line)
					dep++
				}
			}
		}()

		result = f(ctx, w, r)
		if result == nil {
			return
		}
		switch result.Type {
		case "JSON":
			w.Header().Set("Content-Type", "application/json")
			w.Write(result.Data.([]byte))
		case "HTML":
			w.Header().Add("Content-Type", "text/html; charset=UTF-8")
			w.Header().Add("Content-Length", strconv.Itoa(len(result.Data.(string))))
			w.Write([]byte(result.Data.(string)))
		case "Redirect":
			http.Redirect(w, r, result.Data.(string), http.StatusFound)
		default:
			w.Write([]byte("found error:" + ctx.FlowID))
		}
	}
}

func GenerateNonce(n int) string {
	if n == 0 {
		n = 32
	}
	b := make([]byte, n/2)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return strings.ToUpper(hex.EncodeToString(b))
}

func ReadPostData(ctx *config.Context, r *http.Request, v any) error {
	if r.Body == nil {
		return errors.New("body empty")
	}
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	ctx.PostData = string(bs)

	return json.Unmarshal(bs, v)
}

func Hash(data string) string {
	hash := sha512.Sum512([]byte(data))
	return hex.EncodeToString(hash[:])
}

func Encrypt(plainText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plainText))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plainText)

	return ciphertext, nil
}

func Decrypt(cipherText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(cipherText) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}
