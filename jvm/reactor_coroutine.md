# redis

## reactor, coroutine의 차이, 특징

webflux, reactor, coroutine

reactive programming
- api 호출 후 결과를 대기하고 받는 게 아님
- stream을 구독해서 데이터를 응답 받는다
- async

Reactor

- Mono
    - 0~1개만 데이터 발행
    - 보통 API처럼 하나의 응답을 하면 끝내는 경우
- Flux
    - 지속적인 이벤트 처리
    - 스트림 처리 등

Coroutine

쓰레드는 아니고, 쓰레드 안에서 돌아가는 작업
경량 쓰레드
한 쓰레드에 종속적이지 않다 -> 다른 쓰레드에서 실행될 수도 있다
!!!context switching 없이 코루틴 바꿔가면서 실행할 수 있음

redis 쓴 이유?
입금 통보 후 상태 변경이 될 수 있어서, 2분 정도 지연 통지? reactor, subscribe 사용

구독을 하고 있는데 노드가 여러 대라 서로 다른 노드

key space 이벤트에 대해서는 하나의 레디스처럼 동작하는 게 아니라, 어떤 노드에 저장되어 있는지 정확히 알 수가 없다... 지연 통보는 당장은 구현되어 있지 않음
