# Loops

## for loop

```kotlin
for (item in collection) print(item)

for (item: Int in ints) {
    // ...
}
```

## [ranges and progressions](https://kotlinlang.org/docs/ranges.html)

```kotlin
// Closed-ended range
println(4 in 1..4)
// true

// Open-ended range
// `<` 표현식은 1.7부터 사용 가능합니다.
println(4 in 1..<4)
// false

for (i in 1..4) print(i)
// 1234

for (i in 4 downTo 1) print(i)
// 4321

for (i in 0..8 step 2) print(i)
// 02468

// `<` 표현식은 1.7부터 사용 가능합니다.
for (i in 0..<8 step 2) print(i)
// 0246

for (i in 8 downTo 0 step 2) print(i)
// 86420
```

## `forEach`에서 `continue`와 `break`

- `return@forEach`를 사용한다

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
