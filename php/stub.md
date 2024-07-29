# stubs

- [stubs](#stubs)
    - [intellij](#intellij)
        - [PHP Runtime 추가](#php-runtime-추가)
        - [또는 `phpstorm-stubs` 직접 추가](#또는-phpstorm-stubs-직접-추가)

## intellij

### PHP Runtime 추가

`Settings > Languages & Frameworks > PHP` 이동. PHP Runtime 탭에서 전체 체크

### 또는 `phpstorm-stubs` 직접 추가

```log
You're looking at immutable embedded tstubs. To be able to edit them you need to clone a stubs project, then provide a default stubs path via PHP / PHP Runtime / Advanced settings
```

```shell
g clone https://github.com/JetBrains/phpstorm-stubs.git
```

그리고 `Settings > Languages & Frameworks > PHP`에서 Include Path에 해당 경로 추가
