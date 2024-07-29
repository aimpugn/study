# test

## [check if a file or directory exists](https://linuxize.com/post/bash-check-if-file-exists/)

```shell
FILE=/etc/resolv.conf
if test -f "$FILE"; then
    echo "$FILE exists."
fi

FILE=/etc/resolv.conf
if [ -f "$FILE" ]; then
    echo "$FILE exists."
fi

FILE=/etc/resolv.conf
if [[ -f "$FILE" ]]; then
    echo "$FILE exists."
fi
```

## `-eq`, `-gt`, `-ge`

```shell
#!/bin/bash

[ "100" -eq "100" ]

[ "1000" -gt "100" ]

[ "1000" -ge "100" ]
```
