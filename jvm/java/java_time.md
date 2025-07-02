# Java Time

- [Java Time](#java-time)
    - [시간 관련 주요 클래스](#시간-관련-주요-클래스)
    - [`Temporal`](#temporal)
    - [레거시 API](#레거시-api)
    - [`LocalDateTime`과 `Instant` 상호 변환](#localdatetime과-instant-상호-변환)
        - [`LocalDateTime` to `Instant`](#localdatetime-to-instant)
        - [`Instant` to `LocalDateTime`](#instant-to-localdatetime)

## 시간 관련 주요 클래스

- `LocalDate`: 날짜(연,월,일), 시간 없음

    ```java
    LocalDate.now();
    LocalDate.of(2024, 3, 15); // 2024-03-15
    ```

- `LocalTime`: 시간(시,분,초,나노초), 날짜 없음

    ```java
    LocalTime.now();
    LocalTime.of(14, 30); // 14:30
    ```

- `LocalDateTime`: 날짜 + 시간, 타임존 없음

    ```java
    LocalDateTime.now();
    LocalDateTime.of(2024, 3, 15, 14, 30);
    // LocalDate + LocalTime => LocalDateTim
    LocalDateTime combined = localDate.atTime(localTime);
    LocalDateTime combined2 = localTime.atDate(localDate);

    // LocalDateTime => LocalDate, LocalTime
    LocalDate dt = dateTime.toLocalDate();
    LocalTime tm = dateTime.toLocalTime();
    ```

- `OffsetDateTime`: 날짜 + 시간 + UTC 오프셋

    ```java
    OffsetDateTime.now();
    ```

- `Instant`: UTC 기준 타임라인상의 한 시점

    UTC 기준(1970-01-01T00:00:00Z)으로부터 경과된 시간을 나노초(nanoseconds) 정밀도로 표현하는 타임스탬프입니다.
    시간대 정보가 없으며, 오직 UTC를 기준으로 한 절대적인 시간 값만을 가집니다.

    ```java
    Instant.now();

    System.out.println("Epoch 밀리초로 생성: " + Instant.ofEpochMilli(System.currentTimeMillis()));
    ```

- `Duration`: 시간 기반의 시간의 양을 나타냅니다. (시, 분, 초, 나노초 단위)

    ```java
    System.out.println(Instant.now()); // 2025-07-02T14:10:17.449184400Z
    System.out.println(Instant.now().minusSeconds(60 * 60 * 5)); // 2025-07-02T09:10:17.449184400Z
    System.out.println(Duration.between(Instant.now(), Instant.now().minusSeconds(60 * 60 * 5))); // PT-5H

    Duration twoHours = Duration.ofHours(2);
    ```

- `Period`: 두 날짜 간의 기간(년,월,일)

    ```java
    Period period = Period.between(date1, date2);
    Period fiveDays = Period.ofDays(5);
    Period.between(d1, d2);
    ```

- `Year`, `Month`, `YearMonth`, `MonthDay`: 부분적 날짜 표현

    ```java
    Year.now();
    MonthDay.now();
    ```

- `Clock`: 시간 소스 추상화(테스트, 모킹용)

    ```java
    Clock.systemUTC();
    ```

- `ZonedDateTime`: 날짜 + 시간 + 타임존

    `LocalDateTime`에 시간대(`ZoneId`) 정보가 결합된 형태입니다.
    서머타임(DST)과 같은 시간대 규칙을 자동으로 처리하므로, 가장 정확한 시간 표현입니다.

    ```java
    ZonedDateTime.now();
    ```

- `ZoneId`, `ZoneOffset`

    `ZoneId`은 'Asia/Seoul'이나 'America/New_York'과 같은 시간대 식별자입니다.
    `ZoneOffset`은 UTC와의 시차를 나타냅니다(예: +09:00). `ZoneId`와 달리 서머타임 규칙을 포함하지 않습니다.

    ```java
    ZoneId.of("Asia/Seoul");
    ```

- `DateTimeFormatter`

    ```java
    DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
    String formatted = dateTime.format(formatter);
    LocalDateTime parsed = LocalDateTime.parse("2024-03-15 14:30:00", formatter);
    ```

## `Temporal`

```java
System.out.println(LocalDate.now()); // 2025-07-02
// 다음 금요일 찾기
System.out.println(LocalDate.now().with(TemporalAdjusters.next(DayOfWeek.FRIDAY))); // 2025-07-04
// 월의 마지막 날
System.out.println(LocalDate.now().with(TemporalAdjusters.lastDayOfMonth())); // 2025-07-31

// 특정 단위로 계산
System.out.println(LocalDate.now().plus(5, ChronoUnit.DAYS)); // 2025-07-07
System.out.println(ChronoUnit.DAYS.between(
    LocalDate.of(2025, 7, 4),
    LocalDate.of(2025, 7, 31)
)); // 27
```

## 레거시 API

```java
// Date => Instant => LocalDateTime
Date legacyDate = new Date();
Instant instant = legacyDate.toInstant();
LocalDateTime localDateTime = instant.atZone(ZoneId.systemDefault()).toLocalDateTime();

// LocalDateTime => Instant => Date
Date convertedDate = Date.from(localDateTime.atZone(ZoneId.systemDefault()).toInstant());
```

## `LocalDateTime`과 `Instant` 상호 변환

`Instant`는 UTC 기준의 절대적 시점을 나타내며, `LocalDateTime`은 특정 시간대에서의 상대적 시간을 표현합니다.

- `LocalDateTime` to `Instant`: `LocalDateTime + ZoneId/ZoneOffset` => UTC 기준 절대 시점
- `Instant` to `LocalDateTime`: `UTC 절대 시점 + ZoneId` => 특정 시간대의 상대 시간

```java
// 동일한 순간을 다른 시간대로 표현
var seoulTime = LocalDateTime.of(2024, 1, 1, 9, 30);
var moment = seoulTime.atZone(ZoneId.of("Asia/Seoul")).toInstant();

// 같은 순간을 뉴욕 시간으로 표현
var nyTime = moment.atZone(ZoneId.of("America/New_York")).toLocalDateTime();

System.out.println("Seoul: " + seoulTime); // Seoul: 2024-01-01T09:30
System.out.println("UTC: " + moment); // UTC: 2024-01-01T00:30:00Z
System.out.println("New York: " + nyTime); // New York: 2023-12-31T19:30
```

### `LocalDateTime` to `Instant`

`LocalDateTime`은 시간대 정보가 없는 날짜 및 시간이므로, `Instant`로 변환하려면 반드시 시간대 정보를 제공해야 합니다.

`ZoneId`는 시간대 규칙을 식별하는 클래스이며, `ZoneOffset`은 UTC로부터의 고정된 시차를 나타냅니다. ZoneId로부터 특정 시점의 `ZoneOffset`을 얻을 수 있습니다:
이는 서머타임([Daylight Saving Time, DST](https://ko.wikipedia.org/wiki/%EC%9D%BC%EA%B4%91_%EC%A0%88%EC%95%BD_%EC%8B%9C%EA%B0%84%EC%A0%9C)) 처리 때문에 필요합니다.
동일한 `ZoneId`라도 시점에 따라 `ZoneOffset`이 달라질 수 있기 때문입니다.

```java
var endDt = LocalDateTime.of(2025, 6, 24, 16, 30, 00);
var startDt = endDt.minusHours(1);

// Instant 타입으로 변환하기 위해 시간대 정보 얻기
var zoneId = ZoneId.of("Asia/Seoul"); // 특정 지역의 ZoneId 얻기
System.out.println("zoneId: " + zoneId); // Asia/Seoul
System.out.println("ZoneId.systemDefault: " + ZoneId.systemDefault()); // Asia/Seoul

var endInstant = endDt.atZone(zoneId).toInstant();
var startInstant = startDt.atZone(zoneId).toInstant();
System.out.println("Instant.parse: " + Instant.parse("2007-12-03T10:15:30.00Z")); // 2007-12-03T10:15:30Z
System.out.println("endInstant: " + endInstant); // 2025-06-24T07:30:00Z
System.out.println("startInstant: " + startInstant); // 2025-06-24T06:30:00Z

// ZoneOffset
System.out.println("ZoneId getOffset: " + ZoneId.systemDefault().getRules().getOffset(startInstant)); // +09:00
```

또는 고정된 `ZoneOffset`을 사용하여 변환할 수 있습니다.

```java
var localDateTime = LocalDateTime.now();

// 고정된 오프셋 사용
var kstOffset = ZoneOffset.ofHours(9);
var instant = localDateTime.toInstant(kstOffset);
```

### `Instant` to `LocalDateTime`

특정 시간대 또는 시스템 기본 시간대로 변환할 수 있습니다.

```java
// 특정 시간대로 변환
System.out.println(LocalDateTime.ofInstant(Instant.now(), ZoneId.of("Asia/Seoul")));
// 2025-07-02T22:39:07.319995900

// 시스템 기본 시간대로 변환
System.out.println(LocalDateTime.ofInstant(Instant.now(), ZoneId.systemDefault()));
//2025-07-02T22:39:07.319995900
```

```java
var ldt3 = Instant.now().atZone(ZoneId.of("Asia/Seoul")).toLocalDateTime();
// 2025-07-02T22:38:32.755535200
```
