# Loops

## `forEach`에서 `continue`와 `break`

- `return@forEach`를 사용한다
- https://ponyozzang.tistory.com/316

```kotlin
// continue
value.forEach {
    if (it == null) return@forEach
    list.add(toCharset(it, charset))
}
```

```kotlin
// break
run {
    (1..10).forEach {
        if (it == 3) {
            return@run
        }
        println(it)
    }
}
```

