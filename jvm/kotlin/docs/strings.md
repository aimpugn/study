# Strings

## byte or char to hex

- [시프트 연산](https://codedragon.tistory.com/7998)
  - `shr`, `shl`, `ushr` 등

```kotlin
// https://docs.oracle.com/javase/tutorial/i18n/text/examples/UnicodeFormatter.java
// https://docs.oracle.com/javase/tutorial/i18n/text/string.html
fun byteToHex(b: Byte) : String {
  // Returns hex String representation of byte b
  val hexDigit = listOf(
    '0', '1', '2', '3', '4', '5', '6', '7',
    '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'
  )
  val arr = listOf(hexDigit[(b.toInt() shr 4) and 0x0f], hexDigit[b.toInt() and 0x0f]).map{ it.code.toByte() }.toByteArray()
  return String(arr);
}

fun charToHex(c: Char) : String{
    // Returns hex String representation of char c
    // Byte hi = (Byte) (c >>> 8);
    // Byte lo = (Byte) (c & 0xff);
    val hi = (c.code ushr 8).toByte()
    val lo = (c.code and 0xff).toByte()
    return byteToHex(hi) + byteToHex(lo);
}

fun printBytes(bytes: ByteArray?) {
    if (bytes == null) {
        return
    }
    for (byte in bytes) {
        println(byteToHex(byte))
    }
}
```

## 인코딩 처리

- [[코틀린] 한글 깨질 때 인코딩처리](https://jessyt.tistory.com/119)

## StringBuilder

- 아래 두 코드의 결과가 상이하다. 왜일까?

```kotlin
val sb: StringBuilder = StringBuilder(propertyName)
sb.setCharAt(0, firstChar.uppercaseChar())
return sb.toString()
```

```kotlin
return StringBuilder(propertyName)
    .setCharAt(0, firstChar.uppercaseChar())
    .toString()
```