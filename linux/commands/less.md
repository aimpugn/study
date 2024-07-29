# less

## UTF-8 문자셋 적용

- https://serverfault.com/questions/414760/how-to-make-the-less-command-handle-utf-8/414763#414763

```shell
export LESSCHARSET=utf-8

less -r <file_name>
```