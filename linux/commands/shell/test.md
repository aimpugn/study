# test

## 기본 조건문: `[ ]`

```bash
if [ "$var" = "value" ]; then
    echo "Condition is true"
fi
```

## `&&`

```bash
if [[ "$var1" == "value1" && "$var2" == "value2" ]]; then
    echo "Both conditions are true"
fi
```

## 확장 조건문: `[[ ]]`

```bash
if [[ "$var" == "value" ]]; then
    echo "Condition is true"
fi
```
