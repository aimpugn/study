# brace expansion

## brace expansion?

중괄호 확장(Brace Expansion)은 쉘 스크립트에서 반복 작업을 간단하게 처리할 수 있는 강력한 기능입니다.
파일 및 디렉토리 생성, 문자열 조합, 숫자 및 문자 범위 생성 등 다양한 작업에 유용하게 사용할 수 있습니다.

중괄호 `{}` 안에 쉼표로 구분된 항목들을 나열하여 여러 문자열을 생성합니다.

```bash
echo {a,b,c}
# 출력: a b c
```

### 예시

여러 파일이나 디렉토리를 한 번에 생성할 수 있습니다.

```bash
# 파일 생성
touch file{1,2,3}.txt
# 결과: file1.txt, file2.txt, file3.txt 생성

# 디렉토리 생성
mkdir dir{A,B,C}
# 결과: dirA, dirB, dirC 생성
```

중괄호 확장을 사용하여 문자열을 조합할 수 있습니다.

```bash
echo {a,b,c}{1,2,3}
# 출력: a1 a2 a3 b1 b2 b3 c1 c2 c3
```

숫자 범위를 사용하여 일련의 숫자를 생성할 수 있습니다.

```bash
echo {1..5}
# 출력: 1 2 3 4 5

# 역순으로 생성
echo {5..1}
# 출력: 5 4 3 2 1
```

문자 범위를 사용하여 일련의 문자를 생성할 수 있습니다.

```bash
echo {a..e}
# 출력: a b c d e

# 대문자 범위
echo {A..E}
# 출력: A B C D E
```

중첩된 중괄호를 사용하여 복잡한 문자열을 생성할 수 있습니다.

```bash
echo {{a,b},{1,2}}
# 출력: a b 1 2

echo {a,b}{1,2}
# 출력: a1 a2 b1 b2
```

중괄호 확장에 접두사와 접미사를 추가할 수 있습니다.

```bash
echo pre{a,b,c}post
# 출력: prea preb precpost

# 파일 이름에 접두사와 접미사 추가
touch file_{1..3}.txt
# 결과: file_1.txt, file_2.txt, file_3.txt 생성
```

복잡한 문자열 조합을 생성할 수 있습니다.

```bash
echo {a,b,c}{1..3}{X,Y}
# 출력: a1X a1Y a2X a2Y a3X a3Y b1X b1Y b2X b2Y b3X b3Y c1X c1Y c2X c2Y c3X c3Y
```

## 주의사항

- 공백 주의: 중괄호 안에 공백이 있으면 예상치 못한 결과가 나올 수 있습니다. 공백을 피하거나 따옴표로 묶어야 합니다.

  ```bash
  echo {a, b, c}
  # 출력: a b c (공백 포함)

  echo "{a, b, c}"
  # 출력: {a, b, c} (따옴표로 묶음)
  ```

- 쉘 지원: 중괄호 확장은 대부분의 쉘(Bash, Zsh 등)에서 지원되지만, 모든 쉘에서 지원되는 것은 아닙니다.

## 디렉토리 내의 여러 파일을 간단하게 표현하는 수학적 방법

중괄호 확장(Brace Expansion)을 사용하는 것이 일반적입니다. 이 방법은 특히 쉘 스크립트에서 자주 사용됩니다.

```bash
dir/{a,b,c}.txt
```

### 설명

- 중괄호 확장(Brace Expansion): 중괄호 `{}` 안에 쉼표로 구분된 항목들을 나열하여, 여러 파일이나 디렉토리를 간단하게 표현할 수 있습니다.
    - 예: `dir/{a,b,c}.txt`는 `dir/a.txt`, `dir/b.txt`, `dir/c.txt`를 의미합니다.

### 다른 표현 방법

- 파이프(|) 사용: 파이프는 일반적으로 정규 표현식에서 "또는"을 의미하지만, 쉘 스크립트에서는 중괄호 확장과 함께 사용되지 않습니다.
    - `dir/{a|b|c}.txt` 또는 `dir/[a|b|c].txt`는 올바른 표현이 아닙니다.

### 예시

#### 쉘 스크립트에서의 사용 예시

```bash
# 파일 생성
touch dir/{a,b,c}.txt

# 결과 확인
ls dir/
# 출력:
# a.txt
# b.txt
# c.txt
```

이와 같이, `dir/{a,b,c}.txt`는 `dir/a.txt`, `dir/b.txt`, `dir/c.txt`를 간단하게 표현하는 올바른 방법입니다.
