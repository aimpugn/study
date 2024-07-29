# ob functions

- [ob functions](#ob-functions)
    - [ob? Output Control](#ob-output-control)
    - [출력 버퍼링의 기본 개념](#출력-버퍼링의-기본-개념)
    - [주요 `ob_*` 함수들과 그 용도](#주요-ob_-함수들과-그-용도)
        - [`ob_start()`: 출력 버퍼링 시작](#ob_start-출력-버퍼링-시작)
        - [`ob_get_clean()`: 버퍼의 내용을 가져오고 버퍼를 비움](#ob_get_clean-버퍼의-내용을-가져오고-버퍼를-비움)
        - [`ob_end_clean()`: 버퍼를 비우고 출력 버퍼링을 종료](#ob_end_clean-버퍼를-비우고-출력-버퍼링을-종료)
        - [`ob_get_flush`: 현재 버퍼의 내용을 출력하고, 버퍼의 내용을 가져오고, 버퍼를 비움](#ob_get_flush-현재-버퍼의-내용을-출력하고-버퍼의-내용을-가져오고-버퍼를-비움)
        - [`ob_end_flush()`: 현재 버퍼의 내용을 출력하고 버퍼를 비움](#ob_end_flush-현재-버퍼의-내용을-출력하고-버퍼를-비움)
        - [`ob_get_contents()`: 현재 버퍼의 내용을 가져옴 (버퍼를 비우지 않음)](#ob_get_contents-현재-버퍼의-내용을-가져옴-버퍼를-비우지-않음)
        - [`ob_flush()`: 버퍼의 내용을 출력하지만 버퍼를 비우지 않음](#ob_flush-버퍼의-내용을-출력하지만-버퍼를-비우지-않음)
        - [`ob_clean()`: 버퍼의 내용을 비우지만 출력하지 않음](#ob_clean-버퍼의-내용을-비우지만-출력하지-않음)
    - [실제 사용 사례](#실제-사용-사례)
        - [헤더 수정](#헤더-수정)
        - [템플릿 시스템](#템플릿-시스템)
        - [긴 처리 작업 중 프로그레스 표시](#긴-처리-작업-중-프로그레스-표시)
        - [캐싱](#캐싱)
    - [주의사항](#주의사항)
    - [하나의 버퍼만 남기기](#하나의-버퍼만-남기기)

## ob? Output Control

PHP의 `ob_*` 함수들은 출력 버퍼링(Output Buffering)을 제어하는 함수들입니다.
이 기능은 PHP 스크립트의 출력을 즉시 브라우저로 보내지 않고 임시로 저장했다가 한 번에 보내거나 수정할 수 있게 해줍니다.

## 출력 버퍼링의 기본 개념

PHP의 `ob_*` 함수들은 출력 버퍼링을 제어하는 함수들입니다.
이 함수들을 이해하기 위해서는 먼저 출력 버퍼링의 개념과 동작 방식을 이해해야 합니다.

출력 버퍼링은 *PHP 스크립트의 출력을 즉시 웹 서버로 보내지 않고, 메모리 버퍼에 임시로 저장*하는 메커니즘입니다.
이를 통해 출력을 조작하거나 지연시킬 수 있습니다.

기본적인 동작 방식:
1. PHP 스크립트 실행 시작
2. 출력 버퍼 생성
3. 스크립트의 출력이 버퍼에 저장
4. 스크립트 실행 종료 또는 버퍼가 가득 차면 버퍼의 내용을 웹 서버로 전송

## 주요 `ob_*` 함수들과 그 용도

### `ob_start()`: 출력 버퍼링 시작

스크립트 시작 부분이나 출력을 캡처하고 싶은 부분 직전에 사용합니다.

```php
ob_start();
echo "This will be buffered";
$content = ob_get_clean(); // 버퍼 내용을 가져오고 버퍼를 비움
```

### `ob_get_clean()`: 버퍼의 내용을 가져오고 버퍼를 비움

> 1. Get the contents of the active output buffer
> 2. and turn it off

버퍼의 내용을 가져오고 즉시 새로운 출력을 시작하고 싶을 때 사용합니다.

```php
ob_start();
echo "This is cached";
$cached = ob_get_clean();
echo "This is immediate output";
// 결과: "This is immediate output"이 즉시 출력되고, $cached에 "This is cached"가 저장됨
```

### `ob_end_clean()`: 버퍼를 비우고 출력 버퍼링을 종료

> 1. Clean (erase) the contents of the active output buffer
> 2. and turn it off

### `ob_get_flush`: 현재 버퍼의 내용을 출력하고, 버퍼의 내용을 가져오고, 버퍼를 비움

> 1. Flush (send) the return value of the active output handler,
> 2. return the contents of the active output buffer
> 3. and turn it off

### `ob_end_flush()`: 현재 버퍼의 내용을 출력하고 버퍼를 비움

> 1. Flush (send) the return value of the active output handler
> 2. and turn the active output buffer off

버퍼의 내용을 출력하고 싶을 때 사용합니다.

```php
ob_start();
echo "Hello, World!";
ob_end_flush(); // "Hello, World!" 출력
```

### `ob_get_contents()`: 현재 버퍼의 내용을 가져옴 (버퍼를 비우지 않음)

버퍼의 내용을 조작하거나 저장하고 싶을 때 사용합니다.

```php
ob_start();
echo "Original content";
$content = ob_get_contents();
$modified_content = str_replace("Original", "Modified", $content);
ob_end_clean(); // 버퍼를 비우고 출력하지 않음
echo $modified_content; // "Modified content" 출력
```

### `ob_flush()`: 버퍼의 내용을 출력하지만 버퍼를 비우지 않음

버퍼의 현재 내용을 출력하고 계속해서 버퍼링하고 싶을 때 사용합니다.

```php
ob_start();
echo "Part 1";
ob_flush(); // "Part 1" 출력
echo "Part 2";
ob_end_flush(); // "Part 2" 출력
```

### `ob_clean()`: 버퍼의 내용을 비우지만 출력하지 않음

버퍼의 내용을 삭제하고 새로 시작하고 싶을 때 사용합니다.

```php
ob_start();
echo "This will be discarded";
ob_clean();
echo "This will be kept";
ob_end_flush(); // "This will be kept" 출력
```

## 실제 사용 사례

### 헤더 수정

```php
ob_start();
echo "Some content";
if (some_condition()) {
    header("Location: new_page.php");
    exit;
}
ob_end_flush();
```

이 경우, 출력이 시작된 후에도 헤더를 수정할 수 있습니다.

### 템플릿 시스템

```php
ob_start();
include "template.php";
$template_content = ob_get_clean();
echo str_replace("{{TITLE}}", "My Page Title", $template_content);
```

템플릿의 내용을 캡처하고 동적으로 수정할 수 있습니다.

### 긴 처리 작업 중 프로그레스 표시

```php
ob_start();
for ($i = 0; $i < 100; $i++) {
    echo ".";
    if ($i % 10 == 0) {
        ob_flush();
        flush();
    }
    sleep(1);
}
ob_end_flush();
```

이 예제는 긴 처리 작업 중 진행 상황을 실시간으로 표시합니다.

### 캐싱

```php
$cache_file = 'cache.txt';
if (file_exists($cache_file) && (time() - filemtime($cache_file) < 3600)) {
    readfile($cache_file);
} else {
    ob_start();
    // 시간이 많이 걸리는 작업 수행
    echo "This is cached content";
    $cached_content = ob_get_clean();
    file_put_contents($cache_file, $cached_content);
    echo $cached_content;
}
```

이 예제는 출력을 캐시하여 성능을 향상시킵니다.

## 주의사항

- 중첩된 버퍼: `ob_start()`를 여러 번 호출하면 중첩된 버퍼가 생성됩니다. 이 경우 `ob_end_*` 함수들은 가장 안쪽의 버퍼부터 작동합니다.
- 메모리 사용: 큰 출력을 버퍼링할 때는 메모리 사용량에 주의해야 합니다.
- 출력 콜백: `ob_start()`에 콜백 함수를 지정하여 버퍼의 내용을 자동으로 처리할 수 있습니다.

출력 버퍼링은 PHP에서 강력한 도구이지만, 적절히 사용하지 않으면 예기치 않은 결과를 초래할 수 있습니다.각 함수의 동작을 정확히 이해하고 사용하는 것이 중요합니다.

## 하나의 버퍼만 남기기

```php
if (ob_get_level() == 0) {
    ob_start();
} else {
    while (ob_get_level() > 1) {
        ob_end_clean();
    }
}
```

이 코드는 다음과 같은 동작을 수행합니다:

1. `ob_get_level() == 0`인 경우:
   - 현재 활성화된 출력 버퍼가 없다는 의미입니다.
   - `ob_start()`를 호출하여 새로운 출력 버퍼를 시작합니다.

2. `ob_get_level() > 0`인 경우:
   - 이미 하나 이상의 출력 버퍼가 활성화되어 있다는 의미입니다.
   - `while` 루프를 사용하여 `ob_get_level()`이 1보다 큰 동안 (즉, 2개 이상의 버퍼가 있는 경우) 반복합니다.
   - 각 반복에서 `ob_end_clean()`을 호출하여 가장 안쪽의 버퍼를 비우고 종료합니다.

이 코드의 목적은 다음과 같습니다:
- 출력 버퍼가 없는 경우, 새로운 버퍼를 시작합니다.
- 출력 버퍼가 이미 존재하는 경우, 최상위 레벨의 버퍼만 남기고 나머지는 모두 제거합니다.

이런 방식은 출력 버퍼의 상태를 일관되게 유지하고, 불필요한 중첩 버퍼를 제거하여 출력 관리를 단순화하는 데 사용됩니다.
특히 복잡한 애플리케이션이나 프레임워크에서 출력을 일관되게 제어하고자 할 때 유용할 수 있습니다.
