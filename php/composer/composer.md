# Composer

- [Composer](#composer)
    - [commands](#commands)
        - [downgrade](#downgrade)
        - [`why-not`](#why-not)
        - [업데이트](#업데이트)
        - [삭제](#삭제)

## [commands](https://getcomposer.org/doc/03-cli.md)

### downgrade

```shell
composer self-update --2.2

# Warning: You forced the install of 2.2.22 via --2.2, but 2.6.5 is the latest stable version. Updating to it via composer self-update --stable is recommended.
# Upgrading to version 2.2.22 (2.2.x channel).
   
# Use composer self-update --rollback to return to version 2.6.5
```

### `why-not`

- Composer의 의존성 트리를 확인하여 어떤 패키지가 `phpseclib/phpseclib`와 충돌하는지 확인할 수 있다

```shell
composer why-not phpseclib/phpseclib:3.0.33
```

### 업데이트

```bash
#!/bin/bash

# 특정 패키지만 업데이트: https://stackoverflow.com/a/16740711
composer update package/name:{version}

# 경로 지정
composer install --working-dir=/home/someuser/myproject

composer update --working-dir=/path/to/IdeaProjects/service/directory package/name:dev-branch
```

### 삭제

```bash
❯ which php-cs-fixer
/Users/rody/.composer/vendor/bin/php-cs-fixer
```

실행 파일만 삭제하는 것보다는 `php-cs-fixer`와 관련된 다른 파일들도 함께 삭제하는 것이 좋습니다.
`composer`를 사용하여 설치한 패키지는 일반적으로 여러 종속성 파일과 함께 설치됩니다.
이를 모두 삭제하려면 다음과 같이 합니다.

터미널에서 다음 명령을 실행하여 `php-cs-fixer`가 설치된 디렉토리로 이동하세요:

```bash
cd /Users/rody/.composer/vendor
```

그 후, `php-cs-fixer`와 관련된 디렉토리를 찾아 삭제합니다. 대부분의 경우 패키지 이름은 해당 패키지의 디렉토리 이름과 일치합니다. 다음 명령을 사용할 수 있습니다:

```bash
rm -rf friendsofphp/
```

패키지를 수동으로 삭제한 후에는 `composer.json` 파일과 `composer.lock` 파일을 업데이트해야 할 수 있습니다.
이 파일들은 종속성 관리에 사용되므로, 이들에서 `php-cs-fixer` 관련 항목을 제거해야 할 수 있습니다.

직접 편집하거나, 패키지를 설치할 때 사용한 `composer` 명령을 통해 패키지를 제거할 수도 있습니다.
다음 명령을 사용하면 `composer`가 자동으로 이 파일들을 업데이트해 줍니다:

```bash
# 이 명령은 `composer`를 통해 글로벌로 설치된 `php-cs-fixer`를 제거합니다.
composer global remove friendsofphp/php-cs-fixer
```

마지막으로, 시스템이 오래된 파일을 참조하는 것을 방지하기 위해 캐시를 정리하는 것이 좋습니다. `composer`의 캐시를 정리하려면 다음 명령을 사용합니다:

```bash
composer clear-cache
```
