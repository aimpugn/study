use std::thread;

/// 스레드를 테스트합니다.
///
/// 새로운 스레드 스폰시 [thread::Builder::spawn_unchecked_]를 호출하게되고,
/// 최종적으로 [std::sys::pal::thread::Thread::new]를 호출합니다.
/// 그리고 `libc::pthread_create`를 호출하여 스레드를 생성합니다.
/// 즉, 내부적으로는 `pthread` (POSIX 스레드) 라이브러리를 사용하여 스레드를 생성합니다.
fn main() {
    // JoinHandle
    let thrd = thread::spawn(|| {
        println!("1. Hello, World! from thread");
    });

    match thrd.join() {
        Ok(ok) => {
            println!("2.1. join success, ok is {:?}", ok) // join success, ok is ()
        }
        Err(e) => {
            println!("2.2. failed to join, error is {:?}", e)
        }
    }

    builder_and_join_handle()
}


/// [thread::Builder::spawn]의 결과 [thread::JoinHandle]를 얻을 수 있습니다.
/// [thread::JoinHandle]은 [thread] 통해 생성된 새 스레드에 대한 참조를 유지합니다.
/// 타입 이름 그대로 새 스레드가 호출한 스레드로 다시 join 할 수 있도록 다루기 위한 타입입니다.
///
/// `join`은 한 스레드가 다른 스레드의 완료를 기다리는 동작을 의미합니다.
/// [thread::JoinHandle::join]을 호출하면 다음과 같은 동작이 이뤄집니다:
/// 1. 호출한 스레드(주로 메인 스레드)가 [thread::JoinHandle]이 참조하는 스레드의 완료를 기다립니다.
/// 2. [thread::JoinHandle::join]을 호출한 스레드(여기서는 메인 스레드)는 대상 스레드가 종료될 때까지 블로킹(실행 중지) 상태가 됩니다.
/// 3. 대상 스레드가 완료되면, 그 시점에서 호출 스레드의 실행이 재개됩니다.
fn builder_and_join_handle() {
    // io::Result<JoinHandle<T>>
    let thrd2 = thread::Builder::new()
        // .stack_size() 새로 생성할 스택의 사이즈 조절
        .name("thrd2".to_string())
        .spawn(|| {
            println!("3.2. Hello, new thread via builder");
        });

    match thrd2 {
        Ok(join_handle) => {
            println!("3.1. thrd2 spawned successfully");
            match join_handle.join() {
                Ok(ok) => {
                    println!("3.3.1. thrd2.join success, ok is {:?}", ok)
                }
                Err(e) => {
                    println!("3.3.2. thrd2.join failed, error is {:?}", e)
                }
            }
        }
        Err(e) => {
            println!("thrd2 spawn failed, error is {:?}", e);
        }
    };
}
