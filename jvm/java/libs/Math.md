# Math

- [Math](#math)
    - [`Math.ceil`](#mathceil)
    - [`Math.ceilDiv`](#mathceildiv)

## `Math.ceil`

```java
System.out.println(5 / 2); // 2
System.out.println(Math.ceil(5 / 2)); // 2.0
System.out.println(Math.ceil((double) 5 / 2)); // 3.0
System.out.println(Math.ceil((double) 70 / 30)); // 3.0
System.out.println(8 / 3); // 2
System.out.println(Math.ceil(8 / 3)); // 2.0
System.out.println(Math.ceil((double) 8 / 3)); // 3.0
```

## `Math.ceilDiv`

java 18 버전부터 사용 가능합니다.

```java
System.out.println(Math.ceilDiv(5, 2));
// 3
System.out.println(Math.ceilDiv(8, 3));
// 3
```

18 버전 전의 경우에는 아래와 같이 한 수를 double로 만들어서 실수로 계산이 되도록 하고,
그 다음에 올림 처리한 후 다시 정수로 변환합니다.([프로그래머스 '기능개발' 문제 참고](../../../algorithms/programmers/src/main/java/p42586/Main.java))

```java
Math.ceil(((double) 100 - progress) / speed);
```
