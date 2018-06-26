package core

import (
	"bytes"
	des1 "crypto/des"
	"encoding/base64"
	"errors"
	"strings"

	"tcp.fresh.com/utillib"
)

//数据包加密
func Des(data, k string) string {
	result := utillib.EncryptECB(data, k)
	return strings.TrimSpace(result)
}

//数据包解密
func Undes(result []byte, k string) string {
	data_base, _ := base64.StdEncoding.DecodeString(string(result))
	r := []byte(data_base)
	key := []byte(k)
	origData, err := decrypt(r, key)

	if err != nil {
		panic(err)
	}
	//去掉结果两端的空格，因为在算法中填充的值，计算后产生的非正常值均被替换为空格，故去掉两端空格
	return strings.TrimSpace(string(origData))
}

//ECB PKCS5Padding
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//ECB PKCS5Unpadding
func PKCS5Unpadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//des1 加密
func encrypt(origData1, key []byte) ([]byte, error) {
	///////////////////修改长度，填充byte 0
	block, err := des1.NewCipher(key)
	bs := block.BlockSize()
	len_origData1 := len(origData1) + (len(origData1) % bs)
	for len_origData1%bs != 0 {
		len_origData1 += (len(origData1) % bs)
	}
	origData := make([]byte, len_origData1, len_origData1)

	for k, v := range origData1 {
		origData[k] = v
	}

	i_len := len(origData1) % bs
	len_crypted1_temp := len(origData1)
	for i := 0; i < i_len; i++ {
		origData[len_crypted1_temp+i] = 0
	}

	if len(origData) < 1 || len(key) < 1 {
		return nil, errors.New("wrong data or key")
	}

	if err != nil {
		return nil, err
	}

	if len(origData)%bs != 0 {
		//return nil, errors.New("wrong padding")
	}
	out := make([]byte, len(origData))
	dst := out
	for len(origData) > 0 {
		block.Encrypt(dst, origData[:bs])
		origData = origData[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

//Des解密 ECB
func decrypt(crypted1, key []byte) ([]byte, error) {
	//////////////////修改长度
	block, err := des1.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	len_crypted1 := len(crypted1) + (len(crypted1) % bs)
	crypted := make([]byte, len_crypted1, len_crypted1)
	for k, v := range crypted1 {
		crypted[k] = v
	}

	len_crypted := len(crypted)

	i_len := len(crypted1) % bs
	len_crypted1_temp := len(crypted1)
	for i := 0; i < i_len; i++ {
		crypted[len_crypted1_temp+i] = 0
	}
	////////////////end
	if len(crypted) < 1 || len(key) < 1 {
		return nil, errors.New("wrong data or key")
	}

	out := make([]byte, len_crypted)
	dst := out

	if len_crypted%bs != 0 {

		//return nil, errors.New("wrong crypted size")
	}

	for len(crypted) > 0 {
		block.Decrypt(dst, crypted[:bs])
		crypted = crypted[bs:]
		dst = dst[bs:]
	}

	//由于算法缺陷，这里需要过滤非正常ASCII值
	for k, v := range out {
		//小于等于31的都是控制值，非字符。127是删除符，非普通字符。
		if v <= 31 || v == 127 {
			out[k] = 32
		}
	}
	return out, nil
}

//[golang ECB 3DES Encrypt]
func TripleEcbDesEncrypt(origData, key []byte) ([]byte, error) {
	tkey := make([]byte, 24, 24)
	copy(tkey, key)
	k1 := tkey[:8]
	k2 := tkey[8:16]
	k3 := tkey[16:]

	block, err := des1.NewCipher(k1)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	origData = PKCS5Padding(origData, bs)

	buf1, err := encrypt(origData, k1)
	if err != nil {
		return nil, err
	}
	buf2, err := decrypt(buf1, k2)
	if err != nil {
		return nil, err
	}
	out, err := encrypt(buf2, k3)
	if err != nil {
		return nil, err
	}
	return out, nil
}

//[golang ECB 3DES Decrypt]
func TripleEcbDesDecrypt(crypted, key []byte) ([]byte, error) {
	tkey := make([]byte, 24, 24)
	copy(tkey, key)
	k1 := tkey[:8]
	k2 := tkey[8:16]
	k3 := tkey[16:]
	buf1, err := decrypt(crypted, k3)
	if err != nil {
		return nil, err
	}
	buf2, err := encrypt(buf1, k2)
	if err != nil {
		return nil, err
	}
	out, err := decrypt(buf2, k1)
	if err != nil {
		return nil, err
	}
	out = PKCS5Unpadding(out)
	return out, nil
}
