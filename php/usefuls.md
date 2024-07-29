# Utils

## print as php array

```php
public function printArray($arr, $tab = '    ')
{
    echo PHP_EOL;
    $str = json_encode($arr, JSON_PRETTY_PRINT | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASH);
    $str = preg_replace('/\{/', '[', $str);
    $str = preg_replace('/\}/', ']', $str);
    $str = preg_replace('/"/', "'", $str);
    $str = preg_replace('/: /', ' => ', $str);
    print_r($str);
    echo PHP_EOL;
}
```

```log
Array
(
    [0] => Array
        (
            [pg_provider] => html5_inicis
            [pg_id] => MOImovin0
            [pg_secret] => 1111
            [cancel_password] =>
            [sandbox] =>
            [active] => 1
            [channel_id] =>
        )

    [1] => Array
        (
            [pg_provider] => html5_inicis
            [pg_id] => MOVmovingc
            [pg_secret] =>
            [cancel_password] =>
            [sandbox] =>
            [active] => 1
            [channel_id] =>
        )

)
```

```php
[

    [
        'pg_provider' => 'html5_inicis',
        'pg_id' => 'MOImovin0',
        'pg_secret' => '1111',
        'cancel_password' => '',
        'sandbox' => false,
        'active' => true,
        'channel_id' => null
    ],
    [
        'pg_provider' => 'html5_inicis',
        'pg_id' => 'MOVmovingc',
        'pg_secret' => '',
        'cancel_password' => '',
        'sandbox' => false,
        'active' => true,
        'channel_id' => null
    ]
]
```
