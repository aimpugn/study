# ObjectMapper

## Naming Strategy

- 아래 과정에서 네이밍을 수정하는데, 문제는 `POJOPropertyBuilder[] props = propMap.values().toArray(new POJOPropertyBuilder[propMap.size()]);` 시에 원래 속성과 camelCase 된 속성이 추가된다는 점.

```java
// modules-2/files-2.1/com.fasterxml.jackson.core/jackson-databind/2.13.1/3a556489e4a16b4837fd640c46a858c8402a1eb/jackson-databind-2.13.1-sources.jar!/com/fasterxml/jackson/databind/introspect/POJOPropertiesCollector.java
public class POJOPropertiesCollector {
    protected void _renameUsing(Map<String, POJOPropertyBuilder> propMap,
                                PropertyNamingStrategy naming) {
        if (_forSerialization) {
            if (prop.hasGetter()) {
                rename = naming.nameForGetterMethod(_config, prop.getGetter(), fullName.getSimpleName());
            } else if (prop.hasField()) {
                rename = naming.nameForField(_config, prop.getField(), fullName.getSimpleName());
            }

        }
    }
}
```

## `@JsonNaming` vs `@JsonProperty`

- `@JsonNaming` 적용 시 `@JsonProperty`가 있으면, `@JsonProperty`가 선언된 필드는 `translate` 과정에서 제외된다
