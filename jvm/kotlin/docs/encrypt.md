# Encrypt

- [Encrypt](#encrypt)
    - [`SecretKeyFactory` or `KeyGenerator`](#secretkeyfactory-or-keygenerator)
    - [String to Hex, Hex to String](#string-to-hex-hex-to-string)
    - [crypt](#crypt)
        - [encrypt \& decrypt](#encrypt--decrypt)

## `SecretKeyFactory` or `KeyGenerator`

- [Generating a Secure AES Key in Java](https://www.baeldung.com/java-secure-aes-key)
- [Engine Classes and Algorithms](https://docs.oracle.com/javase/8/docs/technotes/guides/security/crypto/CryptoSpec.html#Engine)
- `SecretKeyFactory`:
    - used to convert existing opaque cryptographic keys of type SecretKey into key specifications
    - 평문 키가 있다면 해당 키를 `SecretKey` 오브젝트로 만드는 데 사용
- `KeyGenerator`:
    - used to generate a new pair of public and private keys suitable for use with a specified algorithm.
    - 새로운 키를 생성해야 하는 경우
- `Cipher.UNWRAP_MODE`: 암호화된 키가 있다면 `Cipher` 통해서 unwrap

## String to Hex, Hex to String

- [Convert Byte Arrays to Hex Strings in Kotlin](https://www.baeldung.com/kotlin/int-to-hex-string)
- [Kotlin convert hex string to ByteArray](https://stackoverflow.com/questions/66613717/kotlin-convert-hex-string-to-bytearray)

## crypt

- [Crypto.kt](https://github.com/tlaukkan/kui/blob/master/kui-core/src/main/kotlin/org/kui/security/Crypto.kt)

```kotlin
package org.kui.security

import org.kui.util.getProperty
import java.security.SecureRandom
import java.util.*
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.spec.GCMParameterSpec
import javax.crypto.spec.SecretKeySpec
import javax.crypto.SecretKeyFactory
import javax.crypto.spec.PBEKeySpec

/**
 * Application cryptographic services including:
 * - GCM encrypt with system secret key
 * - Password hashing
 * - Security token generation and hashing
 */
object Crypto {
    /**
     * Secure random used by Crypto implementation to generate cryptographically strong random values.
     */
    private val secureRandom = SecureRandom.getInstanceStrong()!!
    /**
     * Security token salt. Generated on boot as tokens live as long as the instance lives.
     */
    private val securityTokenSalt: ByteArray = ByteArray(20)
    /**
     * The system encryption key. Loaded from property file.
     */
    private val systemEncryptionKey: SecretKeySpec

    /**
     * Static initialization of securityTokenSalt and systemEncryptionKey.
     */
    init {
        secureRandom.nextBytes(securityTokenSalt)

        val propertyCategoryKey = "security"
        val systemEncryptionKeyBase64 = getProperty(propertyCategoryKey, "system.encryption.key")
        if (systemEncryptionKeyBase64.trim().isEmpty()) {
            val keyGen = KeyGenerator.getInstance("AES")
            keyGen.init(AES_KEY_SIZE, secureRandom)
            val key = keyGen.generateKey()
            val systemEncryptionKeyBase64 = encode(key.encoded)
            println("System encryption key not set. Set property system.encryption.key to $systemEncryptionKeyBase64 to use new encryption key.")
            throw SecurityException("System encryption key not set.")
        } else {
            systemEncryptionKey = SecretKeySpec(decode(systemEncryptionKeyBase64), "AES")
        }
    }

    /**
     * Encrypts [plainText] with GSM using given [nonce] and [additionAuthenticatedData]
     * @return cipher text
     */
    fun encrypt(nonce: ByteArray, additionAuthenticatedData: ByteArray, plainText: ByteArray): ByteArray {
        val cipher = Cipher.getInstance(ENCRYPTION_ALGORITHM_AES_GCM, SECURITY_PROVIDER)
        val spec = GCMParameterSpec(GCM_TAG_LENGTH * 8, nonce)
        cipher.init(Cipher.ENCRYPT_MODE, systemEncryptionKey, spec)
        cipher.updateAAD(additionAuthenticatedData)
        return cipher.doFinal(plainText)
    }

    /**
     * Dencrypts [cipherText] with GSM using given [nonce] and [additionAuthenticatedData]
     * @return plain text
     */
    fun decrypt(nonce: ByteArray, additionAuthenticatedData: ByteArray, cipherText: ByteArray): ByteArray {
        val cipher = Cipher.getInstance(ENCRYPTION_ALGORITHM_AES_GCM, SECURITY_PROVIDER)
        val spec = GCMParameterSpec(GCM_TAG_LENGTH * 8, nonce)
        cipher.init(Cipher.DECRYPT_MODE, systemEncryptionKey, spec)
        cipher.updateAAD(additionAuthenticatedData)
        return cipher.doFinal(cipherText)
    }

    /**
     * Calculates [password] hash with given [salt].
     * @return hash bytes
     */
    fun passwordHash(salt: String, password: String) : ByteArray {
        val pbeKeySpec = PBEKeySpec(password.toCharArray(), salt.toByteArray(), PASSWORD_HASH_ITERATIONS, PASSWORD_HASH_KEY_SIZE)
        val secretKeyFactory = SecretKeyFactory.getInstance(PASSWORD_HASH_ALGORITHM_PBKDF2_HMAC_SHA512)
        return secretKeyFactory.generateSecret(pbeKeySpec).encoded
    }

    /**
     * Calculates [securityToken] hash.
     * @return hash bytes
     */
    fun securityTokenHash(securityToken: String) : ByteArray {
        val pbeKeySpec = PBEKeySpec(securityToken.toCharArray(), securityTokenSalt, PASSWORD_HASH_ITERATIONS, PASSWORD_HASH_KEY_SIZE)
        val secretKeyFactory = SecretKeyFactory.getInstance(PASSWORD_HASH_ALGORITHM_PBKDF2_HMAC_SHA512)
        return secretKeyFactory.generateSecret(pbeKeySpec).encoded
    }

    /**
     * Generates security token.
     */
    fun createSecurityToken() : String {
        var tokenBytes = ByteArray(20)
        secureRandom.nextBytes(tokenBytes)
        return Base64.getEncoder().encodeToString(tokenBytes)
    }

    /**
     * Encodes [bytes] with as Base64.
     * @return base 64 encoded string
     */
    private fun encode(bytes: ByteArray): String {
        return Base64.getEncoder().encodeToString(bytes)
    }

    /**
     * Decodes [text] from base 64 to bytes.
     * @return bytes
     */
    private fun decode(text: String): ByteArray {
        return Base64.getDecoder().decode(text)
    }
}
```

### encrypt & decrypt

```kotlin

import javax.crypto.Cipher
import javax.crypto.spec.SecretKeySpec


fun String.encrypt(cipherAlgo: String, keyAlgo: String, password: String): String {
    val secretKeySpec = SecretKeySpec(password.toByteArray(), keyAlgo)
    val cipher = Cipher.getInstance(cipherAlgo)
    cipher.init(Cipher.ENCRYPT_MODE, secretKeySpec)
    return cipher
        .doFinal(this.toByteArray())
        .toHex()
}

fun ByteArray.decrypt(cipherAlgo: String, keyAlgo: String, password: String) : String {
    val secretKeySpec = SecretKeySpec(password.toByteArray(), keyAlgo)
    val cipher = Cipher.getInstance(cipherAlgo)
    cipher.init(Cipher.DECRYPT_MODE, secretKeySpec)
    return String(cipher.doFinal(this))
}

/**
 * 다음 스펙에 따라 암호화
 *
 * ### Cipher
 * - algorithm: AES
 * - mode: ECB
 * - padding: PKCS5padding
 *
 * ### Key Spec
 * - key spec: Password Based Encryption(PBEKeySpec)
 * - key algorithms: AES
 */
fun String.encryptAesEcdPKCS5Padding(password: String) : String {
    return this.encrypt(
        "AES/ECB/PKCS5padding",
        "AES",
        password
    )
}

/**
 * 다음 스펙에 따라 암호화 된 문자열을 복호화
 *
 * ### Cipher
 * - algorithm: AES
 * - mode: ECB
 * - padding: PKCS5padding
 *
 * ### Key Spec
 * - key spec: Password Based Encryption(PBEKeySpec)
 * - key algorithms: AES
 */
fun String.decryptAesEcdPKCS5Padding(password: String) : String {
    return this.decodeHex().decrypt(
        "AES/ECB/PKCS5padding",
        "AES",
        password
    )
}

fun ByteArray.toHex() : String {
    return fold("") { str, it -> str + "%02x".format(it) }
}

fun String.decodeHex() : ByteArray {
    // 1. Split the string into 2-character pairs, representing each byte.
    val byteIterator = chunkedSequence(2)
        // 2. Parse each hex pair to their integer values.
        // 3. Convert the parsed Ints to Bytes.
        .map { it.toInt(16).toByte() }
        .iterator()

    return ByteArray(length / 2) { byteIterator.next() }
}
```

- [block cipher](https://www.techtarget.com/searchsecurity/definition/block-cipher)
    - `ECB`(Electronic codebook): ECB mode is used to electronically code messages as their plaintext form.
