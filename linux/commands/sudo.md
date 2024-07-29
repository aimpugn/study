# sudo

- [sudo](#sudo)
    - [sudo: unable to resolve host](#sudo-unable-to-resolve-host)
    - [`-s`, `--shell`](#-s---shell)
    - [`-i`, `--login`](#-i---login)
    - [추가적인 유용한 sudo 옵션들](#추가적인-유용한-sudo-옵션들)
    - [실제 사용 예시 및 비교](#실제-사용-예시-및-비교)

## sudo: unable to resolve host

```shell
#!/bin/sh

# sudo: unable to resolve host
# ㄴ https://askubuntu.com/a/119723
# ㄴ /etc/hostname 과 /ets/hosts에 있는 이름이 매치되지 않아서 발생
```

## `-s`, `--shell`

```sh
sudo -s

/home/ubuntu# pwd
/home/ubuntu
```

다음 둘 중 하나의 shell을 실행합니다.
- SHELL 환경 변수가 설정되어 있는 경우 그 변수에서 지정한 셸
- 호출 사용자의 비밀번호 데이터베이스 항목에서 지정한 셸

명령이 지정되면 `-c` 옵션을 사용하여 간단한 명령으로 셸에 전달됩니다.

명령과 인수는 영숫자, 밑줄(`_`), 하이픈(`-`) 및 달러 기호(`$`)를 제외한 각 문자(공백 포함)를 백슬래시('\')로 이스케이프 처리한 후 공백으로 구분하여 연결됩니다.

명령을 지정하지 않으면 대화형 셸이 실행됩니다.

- 기본적으로 현재 사용자의 환경 변수(`$HOME`, `$SHELL`, `$USER`, `$LOGNAME`, `$PATH`)를 유지합니다.
    > 그러나 `/etc/sudoers` 파일의 설정에 따라 이 동작이 변경될 수 있습니다.
    >
    > ```sh
    > Defaults        env_reset
    > Defaults        mail_badpass
    > Defaults        secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin"
    >
    > # This preserves proxy settings from user environments of root
    > # equivalent users (group sudo)
    > # Defaults:%sudo env_keep += "http_proxy https_proxy ftp_proxy all_proxy no_proxy"
    > ... 생략 ...
    > ```

- 현재 디렉토리를 유지합니다.
- 현재 사용자의 `.bashrc` 등을 사용하는 등, 대부분의 현재 사용자 환경 설정을 유지합니다.

대부분의 셸은 명령어를 지정할 때 대화형 세션과 다르게 동작하므로 자세한 내용은 셸의 매뉴얼을 참조하세요.

## `-i`, `--login`

```sh
sudo -i

~# pwd
/root
```

대상 사용자의 비밀번호 데이터베이스 항목에서 로그인 셸로 지정한 셸을 실행합니다.
즉, `.profile`, `.bash_profile` 또는 `.login`과 같은 로그인 관련 리소스 파일이 셸에서 읽혀집니다.

명령이 지정되면 `-c` 옵션을 사용하여 간단한 명령으로 셸에 전달됩니다.
명령과 모든 인수는 영숫자, 밑줄(`_`), 하이픈(`-`) 및 달러 기호(`$`)를 제외한 각 문자(공백 포함)를 백슬래시('\')로 이스케이프 처리한 후 공백으로 구분하여 연결됩니다.

명령을 지정하지 않으면 대화형 셸이 실행됩니다.
`sudo`는 *셸을 실행하기 전에 해당 사용자의 홈 디렉터리로 변경을 시도*합니다.
이 명령은 사용자가 로그인할 때 받는 환경과 유사한 환경에서 실행됩니다.
1. root 사용자의 환경으로 완전히 전환합니다.
2. root의 홈 디렉토리로 이동합니다.
3. root의 환경 변수를 로드합니다.
4. 로그인 셸을 시작합니다.

대부분의 셸은 대화형 세션과 비교하여 명령이 지정될 때 다르게 동작하므로 자세한 내용은 셸의 매뉴얼을 참조하세요.
`sudoers(5)` 매뉴얼의 명령 환경 섹션에는 sudoers 정책이 사용 중일 때 `-i` 옵션이 명령이 실행되는 환경에 어떤 영향을 미치는지에 대한 설명이 나와 있습니다.

## 추가적인 유용한 sudo 옵션들

1. **sudo -u [사용자명]**
    - 지정된 사용자로 명령을 실행합니다.

    ```sh
    sudo -u postgres psql
    ```

2. **sudo -E**
   - 현재 사용자의 환경 변수를 유지한 채 명령을 실행합니다.

   ```sh
   sudo -E command
   ```

3. **sudo -k**
   - sudo 인증 캐시를 무효화합니다.

   ```sh
   sudo -k
   ```

4. **sudo -l**
   - 현재 사용자의 sudo 권한을 나열합니다.

   ```sh
   sudo -l
   ```

5. **sudo -v**
   - sudo 인증을 갱신합니다 (타임아웃 연장).

   ```sh
   sudo -v
   ```

## 실제 사용 예시 및 비교

```sh
# 현재 디렉토리 확인
pwd
/home/user

# sudo -i 사용
sudo -i
pwd  # 결과: /root

# sudo -s 사용
exit  # root에서 나옴
sudo -s
pwd  # 결과: /home/user

# sudo su - 사용
exit  # 다시 일반 사용자로
sudo su -
pwd  # 결과: /root
```

이러한 옵션들을 이해하고 적절히 사용하면 시스템 관리 작업을 더 효과적이고 안전하게 수행할 수 있습니다. 각 옵션의 특성을 고려하여 상황에 맞는 가장 적절한 방법을 선택하는 것이 중요합니다.
