package jabletools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	syncByte = uint8(71) //0x47
)

func decryptAES128(crypted, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func pkcs5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

// Decrypt descryps a segment
func (h *HLS) decrypt(segment *Segment) ([]byte, error) {
	file, err := os.Open(segment.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if segment.Key != nil {
		h.IV = []byte(segment.Key.IV)
		if len(h.IV) == 0 {
			h.IV = defaultIV(segment.SeqId)
		}
		if h.Key == nil || len(h.Key) <= 0 {
			h.Key, h.IV, err = h.getKey(segment)
			if err != nil {
				return nil, err
			}
		}
		data, err = decryptAES128(data, h.Key, h.IV)
		if err != nil {
			return nil, err
		}
	}
	for j := 0; j < len(data); j++ {
		if data[j] == syncByte {
			data = data[j:]
			break
		}
	}

	return data, nil
}

func (h *HLS) getKey(segment *Segment) (key []byte, iv []byte, err error) {
	res, err := h.client.Get(segment.Key.URI)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode != 200 {
		return nil, nil, errors.New("failed to get descryption key")
	}
	key, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	iv = []byte(segment.Key.IV)
	if len(iv) == 0 {
		iv = defaultIV(segment.SeqId)
	}
	return
}

func defaultIV(seqID uint64) []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[8:], seqID)
	return buf
}
