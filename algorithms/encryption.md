# Encryption

- [Encryption](#encryption)
    - [IV](#iv)

safecurves.cr.yp.to

nacl

nonce: 대칭 암호화에 쓰이는 값

## IV

- 키가 달라도 iv가 같으면, 암호화된 코드의 맨 첫 비트 몇 개가 같은 패턴을 반복한다
- iv는 매번 바꿔줘야 하는데, key와 다르게 숨길 필요 없음
- iv를 쉬운 말로 바꾼 게 nonce
- random 64 bit: 서버가 몇 년 단위로 오래 켜지면 충돌 위험이 유의미하게 높아져서
- random 96 bit: 이상 숫자를 사용
