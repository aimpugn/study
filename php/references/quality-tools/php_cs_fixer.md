# php-cs-fixer

## php-cs-fixer란?

php-cs-fixer는 PHP 코드 스타일을 자동으로 수정해주는 도구입니다. 코드 스타일을 일관되게 유지하고, 코드 리뷰 시간을 줄이는 데 도움이 됩니다.

## 설치

```bash
composer require --dev friendsofphp/php-cs-fixer
```

## 사용법

### 설정 파일 생성

```bash
vendor/bin/php-cs-fixer init
```

### 코드 스타일 수정

```bash
vendor/bin/php-cs-fixer fix
```

### 설정 파일을 이용한 코드 스타일 수정

```bash
vendor/bin/php-cs-fixer fix --config=.php_cs.dist
```

### 설정 파일을 이용한 코드 스타일 수정 (대상 파일 지정)

```bash
vendor/bin/php-cs-fixer fix --config=.php_cs.dist src/
```

### 설정 파일을 이용한 코드 스타일 수정 (대상 파일 지정, dry-run)

```bash
vendor/bin/php-cs-fixer fix --config=.php_cs.dist src/ --dry-run
```

### 설정 파일을 이용한 코드 스타일 수정 (대상 파일 지정, dry-run, verbose)

```bash
vendor/bin/php-cs-fixer fix --config=.php_cs.dist src/ --dry-run --verbose
```

### 설정 파일을 이용한 코드 스타일 수정 (대상 파일 지정, dry-run, verbose, diff)

```bash
vendor/bin/php-cs-fixer fix --config=.php_cs.dist src/ --dry-run --verbose --diff
```

## 구성

### IntelliJ file watcher 설정

1. `Preferences` > `Tools` > `File Watchers`로 이동
2. `+` 버튼을 클릭하여 `php-cs-fixer`를 추가
    - `File Type`: PHP
    - Program: `/opt/homebrew/Cellar/php-cs-fixer/3.49.0/bin/php-cs-fixer`
    - Arguments: `fix -vvv --config=$ProjectFileDir$/.php-cs-fixer.php $FileDir$/$FileName$`
    - Working directory: `$ProjectFileDir$`
    - Trigger the watcher on external changes: 체크
3. 저장
