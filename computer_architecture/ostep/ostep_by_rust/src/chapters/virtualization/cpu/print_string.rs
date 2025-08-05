use std::{hint, thread, time, time::Duration};

/// 원래 OSTEP 책에 예제로 나와 있던 코드
///
/// ```
/// if(argc!=2){
///     fprintf(stderr,“usage: cpu <string>\n”);
///     exit(1);
/// }
/// char *str = argv[1];
/// while (1) {
///     Spin(1);
///     printf(“%s\n”, str);
/// }
/// ```
pub fn print_string(c: String) {
    loop {
        spin_by_spinloop(1);
        println!("{}", c);
    }
}

/// [std::hint::spin_loop](https://doc.rust-lang.org/std/hint/fn.spin_loop.html) 함수는
/// 프로세서에게 현재 busy-wait spin-loop ("spin lock")에서 실행 중임을 알리는 기계 명령어를 발생시킵니다.
///
/// 스핀 루프 신호를 수신하면 프로세서는 전력을 절약하거나 하이퍼스레드를 전환하는 등 자신의 동작을 최적화할 수 있습니다.
///
/// 현재 스레드가 자발적으로 실행 시간을 포기하고 다른 스레드가 실행될 수 있도록 하는 [thread::yield_now](https://doc.rust-lang.org/std/thread/fn.yield_now.html)와 다릅니다.
/// `thread::yield_now` 함수는 운영 체제의 스케줄러와 상호작용하여 스케쥴러에게 직접 양보하지만,
/// `spin_loop`는 운영 체제와 상호 작용하지 않습니다.
///
/// spin_loop의 일반적인 사용 사례:
/// synchronization primitives의 CAS 루프 내에서 제한된 낙관적 스피닝을 구현하는 것입니다.
///
/// > [CAS(Compare-And-Swap)](https://en.wikipedia.org/wiki/Compare-and-swap)?
/// >
/// > 동시성 프로그래밍에서 동기화 위해 사용되는 원자적 명령어입니다.
/// > 현재 메모리의 값이 예상한 과거 값과 같다면, 새로운 값으로 교체합니다.
/// > 이 "같음"을 비교함으로써 멀티스레딩 환경에서 데이터의 일관성을 유지합니다.
/// >
/// > 낙관적 스피닝(Optimistic Spinning)?
/// >
/// > 락(lock)을 획득하기 전에 짧은 시간 동안 스핀(spin)하면서 락이 해제되기를 기다리는 기법입니다.
/// > 락이 곧 사용 가능해질 것이라는 '낙관적' 가정하에 동작합니다.
///
/// 우선순위 반전과 같은 문제를 피하기 위해, 스핀 루프를 유한한 횟수의 반복 후에 종료하고
/// 적절한 차단 시스템 호출을 하는 것이 강력히 권장됩니다.
///
/// > [우선순위 역전](https://en.wikipedia.org/wiki/Priority_inversion)?
/// >
/// > 높은 우선순위 태스크가 낮은 우선순위 태스크에 의해 간접적으로 차단되는 현상입니다.
/// > 실시간 시스템에서 심각한 문제를 일으킬 수 있습니다.
///
/// 주의: 스핀 루프 힌트 수신을 지원하지 않는 플랫폼에서는 이 함수가 아무 작업도 수행하지 않습니다.
///
/// - 이 구현의 경우에는 지정된 시간 동안 CPU를 100% 점유합니다.(CPU를 적극적으로 사용)
///   프로그램이 계속해서 명령을 실행하고 CPU는 계속해서 명령을 처리하느라 바쁩니다.
///   마치 엔진을 계속 공회전시키는 것과 비슷합니다.
/// - 다른 프로세스나 스레드가 CPU를 사용하기 어렵게 만듭니다.
/// - 실제 작업 부하를 시뮬레이션하는 데 유용합니다.
fn spin_by_spinloop(seconds: u64) {
    let start = std::time::Instant::now();
    while start.elapsed() < std::time::Duration::from_secs(seconds) {
        // The spin loop is a hint to the CPU that we're waiting,
        // but probably not for very long
        std::hint::spin_loop();
    }
}

/// - `spin_by_spinloop` 함수와 달리 CPU를 거의 사용하지 않습니다.
///   프로그램이 운영체제에 "잠시 쉬겠다"고 알리고,
///   CPU는 다른 작업을 처리하거나 전력 절약 모드로 들어갈 수 있습니다.
///   이는 엔진을 끄고 나중에 다시 켜는 것과 비슷합니다.
/// - 운영 체제의 스케줄링에 따라 약간의 지연 가능 하므로, 정확한 타이밍을 보장하지 않을 수 있습니다.
/// - 실제 CPU 사용을 시뮬레이션하지 않습니다.
fn spin_by_sleep(seconds: u64) {
    thread::sleep(Duration::from_secs(seconds));
}

