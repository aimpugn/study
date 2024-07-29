# Threads

```rs
use std::{
    io::{self, Write},
    sync::{Arc, Mutex},
    thread,
};


fn mythread(arg: &str, stdout: Arc<Mutex<std::io::Stdout>>) {
    let mut guard = stdout.lock().unwrap();
    writeln!(&mut *guard, "{}", arg).unwrap();
    ^^^^^^^^
    Write formatted data into a buffer, with a newline appended.
    In this case, string will be added to stodout
}

fn main() {
    let stdout = Arc::new(Mutex::new(std::io::stdout()));
    {
        let mut guard = stdout.lock().unwrap();
        writeln!(&mut *guard, "main: begin").unwrap();
                 ^^^^^^
                 dereferences and then re-references the mutable reference
    }

    let arg1 = "A";
    let arg2 = "B";

    let stdout_clone1 = Arc::clone(&stdout);
                                   ^^^^^^^ immutable reference
    let thread1 = thread::spawn(move || mythread(arg1, stdout_clone1));
                                ^^^^
                                used to change the way a closure captures its environment
                                the closure takes ownership of the values it uses from the environment.

    let stdout_clone2 = Arc::clone(&stdout);
    let thread2 = thread::spawn(move || mythread(arg2, stdout_clone2));

    thread1.join().unwrap();
            ^^^^^^ Waits for the associated thread to finish.
    thread2.join().unwrap();

    {
        let mut guard = stdout.lock().unwrap();
                ^^^^^ MutexGuard<Stdout>
        let ref_mut_dereference = &mut guard;
            ^^^^^^^^^^^^^^^^^^^ &mut MutexGuard<Stdout>
        writeln!(ref_mut_dereference, "[Main] done").unwrap();
                 ^^^^^^^^^^^^^^^^^^^
                 works because of auto deref coercion
                 This is done through the Deref and DerefMut traits.

    }
}

// #[stable(feature = "rust1", since = "1.0.0")]
// impl<T: ?Sized> DerefMut for MutexGuard<'_, T> {
//     fn deref_mut(&mut self) -> &mut T {
//         unsafe { &mut *self.lock.data.get() }
//     }
// }
```

In this version, the `mythread` function takes an additional argument, `stdout: Arc<Mutex<std::io::Stdout>>`, which is a shared, synchronized handle to the standard output.

A Mutex (short for "mutual exclusion") is a synchronization primitive that provides exclusive access to a shared resource. In Rust, it is part of the `std::sync` module. A `Mutex` ensures that only one thread can access the protected resource at a time. When one thread locks a `Mutex`, any other threads attempting to lock the same `Mutex` will block until the first thread releases the lock. This prevents race conditions and ensures that the output from different threads does not interleave.

In the provided code, we use an `Arc<Mutex<std::io::Stdout>>` to share a synchronized handle to the standard output (stdout) among multiple threads. `Arc` stands for **Atomically Reference Counted** and is a thread-safe reference-counting pointer. It allows multiple threads to share ownership of a value, and the value will be deallocated only when the last reference is dropped. Combining `Arc` with `Mutex` enables sharing mutable state across threads safely.

In the main function, we create an `Arc<Mutex<std::io::Stdout>>` by wrapping the standard output (`std::io::stdout()`) in a `Mutex` and then an `Arc`. We then clone this `Arc` for each thread, ensuring that they share the same synchronized handle to stdout.

Inside the `mythread` function, we lock the stdout mutex using the lock method. This method returns a `Result` that contains a `MutexGuard` if the lock is acquired successfully, or an error if the lock is poisoned (i.e., a thread panicked while holding the lock). We use the `unwrap` method to assume that the lock is not poisoned and obtain the MutexGuard.

A `MutexGuard` is a smart pointer that automatically *releases the lock when it goes out of scope*. In this case, we use the `MutexGuard` to write to the standard output, ensuring that only one thread at a time can perform this operation. As a result, the output from `mythread` will not interleave, providing accurate synchronization between threads.

> `MutexGuard`?  
>
> An RAII implementation of a "scoped lock" of a mutex. When this structure is dropped (falls out of scope), the lock will be unlocked.

In summary, using a Mutex and Lock in Rust provides a safe and efficient way to synchronize access to shared resources, ensuring that race conditions are avoided and that the output from different threads does not interleave.

```rs
writeln!(&mut *guard, "{}", arg).unwrap();
        ^^^^^^^^^^^^
        no method named `write_fmt` found for mutable reference `&mut Stdout` in the current scope items from traits can only be used if the trait is in scope
```

Due to the `Write` trait not being in scope. The `writeln!` macro requires the `Write` trait to be in the scope, as it uses the `write_fmt` method provided by the trait.

```rs
use std::io::{self, Write};

writeln!(&mut *guard, "{}", arg).unwrap();
```
