# toctou

## toctou?

> Time-of-check to time-of-use

TOCTOU(Time-of-Check to Time-of-Use) 취약점은 컴퓨터 보안에서 발생하는 특정 유형의 취약점입니다.
*시스템이 어떤 자원의 상태를 확인한 후 그 자원을 사용하는 사이에 그 자원의 상태가 변경될 수 있는 상황*을 말합니다.
이로 인해 공격자가 시스템의 의도와 다르게 자원을 사용할 수 있게 됩니다.

이러한 취약점을 방지하기 위해서는 파일 디스크립터 사용, 원자적 연산 사용, 보안 API 사용 등의 방법을 고려할 수 있습니다.

1. **Time-of-Check (확인 시점)**

    시스템이 특정 자원의 상태를 확인하는 시점입니다.

    예를 들어, 파일에 대한 접근 권한을 확인하는 경우가 이에 해당합니다.

2. **Time-of-Use (사용 시점)**

    시스템이 확인한 자원을 실제로 사용하는 시점입니다.

    예를 들어, 파일을 열고 데이터를 쓰는 경우가 이에 해당합니다.

이 두 시점 사이에 자원의 상태가 변경될 수 있는 가능성이 존재합니다.
이로 인해 시스템이 의도한 대로 자원을 사용하지 못하게 되고, 보안 취약점이 발생할 수 있습니다.

```c
// code snippet used in `setuid` program

// `access` is intended to check whether the real user 
// who executed the `setuid` program would normally be allowed to write the file
if (access("file", W_OK) != 0) { // After the access check, before the open, 
                                 // the attacker replaces `file` with a `symlink` 
                                 // to the Unix password file `/etc/passwd`:
                                 //   symlink("/etc/passwd", "file");
    exit(1);
}

fd = open("file", O_WRONLY);     // Actually writing over `/etc/passwd`
write(fd, buffer, sizeof(buffer));
```

위의 코드에서 `access` 함수는 파일에 쓰기 권한이 있는지 확인합니다.
그러나 `access` 함수와 `open` 함수 사이의 시간 간격 동안 공격자가 `file`을 `/etc/passwd` 파일에 대한 심볼릭 링크로 교체할 수 있습니다.
이로 인해 `open` 함수는 원래 의도한 파일이 아닌 `/etc/passwd` 파일을 열게 되고, 결과적으로 시스템의 비밀번호 파일이 변경될 수 있습니다.

## TOCTOU 취약점의 위험성

TOCTOU 취약점은 다음과 같은 이유로 위험합니다:

- **예측 불가능성**: 자원의 상태가 언제 변경될지 예측할 수 없기 때문에, 시스템이 의도한 대로 자원을 사용할 수 없게 됩니다.
- **보안 위협**: 공격자가 이 취약점을 악용하여 시스템의 중요한 자원을 변경하거나 접근할 수 있게 됩니다.
- **데이터 무결성 손상**: 자원의 상태가 변경됨으로써 데이터의 무결성이 손상될 수 있습니다.

## 해결 방법

TOCTOU 취약점을 방지하기 위해서는 다음과 같은 방법을 사용할 수 있습니다:

1. **파일 디스크립터 사용**: 파일을 열 때 파일 디스크립터를 사용하여 파일의 상태를 확인하고, 그 디스크립터를 통해 파일을 조작합니다.
2. **원자적 연산 사용**: 파일의 상태를 확인하고 사용하는 작업을 원자적으로 수행하여 중간에 파일의 상태가 변경되지 않도록 합니다.
3. **보안 API 사용**: 보안이 강화된 API를 사용하여 파일의 상태를 확인하고 사용하는 작업을 안전하게 수행합니다.

이러한 방법들을 통해 TOCTOU 취약점을 방지할 수 있습니다.

## 참고 자료

- [Time-of-check to time-of-use](https://en.wikipedia.org/wiki/Time-of-check_to_time-of-use)
