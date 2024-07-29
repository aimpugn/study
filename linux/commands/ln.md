# ln

- [ln](#ln)
    - [ln?](#ln-1)
    - [hard link](#hard-link)
    - [symbolic(soft)  link](#symbolicsoft--link)
    - [원본 파일을 삭제했을 경우](#원본-파일을-삭제했을-경우)

## ln?

- link 의 약자로 리눅스에서 원본 파일이나 디렉터리를 가리키는 link 를 만드는 명령어

## hard link

- 하드 링크는 원본 파일에 대한 alias 와 비슷한 개념으로 하드 링크를 생성하면 **원본과 동일한 inode**를 갖는 파일이 만들어 집니다.
- 하드 링크는 만들고 나면 무엇이 원본이고 무엇이 링크인지 알기가 어렵습니다.
- 하드 링크는 ln 명령어로 바로 만들수 있으며 다음은 원본 파일인 origin 에 대한 하드 링크인 hard-link 라는 파일을 생성합니다

```shell
ln origin hard-link
```

- 만들어진 하드 링크는 `ls -l` 로 봐도 구분하기가 어려운 문제가 있다
- 하드 링크인지 알아보려면 inode 번호를 보는 옵션인 `-i` 를 같이 사용해서 inode 번호의 일치 여부를 확인

## symbolic(soft)  link

- 심볼릭(소프트) 링크는 윈도우의 바로가기와 비슷한 개념으로 원본을 가리키는 역할을 수행하므로 하드 링크와는 달리
    - **다른 파티션이나 파일 시스템에 있어도 생성이 가능**하며
    - **원본과는 다른 inode number**를 갖게 됩니다.
- 만들려면 `ln` 에 symbolic link 생성 옵션인  `-s` 를 추가

```shell
# 다음은 $HOME 밑에 etc_bashrc 라는 링크 파일을 생성합니다.
ln -s /etc/bashrc ${HOME}/etc_bashrc
```

- 심볼릭 링크는 `ls -l` 로 볼 경우 원본을 포인팅하므로 쉽게 알아볼 수 있다

```shell
$ ls -l $HOME/etc_bashrc 

lrwxrwxrwx. 1 lesstif lesstif 11 Jan 14 23:16 /home/lesstif/etc_bashrc -> /etc/bashrc
```

- 만약 링크 파일이 이미 있을 경우 ln 을 실행하면 다음과 같이 "File exists" 에러가 발생합니다.

```shell
$ ln -s /etc/bashrc ${HOME}/etc_bashrc

ln: failed to create symbolic link '/home/lesstif/etc_bashrc': File exists
```

- 기본에 심볼릭 링크가 있을 경우 덮어쓰는 옵션인 `-f` 를 사용하면 기존 링크 파일이 있어도 정상 동작합니다.

```shell
ln -sf /etc/bashrc ${HOME}/etc_bashrc
```

## 원본 파일을 삭제했을 경우

- 하드 링크?
    - 같은 inode number 를 공유하므로 원본을 삭제해도 다른 하드 링크가 있다면 원본 파일 내용을 읽는데 아무 문제가 없다
- 소프트 링크?
    - 바로가기이므로 원본이 삭제되면 소프트 링크 파일은 사용 불가하게 된다
