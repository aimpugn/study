use std::thread;

static mut WORKER_LOOPS: i32 = 0;
static mut COUNTER: i32 = 0;

pub fn wrongly_working_concurrency_program(loops: i32) {
    unsafe {
        WORKER_LOOPS = loops;
        let first = thread::spawn(wrongly_working_concurrency_worker);
        let second = thread::spawn(wrongly_working_concurrency_worker);
        // Waits for the associated thread to finish.
        let res_first = first.join();
        let res_second = second.join();
        if res_first.is_ok() {
            println!("first done");
        }
        if res_second.is_ok() {
            println!("second done");
        }
        println!("Final value of COUNTER: {}", COUNTER);
    }

    // data was `move`d to the spawned thread,
    // so we cannot use it here
}

/// It wrongly works because
/// the instructions(1, 2, 3) do not execute `atomically`
fn wrongly_working_concurrency_worker() {
    unsafe {
        for _i in 1..=WORKER_LOOPS {
            // 1. load the value of the counter(shared) from memory into a register
            COUNTER += 1; // 2.one to increment it and, 3. store it back into memory
        }
    }
}
