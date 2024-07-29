# Reflection

## 클래스 속성의 값 가져오기

- https://stackoverflow.com/a/70178693

```kotlin
myObj.javaClass.kotlin.memberProperties.foreach { property ->
      property.get(myObj)
}
```