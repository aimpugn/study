# Rust & Thread


Linux에서는 clone 시스템 콜을 사용하여 직접 스레드를 생성할 수 있습니다.
다음은 대략적인 예시입니다:
```rust
use std::arch::asm;
use libc::{c_void, CLONE_VM, CLONE_FS, CLONE_FILES, CLONE_SIGHAND, CLONE_THREAD, CLONE_SYSVSEM, CLONE_SETTLS, CLONE_PARENT_SETTID, CLONE_CHILD_CLEARTID};
use std::ptr;

extern "C" fn thread_entry(arg: *mut c_void) -> *mut c_void {
    // 스레드가 실행할 작업
    println!("Thread started with arg: {:?}", arg);
    ptr::null_mut() // 종료 시 반환값
}

fn create_thread(entry: extern "C" fn(*mut c_void) -> *mut c_void, arg: *mut c_void) -> isize {
    // 스택 크기 설정
    const STACK_SIZE: usize = 1024 * 1024; // 1MB 스택
    let stack = vec![0u8; STACK_SIZE]; // Rust에서는 Vec을 사용하여 메모리 할당
    let stack_ptr = stack.as_ptr().wrapping_add(STACK_SIZE); // 스택의 끝 주소

    unsafe {
        let mut tid: isize;
        asm!(
        "syscall",
        in("rax") 56,  // clone syscall number
        in("rdi") CLONE_VM | CLONE_FS | CLONE_FILES | CLONE_SIGHAND
            | CLONE_THREAD | CLONE_SYSVSEM | CLONE_SETTLS
            | CLONE_PARENT_SETTID | CLONE_CHILD_CLEARTID,
        in("rsi") stack_ptr,   // 스택 주소
        in("rdx") 0,           // parent_tidptr
        in("r10") 0,           // tls
        in("r8") entry as usize,
        in("r9") arg as usize,
        out("rax") tid,
        options(nostack)
        );
        if tid < 0 {
            eprintln!("clone syscall failed with error code: {}", tid);
        }
        tid
    }
}

fn main() {
    let arg = 42 as *mut c_void;
    let tid = create_thread(thread_entry, arg);
    if tid > 0 {
        println!("Thread created with tid: {}", tid);
    } else {
        println!("Failed to create thread.");
    }
}
```
