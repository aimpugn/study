# gRPC MapField

- [gRPC MapField](#grpc-mapfield)
    - [MapField?](#mapfield)
    - [예제](#예제)

## MapField?

`MapField` 클래스는 Google Protocol Buffers의 PHP 구현에서 맵 필드를 다루기 위해 사용됩니다.
이 클래스는 *키-값 쌍을 저장*하는 데 사용되며, 키와 값의 타입을 지정할 수 있습니다.

`MapField`를 사용하려면, 키와 값의 타입을 지정하고, 필요한 경우 메시지나 열거형 클래스 이름을 제공해야 합니다.

키와 값의 타입을 지정하고, 필요한 경우 메시지나 열거형 클래스 이름을 제공하여 `MapField` 객체를 생성할 수 있습니다.
그런 다음, 배열처럼 값을 추가, 접근, 삭제할 수 있습니다.

## 예제

1. **기본 사용법**: 키와 값의 타입을 지정하여 `MapField` 객체를 생성합니다.

    ```php
    use Google\Protobuf\Internal\MapField;
    use Google\Protobuf\Internal\GPBType;

    // 키 타입: INT32, 값 타입: STRING
    $mapField = new MapField(GPBType::INT32, GPBType::STRING);

    // 값 추가
    $mapField->offsetSet(2, new Google\Protobuf\Value([
        'string_value' => 'Value1',
    ]));
    $mapField->offsetSet(3, new Google\Protobuf\Value([
        'string_value' => 'Value2',
    ]));

    // 전체 맵 출력
    /** @var Google\Protobuf\Value $value */
    foreach ($mapField->getIterator() as $key => $value) {
        echo "Key: $key, Value: $value\n";
    }
    ```

2. **메시지 타입 사용**: 값 타입이 메시지인 경우, 메시지 클래스 이름을 제공해야 합니다.

    ```php
    $invalidArgumentViolationField = new Google\Protobuf\Internal\MapField(
        Google\Protobuf\Internal\GPBType::STRING,
        Google\Protobuf\Internal\GPBType::MESSAGE,
        Google\Protobuf\Value::class
    );

    $invalidArgumentViolationField->offsetSet('some_key', new Google\Protobuf\Value([
        'string_value' => 'should not empty',
    ]));
    ```

3. **열거형 타입 사용**: 값 타입이 열거형인 경우, 열거형 클래스 이름을 제공해야 합니다.

    ```php
    use Google\Protobuf\Internal\MapField;
    use Google\Protobuf\Internal\GPBType;
    use YourNamespace\YourEnumClass;

    // 키 타입: INT32, 값 타입: ENUM
    $mapField = new MapField(GPBType::INT32, GPBType::ENUM, YourEnumClass::class);

    // 값 추가
    $mapField[2] = YourEnumClass::VALUE1;
    $mapField[3] = YourEnumClass::VALUE2;

    // 값 접근
    echo $mapField[2];  // 출력: VALUE1
    echo $mapField[3];  // 출력: VALUE2

    // 값 삭제
    unset($mapField[2]);

    // 전체 맵 출력
    foreach ($mapField as $key => $value) {
        echo "Key: $key, Value: $value\n";
    }
    ```
