package encrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testKey = "1159813bdadc416532e0ce953bf38438d7cd735f4fd6945b3abfbb495fff81c1"

func TestAesGcmEncrypt(t *testing.T) {
	plaintext := "62826081"
	encrypted, err := AesGcmEncrypt(testKey, plaintext)
	assert.Nil(t, err)
	assert.NotNil(t, encrypted)

	decrypted, err := AesGcmDecrypt(testKey, []byte(*encrypted))
	assert.Equal(t, plaintext, *decrypted)
}

func TestAesGcmDecrypt(t *testing.T) {
	decrypted, err := AesGcmDecrypt(testKey, []byte("omWQrMZeGMeoUNuilJZ0U4DWa0fAAqBfa1uISGA/U80Ld0NZRPaPSw53FH8whXBHcsIyEA=="))

	assert.Nil(t, err)
	assert.Equal(t, "62826081", *decrypted)
}
