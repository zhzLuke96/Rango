package Auth

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

const defaultRsaLength = 1024
const RSAgap = "<##RSA##>"

func trimHeadTailLine(text string) string {
	tarr := strings.Split(text, "\n")
	return strings.Join(tarr[1:len(tarr)-2], "\n")
}

func rsaPublicFmt(key string) []byte {
	return []byte(
		fmt.Sprintf("-----BEGIN RSA PRIVATE KEY-----\n%s\n-----END RSA PRIVATE KEY-----",
			key))
}

func rsaPrivateFmt(key string) []byte {
	return []byte(
		fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----",
			key))
}

func Decrypt(ciphertext string, decodekey string) (string, error) {
	cipherArr := strings.Split(ciphertext, RSAgap)
	ret := ""
	for _, v := range cipherArr {
		enc, err := doDcrypt(v, decodekey)
		if err != nil {
			return "", err
		}
		ret += enc
	}
	return ret, nil
}

func doDcrypt(ciphertext string, decodekey string) (string, error) {
	rsaPriKey := rsaPublicFmt(decodekey)
	res, err := rsaDecrypt([]byte(ciphertext), rsaPriKey)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func Encrypt(text string, encodekey string) (string, error) {
	pubsize := defaultRsaLength / 16
	if len(text) >= pubsize-11 {
		tempt := text[:pubsize-12]
		next := text[pubsize-12:]
		ten, _ := Encrypt(tempt, encodekey)
		nt, _ := Encrypt(next, encodekey)
		return ten + RSAgap + nt, nil
	}
	rsaPubKey := rsaPrivateFmt(encodekey)
	res, err := rsaEncrypt([]byte(text), rsaPubKey)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func NewKey() (decodekey, encodekey string, err error) {
	prk, puk, err := genRsaKey(defaultRsaLength)
	if err != nil {
		return "", "", err
	}
	// return trimHeadTailLine(string(puk)), trimHeadTailLine(string(prk)), nil
	return trimHeadTailLine(string(prk)), trimHeadTailLine(string(puk)), nil
}

func rsaEncrypt(origData, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}

	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

func rsaDecrypt(ciphertext, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

func genRsaKey(bits int) (prik []byte, pubk []byte, err error) {
	prikey := bytes.NewBuffer(nil)
	pubkey := bytes.NewBuffer(nil)

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	err = pem.Encode(pubkey, block)
	if err != nil {
		return nil, nil, err
	}

	publicKey := &privateKey.PublicKey
	derPkix := x509.MarshalPKCS1PublicKey(publicKey)
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	err = pem.Encode(prikey, block)
	if err != nil {
		return nil, nil, err
	}
	return pubkey.Bytes(), prikey.Bytes(), nil
}
