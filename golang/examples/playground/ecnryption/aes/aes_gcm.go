package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

// AesGcmEncrypt AES-256 GCM 모드로 암호화 하는 함수입니다. key는 hexadecimal 형식의 문자열이어야 합니다.
// `iv + ciphertext + tag + additional data`를 base64 인코딩하여 리턴합니다
func AesGcmEncrypt(keyInHex string, plaintext string) (ciphertext *string, err error) {
	key, err := hex.DecodeString(keyInHex)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// iv (임의의 값, 12 바이트)
	iv := make([]byte, 12)
	if _, err = rand.Read(iv); err != nil {
		return
	}

	// additional data (임의의 값, 16 바이트)
	additionalData := make([]byte, 16)
	if _, err = rand.Read(additionalData); err != nil {
		return
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	sealed := gcm.Seal(nil, iv, []byte(plaintext), additionalData)

	gcmTagSize := 16
	tagStartIdx := len(sealed) - gcmTagSize

	ciphertextWithoutTag := sealed[:tagStartIdx]
	tag := sealed[tagStartIdx:]

	var toEncode []byte
	toEncode = append(iv, ciphertextWithoutTag...)
	toEncode = append(toEncode, tag...)
	toEncode = append(toEncode, additionalData...)

	base64Dst := make([]byte, base64.StdEncoding.EncodedLen(len(toEncode)))
	base64.StdEncoding.Encode(base64Dst, toEncode)
	ciphertext = Pointer(string(base64Dst))

	return
}

// AesGcmDecrypt AES-256 GCM 모드로 암호화 하는 함수입니다. key는 hexadecimal 형식의 문자열이어야 합니다.
// `base64(iv + ciphertext + tag + additional data)` 형식의 데이터를 받아서 복호화된 평문을 리턴합니다.
func AesGcmDecrypt(keyInHex string, ciphertext []byte) (plaintext *string, err error) {
	key, err := hex.DecodeString(keyInHex)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	base64Dst := make([]byte, base64.StdEncoding.DecodedLen(len(ciphertext)))
	n, err := base64.StdEncoding.Decode(base64Dst, ciphertext)
	if err != nil {
		return nil, err
	}
	base64Dst = base64Dst[:n]

	gcmStandardNonceSize := 12
	iv := base64Dst[:gcmStandardNonceSize]
	encryptedWithTagAndAdd := base64Dst[gcmStandardNonceSize:]

	// `ciphertext + tag`와 `additional data` 추출
	// `tag`는 `gcm.Open` 내에서 추출된다
	addInBitsLen := 128
	addInBytesLen := addInBitsLen / 8

	ciphertextWithTag := encryptedWithTagAndAdd[:len(encryptedWithTagAndAdd)-addInBytesLen]
	additionalData := encryptedWithTagAndAdd[len(ciphertextWithTag):]

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	plaintextBytes, err := gcm.Open(nil, iv, ciphertextWithTag, additionalData)
	if err != nil {
		return
	}

	plaintext = Pointer(string(plaintextBytes))

	return
}

func RevealPointer[T comparable](pointer *T) T {
	if pointer == nil {
		var zero T
		return zero
	}

	return *pointer
}

func Pointer[T any](v T) *T {
	return &v
}
