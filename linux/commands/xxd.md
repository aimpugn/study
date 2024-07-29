# xxd

- [xxd](#xxd)
    - [xxd?](#xxd-1)
    - [Options](#options)
        - [`-b`](#-b)
    - [예제](#예제)
    - [기타](#기타)

## xxd?

> xxd - make a hexdump or do the reverse.

## Options

### `-b`

16진수 덤프가 아닌 비트(2진수) 덤프로 전환합니다.
이 옵션은 옥텟을 일반 16진수 덤프 대신 8자리 "1"과 "0"으로 기록합니다.
각 줄 앞에는 16진수로 된 줄 번호가 오고 그 뒤에는 ASCII(또는 EBCDIC) 표현이 옵니다.
`-p`, `-i` 명령줄 스위치는 이 모드에서 작동하지 않습니다.

```sh
printf 'W' | xxd -b
00000000: 01010111
```

## 예제

```shell
╰─ echo "e110c300ac0a0094bba84f57b4514b520aad082b47a266a58a2967b95ad49bcee16bac492658cc5494a50236d0a35d1e" | xxd -r -p

��
OWQKR
+Gf)gZԛ�kI&X�6У]%  
```

```shell
╰─ echo "e110c300ac0a0094bba84f57b4514b520aad082b47a266a58a2967b95ad49bcee16bac492658cc5494a50236d0a35d1e" | xxd -r -p | xxd -g 1

00000000: e1 10 c3 00 ac 0a 00 94 bb a8 4f 57 b4 51 4b 52  ..........OW.QKR
00000010: 0a ad 08 2b 47 a2 66 a5 8a 29 67 b9 5a d4 9b ce  ...+G.f..)g.Z...
00000020: e1 6b ac 49 26 58 cc 54 94 a5 02 36 d0 a3 5d 1e  .k.I&X.T...6..].
```

## 기타

- [xxd - Unix, Linux Command](https://www.tutorialspoint.com/unix_commands/xxd.htm)
