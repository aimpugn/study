#![allow(dead_code, unused_imports)]

// here `src/main.rs` is crate root of binary
use std::{collections::HashMap, env};

// tell compiler to include the code it finds in `src/chapters.rs`
pub mod chapters;
pub mod util;

// use public modules from `src/chapters/introduction/mod.rs`
use chapters::{introduction::*, virtualization};
use util::args::{ArgsMap, Config};

fn main() {
    let _args = Config.parse();

    // threads();
    print_characters()
}

fn print_characters() {
    let args: Vec<String> = env::args().collect();
    let str = args.get(1);
    match str {
        Some(arg) => virtualization::cpu::print_string::print_string(arg.to_owned()),
        None => print!("{}", "No argument provided"),
    }
}

fn introduction() {
    // cpu::print_and_sleep(_args.get("--param").unwrap().to_string());

    // memory::print_memory_and_sleep();
    // memory::alloc_then_print_and_sleep();
    // concurrency::wrongly_working_concurrency_program(100_000);
    // persistence::write_hello_world();
}

fn process() {
    // process_run();
    // process::apis::test_fork_by_fork();
    // process::apis::test_fork_by_nix();
    virtualization::cpu::apis::test_fork_by_nix_wait_child();
    // process::apis::test_exec_wc();
    // process::apis::test_exec_output_redirect();
    // process::apis::test_change_variable_from_child();
}

fn process_run() {
    // process
    let s = virtualization::cpu::process_run::Scheduler {
        proc_info: HashMap::new(),
        process_switch_behavior: String::from(""),
        io_done_behavior: String::from(""),
        program: vec!["c7"],
        io_length: 0,
    };
    s.init();
    s.run();
}

fn lottery() {
    virtualization::cpu::lottery::simple_lottery_scheduling();
}

fn memory() {
    // virtualization::memory::heap::test_alloc();
    // virtualization::memory::heap::test_malloc();
    virtualization::memory::address_translation::func_to_increase();
}

fn threads() {
    // chapters::concurrency::threads::test_threads();
    // chapters::concurrency::threads::test_invalid_access_shared_data();
    unsafe { chapters::concurrency::threads::test_ptrhead_create() };
}
