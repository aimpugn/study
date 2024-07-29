# grep

## grep?

대신 `rg` 사용하기

## not matching

```bash
#!/bin/bash

# grep not matching
# - `--invert-match` 옵션
# - https://stackoverflow.com/a/3548465
# - https://stackoverflow.com/a/4538335
grep -v
grep -v -e 'negphrase1' -e 'negphrase2'
grep -v 'negphrase1|negphrase2|negphrase3'
```
