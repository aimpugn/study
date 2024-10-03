# Date

- [Date](#date)
  - [format](#format)
    - [`yyyyMMddHms`](#yyyymmddhms)
    - [`yyyyMMddHHmmss`](#yyyymmddhhmmss)

## format

```kotlin
println(DateTimeFormatter
    .ofPattern("{패턴}")
    .withZone(ZoneId.systemDefault())
    .format(Instant.now()))
```

### `yyyyMMddHms`

- 20221209122536

### `yyyyMMddHHmmss`

- 20221209122600
