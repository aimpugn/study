use std::{
    alloc::{self, Layout, LayoutError},
    mem, process, thread, time,
};

/// ## [Allocating Memory](https://doc.rust-lang.org/nomicon/vec/vec-alloc.html)
///
/// global allocator is not allowed to allocate zero-sized memory block
///
/// - [alloc](https://doc.rust-lang.org/alloc/alloc/fn.alloc.html)
/// - [realloc](https://doc.rust-lang.org/alloc/alloc/fn.realloc.html)
/// - [dealloc](https://doc.rust-lang.org/alloc/alloc/fn.dealloc.html)
///
/// ## [Primitive Type pointer](https://doc.rust-lang.org/std/primitive.pointer.html)
///
/// ## ETC
/// - [Passing a memory address in Rust](https://stackoverflow.com/a/58011841)
pub fn print_memory_and_sleep() {
    // initialize value
    let mut val = 0;
    // get a ref of the value
    let val_ref = &val;
    // convert ref to raw pointer
    let val_ref_raw_ptr = val_ref as *const i32;
    // convert raw pointer to integer
    let val_addr = val_ref_raw_ptr as usize;
    loop {
        println!("[{}] address: {}, value: {}", process::id(), val_addr, val);
        val += 1;
        thread::sleep(time::Duration::from_secs(1));
        if val == 10 {
            break;
        }
    }
}

pub fn alloc_then_print_and_sleep() {
    unsafe {
        // `Layout::new::<u32>();`는 다음과 비슷한 절차를 거친다
        let alignment = mem::align_of::<u64>();
        let size = mem::size_of::<u64>();
        // `Layout`: particular layout of block of memory.
        let layout = alloc::Layout::from_size_align(size, alignment).unwrap();

        // `alloc` return raw pointer `*mut u8`
        let ptr = alloc::alloc_zeroed(layout);
        if ptr.is_null() {
            alloc::handle_alloc_error(layout);
        }
        // Cast pointer to a typed pointer
        // `u64` 값을 저장하고 싶다면, 해당 타입에 맞게 캐스팅한다
        let ptr = ptr as *mut u64;
        *ptr = 0; // `*mut u64`가 가리키는 공간을 dereference

        let mut cnt = 0;
        loop {
            *ptr += 10;
            println!(
                "[{}] address: {}, alignment_of_u64: {}, size_of_u64: {}, value: {}, read: {}, memory_location: {:p}, &ref: {:?}, as_ref: {}",
                process::id(),
                ptr as usize,
                alignment,
                size,
                *ptr,
                ptr.read(),
                ptr,
                &ptr,
                ptr.as_ref().unwrap()
            );

            if cnt == 10 {
                break;
            }
            thread::sleep(time::Duration::from_secs(1));
            cnt += 1;
        }
        alloc::dealloc(ptr as *mut u8, layout);
    }
}
