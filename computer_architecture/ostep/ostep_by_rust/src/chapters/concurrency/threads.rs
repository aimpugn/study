use std::{
    alloc::{dealloc, Layout},
    io::{self, Write},
    mem, ptr,
    ptr::null,
    sync::{Arc, Mutex},
    thread,
};

use libc::{pthread_attr_init, pthread_attr_t, pthread_create, pthread_join, pthread_t};

pub fn test_threads() {
    let stdout = Arc::new(Mutex::new(std::io::stdout()));
    {
        let mut guard = stdout.lock().unwrap();
        writeln!(&mut guard, "[Main] start").unwrap();
    }

    // thread::spawn(move || mythreads("B", Arc::clone(&stdout)));
    // Do not this way.                                ^^ variable moved due to use in closure

    // creates another pointer to the same allocation
    let cloned_stdout1 = Arc::clone(&stdout);
    let t1 = thread::spawn(move || mythreads("[T1] A", cloned_stdout1));

    // creates another pointer to the same allocation
    let cloned_stdout2 = Arc::clone(&stdout);
    let t2 = thread::spawn(move || mythreads("[T2] B", cloned_stdout2));

    // wiat for threads to finish
    t1.join().unwrap();
    t2.join().unwrap();

    {
        let mut guard = stdout.lock().unwrap();
        let ref_mut_dereference = &mut guard;
        writeln!(*ref_mut_dereference, "[Main] done").unwrap();
    }
}

fn mythreads(arg: &str, stdout: Arc<Mutex<std::io::Stdout>>) {
    // Acquires a mutex, blocking the current thread until it is able to do so.
    let mut guard = stdout.lock().unwrap();
    writeln!(&mut *guard, "{}", arg).unwrap();
}

static mut COUNTER: i32 = 0;
pub fn test_invalid_access_shared_data() {
    unsafe {
        println!("[Main] begin, COUNTER is {}", COUNTER);
    }

    let t1 = thread::spawn(move || test_invalid_access_shared_data_thread("T1"));
    let t2 = thread::spawn(move || test_invalid_access_shared_data_thread("T2"));

    t1.join().unwrap();
    t2.join().unwrap();

    unsafe {
        println!("[Main] Done, COUNTER is {}", COUNTER);
    }
}

fn test_invalid_access_shared_data_thread(arg: &str) {
    println!("{} Begin", arg);
    for _ in 1..(1e7 as i32) {
        // use of mutable static is unsafe and requires unsafe function or block
        // mutable statics can be mutated by multiple threads: aliasing violations or data races will cause undefined behavior
        unsafe {
            COUNTER += 1;
        }
    }
    println!("{} Done", arg);
}

pub unsafe fn test_ptrhead_create() {
    // (X) let mut native: *mut pthread_t = mem::zeroed(); // initialize the memory to zero(null)
    //             ^^^^^^ null pointer
    // 위처럼 초기화할 경우 해당 메모리 위치에 유효한 `pthread_t`이 존재하지 않는다
    // 그래서 나중에 이 값을 `pthread_join`에서 `*pthread_t` 역참조 해서 사용하려 하면
    // `segmentation fault` 같은 문제가 발생할 수 있다
    //
    // `mem::zeroed()`는 스택의 `native` 변수에 default 값 0으로 `pthread_t` 타입의 메모리 할당하고 초기화
    // 이때 rust는 `native` 변수에 해당 메모리 공간에 대한 소유권을 부여한다.
    let mut native = mem::zeroed::<pthread_t>();
    let mut attr: pthread_attr_t = mem::zeroed();
    assert_eq!(libc::pthread_attr_init(&mut attr), 0);

    let data = MyArgs { a: 1, b: 2 };
    let immutable_ref_to_data = &data;
    // 불변 참조 타입을 불변 원시 포인터 타입으로 변환 허용
    // 단 원시 포인터는 빌림 체커나 소유권 시스템 보호 받지 못하므로, 이 변환이 안전하다는 것 보장 필요
    let immutable_raw_pointer = immutable_ref_to_data as *const _;
    // 불변 원시 포인터 `*const`와 가변 원시 포인터 `*mut` 간의 변환 가능
    // 다른 스레드에서 접근 시 데이터 경쟁 상황 발생 가능
    let mutable_raw_pointer = immutable_raw_pointer as *mut libc::c_void;
    pthread_create(&mut native, &mut attr, mythread, mutable_raw_pointer);

    println!(
        "[outside of thread] before join a: {}, b: {}",
        data.a, data.b
    );

    let myreturn_heap_space = Box::new(MyReturn::default());
    let address_of_heap_alloc_myreturn = Box::into_raw(myreturn_heap_space);
    let my_return = address_of_heap_alloc_myreturn.cast::<*mut libc::c_void>();
    pthread_join(native, my_return);

    println!(
        "[outside of thread] after join a: {}, b: {}",
        data.a, data.b
    );

    // ostep(39592,0x1fb4a1b40) malloc: *** error for object 0x16b2de478: pointer being freed was not allocated
    // 위에서 `data: MyArgs`는 stack에 할당되고 있기 때문에
    // heap에 할당된 메모리를 해제하는 `Box::from_raw`로 메모리 해제하려고 하면 문제 발생
    // let _ = unsafe { Box::from_raw(mutable_raw_pointer) };

    // Manual cleanup by explicitly running the destructor
    // and deallocating the memory
    ptr::drop_in_place(address_of_heap_alloc_myreturn);
    dealloc(
        address_of_heap_alloc_myreturn as *mut u8,
        Layout::new::<String>(),
    );

    // Converting the raw pointer back into a Box with
    // `Box::from_raw` for automatic cleanup:
    // let _ = unsafe { Box::from_raw(address_of_heap_alloc_myreturn) };
}

extern "C" fn mythread(data: *mut libc::c_void) -> *mut libc::c_void {
    unsafe {
        let data = &mut *(data as *mut MyArgs);
        println!("[inside of thread] a: {}, b: {}", data.a, data.b);
        data.a = data.a * 10;
        data.b = data.b * 10;
    }
    ptr::null_mut()
}

struct MyArgs {
    a: i32,
    b: i32,
}

struct MyReturn {
    x: Option<i32>,
    y: Option<i32>,
}

impl Default for MyReturn {
    fn default() -> Self {
        Self {
            x: Default::default(),
            y: Default::default(),
        }
    }
}
