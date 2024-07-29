# DateTime

## format

PHP에서 `DateTime` 객체를 원하는 형식으로 변환하려면 `format` 메서드를 사용할 수 있습니다.

```php
// DateTime 객체 생성 예제
$createdAt = new DateTime('2024-06-27 02:24:54.930068', new DateTimeZone('+00:00'));
$startAt = new DateTime('2024-06-23 15:00:00.000000', new DateTimeZone('+00:00'));
$endAt = new DateTime('2024-08-23 15:00:00.000000', new DateTimeZone('+00:00'));

// 변환할 DateTime 객체 배열
$dateTimes = [
    'createdAt' => $createdAt,
    'startAt' => $startAt,
    'endAt' => $endAt,
];

// 변환 결과를 저장할 배열
$formattedDates = [];

// DateTime 객체를 'Ymd His' 형식으로 변환
foreach ($dateTimes as $key => $dateTime) {
    $formattedDates[$key] = $dateTime->format('Ymd His');
}

// 결과 출력
print_r($formattedDates);
```

### 설명

1. **DateTime 객체 생성**:
   - `DateTime` 객체를 생성할 때, 날짜와 시간 문자열을 전달하고, 타임존을 설정할 수 있습니다.

   ```php
   $createdAt = new DateTime('2024-06-27 02:24:54.930068', new DateTimeZone('+00:00'));
   $startAt = new DateTime('2024-06-23 15:00:00.000000', new DateTimeZone('+00:00'));
   $endAt = new DateTime('2024-08-23 15:00:00.000000', new DateTimeZone('+00:00'));
   ```

2. **DateTime 객체 배열**:
   - 변환할 `DateTime` 객체들을 배열에 저장합니다.

   ```php
   $dateTimes = [
       'createdAt' => $createdAt,
       'startAt' => $startAt,
       'endAt' => $endAt,
   ];
   ```

3. **DateTime 객체 변환**:
   - `foreach` 루프를 사용하여 각 `DateTime` 객체를 `'Ymd His'` 형식으로 변환합니다.

   ```php
   foreach ($dateTimes as $key => $dateTime) {
       $formattedDates[$key] = $dateTime->format('Ymd His');
   }
   ```

4. **결과 출력**:
   - 변환된 날짜와 시간을 저장한 배열을 출력합니다.

   ```php
   print_r($formattedDates);
   ```

### 결과

위 예제 코드를 실행하면 다음과 같은 결과를 얻을 수 있습니다:

```plaintext
Array
(
    [createdAt] => 20240627 022454
    [startAt] => 20240623 150000
    [endAt] => 20240823 150000
)
```

### 요약

- `DateTime` 객체를 원하는 형식으로 변환하려면 `format` 메서드를 사용합니다.
- `'Ymd His'` 형식은 `Y` (연도), `m` (월), `d` (일), `H` (시), `i` (분), `s` (초)를 나타냅니다.
- `foreach` 루프를 사용하여 여러 `DateTime` 객체를 변환할 수 있습니다.

이 방법을 사용하면 `DateTime` 객체를 원하는 형식으로 변환하여 사용할 수 있습니다.
