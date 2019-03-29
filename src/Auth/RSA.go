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

const defaultRsaLength = 128

func trimHeadTailLine(text string) string {
	tarr := strings.Split(text, "\n")
	return strings.Join(tarr[1:len(tarr)-2], "\n")
}

func rsaPublicFmt(public string) []byte {
	return []byte(
		fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----",
			public))
}

func rsaPrivateFmt(private string) []byte {
	return []byte(
		fmt.Sprintf("-----BEGIN RSA PRIVATE KEY-----\n%s\n-----END RSA PRIVATE KEY-----",
			private))
}

func Decrypt(text string, publicKey string) (string, error) {
	rsaPubKey := rsaPublicFmt(publicKey)
	res, err := rsaDecrypt([]byte(text), rsaPubKey)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func Encrypt(text string, privateKey string) (string, error) {
	rsaPriKey := rsaPrivateFmt(privateKey)
	res, err := rsaEncrypt([]byte(text), rsaPriKey)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func NewKey() (pubilc, private string, err error) {
	puk, prk, err := genRsaKey(defaultRsaLength)
	if err != nil {
		return "", "", err
	}
	return trimHeadTailLine(string(puk)), trimHeadTailLine(string(prk)), nil
}

func rsaEncrypt(origData, publicKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub := pubInterface.(*rsa.PublicKey)
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
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}
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
