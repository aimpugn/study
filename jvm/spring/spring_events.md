# redis

## Spring Event

1. redis subscribe 받아와서
2. 내부적인 비즈니스 로직 후처리. 한 클래스 안에 있어서 보기가 어려워서, 이벤트 한번 더 퍼블리시 해서 리스닝 후 처리
   1. spring event
   2. @Async 사용하면 요청마다 새로운 쓰레드 생성되므로 별도 쓰레드 풀 설정 필요
      ?? 쓰레드 풀이 가득 차면? 작업 큐가 있어서 쓰레드 풀의 쓰레드가 하나씩 해소해 나간다

### WebFlux 기준

개선사항
- 맨 마지막에 이벤트를 전파하도록 한다

spring mvc, web flux