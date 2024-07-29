use libc;
use std::alloc::{alloc, dealloc, Layout};
use std::borrow::Borrow;
use std::mem::{self, size_of_val};

// https://stackoverflow.com/a/56790193/8562273
pub fn test_alloc() {
    unsafe {
        let x_layout = Layout::new::<i32>();
        let x_ptr = alloc(x_layout);
        println!("[test_alloc] x_ptr is {:?}", x_ptr);
        dealloc(x_ptr, x_layout);
    };
}

pub fn test_malloc() {
    unsafe {
        let i32_size_t = mem::size_of::<i32>();
        let f64_size_t = mem::size_of::<f64>();
        let x_ptr = libc::malloc(i32_size_t);
        let d_ptr = libc::malloc(f64_size_t);
        if x_ptr.is_null() || d_ptr.is_null() {
            panic!("failed to allocate memory");
        } else {
            println!(
                "[test_malloc] i32_size_t is {}, x_ptr is {:?}",
                i32_size_t, x_ptr
            );
            println!(
                "[test_malloc] f64_size_t is {}, d_ptr is {:?}",
                f64_size_t, d_ptr
            );
        }
        let vec_with_capa: Vec<i32> = Vec::with_capacity(10);
        println!(
            "size of reference of Vec::with_capacity(10) is {}",
            mem::size_of_val(&vec_with_capa)
        );
        println!(
            "size of Vec::with_capacity(10) is {}",
            mem::size_of_val(&*vec_with_capa)
        );
        let vec: Vec<i32> = vec![0; 10];
        println!(
            "size of reference of vec![0; 10] is {}",
            mem::size_of_val(&vec)
        );
        println!("size of vec![0; 10] is {}", mem::size_of_val(&*vec));

        libc::free(x_ptr);
        libc::free(d_ptr);
    }
}
