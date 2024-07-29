# stdClass vs object

- [stdClass vs object](#stdclass-vs-object)
    - [stdClass와 object의 차이](#stdclass와-object의-차이)
    - [왜 IDE에서 경고가 발생하는가?](#왜-ide에서-경고가-발생하는가)

## stdClass와 object의 차이

- stdClass

    `stdClass`는 PHP에서 객체를 만들 때 기본적으로 사용하는 클래스입니다.
    즉, 명시적으로 클래스를 정의하지 않고 객체를 생성할 때 사용하는 클래스입니다.

    `stdClass`는 `new stdClass()`로 생성된 객체이며, 기본적으로 아무런 메소드나 속성을 갖지 않습니다.

- object

    `object`는 PHP에서 객체 타입을 나타내는 일반적인 타입입니다.
    어떤 클래스에서 생성된 객체인지 상관없이 모든 객체는 `object` 타입에 속합니다.

    `object`는 타입 힌팅과 주석에서 객체 타입을 나타내기 위해 사용됩니다.

## 왜 IDE에서 경고가 발생하는가?

```php
/**
 * @var stdClass{
 *     code: int,
 *     details: string,
 * } $status
 */
$status = new stdClass();
$status->code = -1;
$status->details = '';

... 어떤 로직 ...
list($res, $status) = $this->client->SomeGRPCFunction($req, $metadata)->wait();
... 어떤 로직 ...

if ($status->code !== 0) {
    ^^^^^^^^^^^^^^^^^^^ 
    `@var stdClass` 사용할 경우 이 부분에서 IDE가 아래처럼 알려준다.
    "Condition is always 'true' because '$status->code' is evaluated at this point"
    반면 `@var object`라고 하면 경고가 없다.
    throw new Exception($status->details, $status->code);
}
```

1. `stdClass`로 주석 처리된 경우

    ```php
    /**
     * @var stdClass{
     *     code: int,
     *     details: string,
     * } $status
     */
    $status = new stdClass();
    $status->code = -1;
    $status->details = '';
    ```

    이 경우, `stdClass` 타입으로 주석을 처리했기 때문에 IDE는 `$status`가 항상 `stdClass` 타입의 객체라고 가정합니다.
    따라서 `SomeGRPCFunction` 호출 후에도 여전히 `stdClass` 타입이어야 한다고 판단합니다.
    하지만 `SomeGRPCFunction`에서 반환된 `$status`는 실제로 `stdClass`가 아닌 다른 타입일 수 있습니다.
    이로 인해 IDE가 경고를 발생시키는 것입니다.

2. `object`로 주석 처리된 경우

    ```php
    /**
     * @var object{
     *     code: int,
     *     details: string,
     * } $status
     */
    $status = new stdClass();
    $status->code = -1;
    $status->details = '';
    ```

   이 경우, `object` 타입으로 주석을 처리했기 때문에 IDE는 `$status`가 특정 클래스가 아니라 객체 타입이면 된다고 판단합니다.
   즉, `SomeGRPCFunction`에서 반환된 `$status`가 `stdClass`가 아니더라도 `object` 타입으로 받아들일 수 있으므로 경고가 발생하지 않습니다.

IDE 경고를 피하기 위해서는 `stdClass` 대신 `object` 타입으로 주석을 처리하는 것이 좋습니다.
이는 `SomeGRPCFunction`이 반환하는 객체가 항상 `stdClass`가 아닐 수 있다는 점을 명확히 나타냅니다.
