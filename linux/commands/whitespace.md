# whitespace

## 공백 문자 체크

```bash
#!/bin/bash

FLAG=0
if git --no-pager grep -Il '\s$'; then
  echo $'\nFound trailing whitespaces\n'
  FLAG=1
fi
if git --no-pager grep -Il --perl-regexp '[^\n]\z'; then
  echo $'\nNo LF at EOF\n'
  FLAG=1
fi
if (( FLAG )); then
  echo 'Issue found'
  exit 1
fi

echo 'No trailing whitespaces. All texts have LF at EOF'
```
