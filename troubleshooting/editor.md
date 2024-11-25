# Editor

- [Editor](#editor)
    - [neovim 아이콘 깨짐](#neovim-아이콘-깨짐)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)

## neovim 아이콘 깨짐

### 문제

```shell
[?][?]
[?][?]
```

warp terminal에서는 정상적으로 출력되는 아이콘들이 iterm2에서는 위에처럼 ? 표시로 깨져서 나온다

### 원인

[폰트의 문제로 보인다](https://github.com/ryanoasis/vim-devicons/issues/226#issuecomment-492783382)

### 해결

```shell
brew tap homebrew/cask-fonts
brew install --cask font-hack-nerd-font
```

`.vimrc`, `.zshrc` 등의 파일에 설정

```zshrc
set encoding=utf-8
let g:airline_powerline_fonts = 1
```

[nvim-web-devicons](https://github.com/nvim-tree/nvim-tree.lua) 필요

[nerdfonts](https://www.nerdfonts.com/)에서 hack nerd font 설치

```shell
brew install --cask font-hack-nerd-font
```

그리고 iterm2 설정에서 Nerd Font 설정

warp terminal에서도 아이콘이 보이도록 하려면 [설정에서 Hack Nerd Font 선택 필요](https://github.com/warpdotdev/Warp/issues/1272#issuecomment-1316504220)
