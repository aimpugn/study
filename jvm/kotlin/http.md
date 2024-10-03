# HTTP

- [HTTP](#http)
    - [DTO to x-www-form-urlencoded](#dto-to-x-www-form-urlencoded)
    - [form data request](#form-data-request)
    - [webflux webclient parameters](#webflux-webclient-parameters)

## DTO to x-www-form-urlencoded

- <https://jojoldu.tistory.com/478>

## form data request

- url encoded

```postman
name1:test+data
name2:10000
name3%5B%5D:list+value1
name3%5B%5D:list+value2
name4%5B%5D:1
name4%5B%5D:2
name4%5B%5D:3
name4%5B%5D:4
name5%5Bkey%5D:val
name5%5Bkey2%5D:val2
name6%5Bkey3%5D%5B%5D:v1
name6%5Bkey3%5D%5B%5D:v2
```

```php
Array
(
    [name1] => test+data
    [name2] => 10000
    [name3%5B%5D] => list+value2
    [name4%5B%5D] => 4
    [name5%5Bkey%5D] => val
    [name5%5Bkey2%5D] => val2
    [name6%5Bkey3%5D%5B%5D] => v2
)
```

- url decoded

```postman
name1:test+data
name2:10000
name3[]:list+value1
name3[]:list+value2
name4[]:1
name4[]:2
name4[]:3
name4[]:4
name5[key]:val
name5[key2]:val2
name6[key3][]:v1
name6[key3][]:v2
```

```php
Array
(
    [name1] => test+data
    [name2] => 10000
    [name3] => Array
        (
            [0] => list+value1
            [1] => list+value2
        )

    [name4] => Array
        (
            [0] => 1
            [1] => 2
            [2] => 3
            [3] => 4
        )

    [name5] => Array
        (
            [key] => val
            [key2] => val2
        )

    [name6] => Array
        (
            [key3] => Array
                (
                    [0] => v1
                    [1] => v2
                )

        )

)

```

## webflux webclient parameters

- <https://www.baeldung.com/webflux-webclient-parameters>

```kotlin
internal fun generateWebClient(
    baseUrl: String,
    readTimeout: Long = 5000,
    writeTimeout: Long = 5000,
    responseTimeout: Long = 5000
): WebClient {
    val httpClient = HttpClient
        .create()
        .wiretap(HttpMessageConvert::javaClass.name, LogLevel.DEBUG, AdvancedByteBufFormat.TEXTUAL)
        .option(ChannelOption.CONNECT_TIMEOUT_MILLIS, 5000)
        .responseTimeout(Duration.ofMillis(responseTimeout))
        .doOnConnected {
            it.addHandlerLast(ReadTimeoutHandler(readTimeout, TimeUnit.MILLISECONDS))
            it.addHandlerLast(WriteTimeoutHandler(writeTimeout, TimeUnit.MILLISECONDS))
        }

    return WebClient
        .builder()
        .clientConnector(ReactorClientHttpConnector(httpClient))
        .baseUrl(baseUrl)
        .defaultHeader(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_JSON_VALUE)
        .build()
}

/*// object mapper 사용 경우
val mapper = ObjectMapper()
//                                         `object : `가 붙어야 한다
val convertedMap = mapper.convertValue(dto, object : TypeReference<Map<String, Any?>>() {})
convertedMap.forEach {(key, value) ->
    when (value) {
        is List<*> -> {
            setAsList(key, value, params)
        }
        is Map<*, *> -> {
            setAsMap(key, value, params)
        }
        else -> {
            params[key] = listOf(value.toString())
        }
    }
}*/
/**
 * `x-form-urlencoded` 형식으로 요청하기 위해 DTO를 MultiValueMap<String, String> 형식으로 처리
 * data class Dto(
 *  val name1: String,
 *  val name2: Number,
 *  val name3: List<String>? = null,
 *  val name4: List<Number>? = null,
 *  val name5: Map<String, String>? = null,
 *  val name6: Map<String, List<String>>? = null,
 * )
 * 
 * {
 *  "name1": "test",
 *  "name2": "1000",
 *  "name3": ["value1", "value2", "value3"],
 *  "name4": ["1", "2", "3", "4", "5"],
 *  "name5": {
 *  "innerName2": "innerValue2",
 *  "innerName1": "innerValue1"
 *  },
 *  "name6": {
 *  "innerName3": ["innerValue3", "innerValue4"]
 *  }
 * }
 *
 * @return
 */
fun Any.toMultiValueMap(charset: Charset = Charsets.UTF_8): LinkedMultiValueMap<String, String> {
    try {
        val params = LinkedMultiValueMap<String, String>()
        this.javaClass.kotlin.memberProperties.forEach {
            when (val value = it.get(this)) {
                is List<*> -> {
                    setAsList(it.name, value, params, charset)
                }
                is Map<*, *> -> {
                    setAsMap(it.name, value, params, charset)
                }
                else -> setAsString(it.name, value, params, charset)
            }
        }
        return params
    } catch (e: Exception) {
        throw Exception("Failed convert object to MultiValueMap, ${e.localizedMessage}")
    }
}

private fun setAsList(key: String, value: List<*>, params: LinkedMultiValueMap<String, String>, charset: Charset) {
    val list = mutableListOf<String>()
    val newKey = toCharset("$key[]", charset)
    value.forEach {
        if(it == null) return@forEach
        list.add(toCharset(it, charset))
    }
    params[newKey] = list
}

private fun setAsMap(key: String, value: Map<*, *>, params: LinkedMultiValueMap<String, String>, charset: Charset) {
    value.forEach { (mapKey, mapValue) ->
        val newKey = toCharset("$key[$mapKey]", charset)
        when (mapValue) {
            is List<*> -> setAsList(newKey, mapValue, params, charset)
            is Map<*, *> -> setAsMap(newKey, mapValue, params, charset)
            else -> setAsString(newKey, mapValue, params, charset)
        }
    }
}

private fun setAsString(key: String, value: Any?, params: LinkedMultiValueMap<String, String>, charset: Charset) {
    when (value) {
        null -> {}
        else -> {
            params[key] = toCharset(value, charset)
        }
    }
}

fun toCharset(raw: Any, charset: Charset) : String {
    return when (charset) {
        Charsets.UTF_8 -> raw.toString()
        else -> String(raw.toString().toByteArray(charset), charset)
    }
}
```
