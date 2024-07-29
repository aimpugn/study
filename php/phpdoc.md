# PHP Doc

- [PHP Doc](#php-doc)
    - [PHPDoc tags](#phpdoc-tags)
    - [`@property`](#property)
    - [`@link`](#link)
    - [`@see`](#see)

## PHPDoc tags

- [fig-standards](https://github.com/phpDocumentor/fig-standards/blob/master/proposed/phpdoc-tags.md#517-var)

## `@property`

- PHPDoc의 `@property` 태그는 `__get()` /`__set()` 메서드를 통해 액세스되는 magic(not real / dynamic) 필드에만 사용된다
- `bar` real property에 대한 타입 힌트가 필요한 경우 적절한 `@var` 태그를 사용하여 입력 필요

    ```php
    /** @var string My cool bar property */
    public $bar;
    ```

## `@link`

`@link` 태그는 일반적으로 URL을 참조할 때 사용되며, 코드 내의 특정 변수나 메서드를 참조하는 데 사용되지 않습니다.

## `@see`

변수나 메서드를 참조하려면 `@see` 태그를 사용하는 것이 적합합니다.

- 정적 변수를 참조할 때는 `@link` 대신 `@see`를 사용합니다.
- 정적 변수에 접근할 때는 `$`를 사용하지 않습니다. `self::` 또는 `static::`를 사용합니다.

    PHPDoc 주석에서 정적 변수를 참조할 때 `self::`를 붙이지 않아도 됩니다.
    주석 내에서는 변수명만 사용해도 충분합니다.
    다만, 정적 변수는 `self::$someStaticClass`처럼 사용되기 때문에 `self::`를 붙이는 것이 더 명확할 수 있습니다.

    따라서 `{@link $someStaticClass}` 대신 `{@see self::$someStaticClass}`를 사용하는 것이 좋습니다.

- 주석에서 변수를 명확하게 참조하려면 `@see self::$someStaticClass`와 같은 형태를 사용합니다.

```php
class Test {
    /** @var SomeClass */
    private static $someStaticClass;

    /**
     * {@see self::$someStaticClass}를 반환합니다. (O)
     * {@link someStaticClass}를 반환합니다. (X)
     *
     * @return SomeClass
     */
    private static function test() {
        return self::$someStaticClass;
    }
}
```
