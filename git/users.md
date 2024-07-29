# git users

- [git users](#git-users)
    - [user config 확인](#user-config-확인)
    - [SSH key 생성](#ssh-key-생성)
    - [기타](#기타)

## user config 확인

```shell
git config --global user.name
git config --global user.email
```

## SSH key 생성

```shell
# 제공된 이메일을 레이블로 사용하여 새 SSH 키가 생성
ssh-keygen -t ed25519 -C "aimpugn@gmail.com"

# Ed25519 알고리즘을 지원하지 않는 레거시 시스템을 사용하는 경우
# ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
```

```log
❯ ssh-keygen -t ed25519 -C "aimpugn@gmail.com"
Generating public/private ed25519 key pair.
Enter file in which to save the key (/Users/rody/.ssh/id_ed25519): git_aimpugn_ed25519
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved in git_aimpugn_ed25519
Your public key has been saved in git_aimpugn_ed25519.pub
The key fingerprint is:
SHA256:PcadLpSLj9J4vDoN8pbpXxzsZMmFSwWZ96n1IBcGhtU aimpugn@gmail.com
The key's randomart image is:
+--[ED25519 256]--+
|          .*=o   |
|          ++. E  |
|          o..o o |
|         * *..*  |
|        S / o= o |
|    . .  O =.   .|
|     o O. * .    |
|      O =+ .     |
|     oo*+..      |
+----[SHA256]-----+
```

macOS Sierra 10.12.2 이상을 사용하는 경우 ssh-agent에 키를 자동으로 로드하고 키 집합에 암호를 저장하도록 `~/.ssh/config` 파일을 수정

```conf
Host github.com
  AddKeysToAgent yes
  UseKeychain yes
  IdentityFile ~/.ssh/id_ed25519
```

```shell
ssh-add --apple-use-keychain ~/.ssh/git_aimpugn_ed25519
# Identity added: /Users/rody/.ssh/git_aimpugn_ed25519 (aimpugn@gmail.com)
```

## 기타

- [GitHub SSH 접속 설정하기](https://devocean.sk.com/blog/techBoardDetail.do?ID=163311)
