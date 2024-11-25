# Terminal

- [Terminal](#terminal)
    - [iterm2 - looks like you're trying to copy to the pasteboard](#iterm2---looks-like-youre-trying-to-copy-to-the-pasteboard)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [`p9k_configure` 후 에러](#p9k_configure-후-에러)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)

## iterm2 - looks like you're trying to copy to the pasteboard

### 문제

마우스로 여러 문자열을 선택하고, cmd + c로 배너로 경고가 나옴

```warn
Looks like you're trying to copy to the pasteboard, but mouse reporting has prevented making a selection. Disable mouse reporting?
```

### 원인

- [Annoying banner: "Looks like you're trying to copy to the pasteboard..."](https://gitlab.com/gnachman/iterm2/-/issues/8905)

> You just made a selection and then pressed cmd-c.  
> Because mouse reporting was on, iTerm2 doesn't know about the selection and is unable to copy to the clipboard.  
> You might have been tricked into believing that cmd-C has an effect because some applications running in the terminal save selections value to your clipboard when you finish dragging.  
> In any case, pressing cmd-c has no effect.

### 해결

1. ~~아래 명령어를 사용하면 문제가 해결된다고 한다~~ 해결 안됨

```shell
defaults write com.googlecode.iterm2 NoSyncNeverAskAboutMouseReportingFrustration -bool true
```

2. [또는 `Option`을 누르면서 선택 후 cmd + c 하면 복사할 수 있다고 한다](https://stackoverflow.com/a/19843650)

3. [~~`또는`iTerm2 -> Settings -> General -> Selection`에서`Applications in terminal may access clipboard` 체크~~](https://stackoverflow.com/a/38849483)

## `p9k_configure` 후 에러

### 문제

```bash
gitstatus_stop_p9k_:zle:41: No handler installed for fd 14
gitstatus_stop_p9k_:49: failed to close file descriptor 14: bad file descriptor 
```

### 원인

비동기 코드에서 문제가 있는 것으로 보인다.

> This is probably another instance of the bug that [zsh-users/zsh-autosuggestions#753](https://github.com/zsh-users/zsh-autosuggestions/pull/753) should fix. You can verify this by adding the following line at the very bottom of `~/.zshrc`:
>
> ```bash
> unset ZSH_AUTOSUGGEST_USE_ASYNC
> ```

Zsh의 Powerlevel10k 테마와 특히 Zsh Line Editor(ZLE)와의 상호작용에서 비동기 작업과 관련된 문제를 나타낸다.

잠재적인 해결책 중 하나는 `ZSH_AUTOSUGGEST_USE_ASYNC` 환경 변수를 설정 해제하여 Zsh에서 비동기 제안을 비활성화하는 것이다.

### 해결

잘못된 설정 프로그램을 실행했다..

```bash
p10k configure
```

```zsh
# p9k인 경우
# `.zshrc` 파일에 추가
# unset ZSH_AUTOSUGGEST_USE_ASYNC
```
