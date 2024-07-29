# Supported Annotations

## [Supported Annotations](https://psalm.dev/docs/annotating_code/supported_annotations/)

## `@psalm-type`

- 다른 타입에 대한 별칭을 정의할 수 있다

```php
<?php
/**
 * @psalm-type PhoneType = array{phone: string}
 */
class Phone {
    /**
     * @psalm-return PhoneType
     */
    public function toArray(): array {
        return ["phone" => "Nokia"];
    }
}
```

## `@psalm-import-type`

- 어딘가에 `@psalm-type`으로 정의되어 있는 타입을 import 할 수 있다.

```php
<?php
/**
 * @psalm-import-type PhoneType from Phone
 */
class User {
    /**
     * @psalm-return PhoneType
     */
    public function toArray(): array {
        return array_merge([], (new Phone())->toArray());
    }
}
```

- import 할 때 별칭(`as`)을 사용할 수 있다

```php
<?php
/**
 * @psalm-import-type PhoneType from Phone as MyPhoneTypeAlias
 */
class User {
    /**
     * @psalm-return MyPhoneTypeAlias
     */
    public function toArray(): array {
        return array_merge([], (new Phone())->toArray());
    }
}
```
