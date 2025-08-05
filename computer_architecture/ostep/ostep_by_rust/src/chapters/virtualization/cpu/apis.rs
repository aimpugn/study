use fork::{
    daemon, fork,
    Fork::{Child as ForkChild, Parent as ForkParent},
};
use libc::STDOUT_FILENO;
use nix::{
    errno::{self, Errno},
    fcntl::OFlag,
    sys::stat::Mode,
    sys::wait::{waitpid, WaitPidFlag, WaitStatus},
    unistd::{
        self, execv, execve, execvp, fork as nix_fork, getpid, getppid, write,
        ForkResult::{Child as NixChild, Parent as NixParent},
    },
};
use std::{
    borrow::Borrow,
    env,
    ffi::{CStr, CString},
    os::unix::prelude::IntoRawFd,
    path::{Path, PathBuf},
    process, thread, time,
};

pub fn test_fork_by_fork() -> i32 {
    // 3. child process does NOT start running at here
    println!("hello world: (pid: {})", process::id());

    // 1. system call to create new process
    let return_value = fork();
    // 2. ALMOST exact copy of the calling process starts running at here
    //    now child has its own register, PC, and so forth from here
    //    - `return_value` in parent: PID of child process
    //    - `return_value` in child: 0

    // 4. Because there are two processes, since now, which of process
    //    will be executed is determined by OS scheduler(non-deterministic)
    match return_value {
        // -1
        Err(_) => println!("Fork failed"),
        // 0
        Ok(ForkChild) => {
            // println!("hello, I am child: {}", process::id());
            write(
                libc::STDOUT_FILENO,
                format!("hello, I am child: {}\n", process::id()).as_bytes(),
            )
            .ok();
            unsafe { libc::_exit(0) };
        }
        // else
        Ok(ForkParent(child)) => {
            // parent goes down this path (main)
            println!(
                "hello, I am parent(pid: {}) of child {}",
                process::id(),
                child,
            );
        }
    }
    0
}

pub fn test_fork_by_nix() {
    println!("hello world: (pid: {})", process::id());
    // libc::fork() 사용
    match unsafe { nix_fork() } {
        // -1
        Err(_) => println!("Fork failed"),
        // 0
        Ok(NixChild) => {
            // Unsafe to use `println!` (or `unwrap`) here. See Safety.
            write(
                libc::STDOUT_FILENO,
                format!("hello, I am child: {}\n", process::id()).as_bytes(),
            )
            .ok();
            unsafe { libc::_exit(0) };
        }
        // else
        Ok(NixParent { child, .. }) => {
            // parent goes down this path (main)
            println!(
                "hello, I am parent(pid: {}) of child {}",
                process::id(),
                child,
            );
        }
    }
}

/// It will print string like below to terminal
///
/// ```log
/// hello world: (pid: 54696)
/// hello, I am child: 54706
/// [0] sleep 1 sec in child(pid: 54706)
/// [1] sleep 1 sec in child(pid: 54706)
/// [2] sleep 1 sec in child(pid: 54706)
/// [3] sleep 1 sec in child(pid: 54706)
/// [4] sleep 1 sec in child(pid: 54706)
/// [5] sleep 1 sec in child(pid: 54706)
/// [6] sleep 1 sec in child(pid: 54706)
/// [7] sleep 1 sec in child(pid: 54706)
/// [8] sleep 1 sec in child(pid: 54706)
/// [9] sleep 1 sec in child(pid: 54706)
/// child 54706, pid 54706 is exited, exit code is 0
/// hello, I am parent(pid: 54696) of child 54706
/// ```
pub fn test_fork_by_nix_wait_child() {
    println!("hello world: (pid: {})", process::id());
    match unsafe { nix_fork() } {
        // -1
        Err(_) => println!("Fork failed"),
        // 0
        Ok(NixChild) => {
            // Unsafe to use `println!` (or `unwrap`) here. See Safety.
            write_stdout(format!("hello, I am child: {}", process::id()));
            for _i in 0..10 {
                thread::sleep(time::Duration::from_millis(1000));
                write_stdout(format!(
                    "[{}] sleep 1 sec in child(pid: {})",
                    _i,
                    process::id()
                ));
            }
            unsafe { libc::_exit(0) };
        }
        // else
        Ok(NixParent { child, .. }) => {
            // Parent might run first, but by calling `waitpid`,
            // parent process will wait child process to be done
            let result = waitpid(
                child,
                Some(
                    WaitPidFlag::WCONTINUED // // parent  -> child -> parent
                    // WaitPidFlag::WEXITED, // EINVAL: Invalid argument
                    // WaitPidFlag::WNOHANG, // child 21482 is still alive.
                    // WaitPidFlag::WNOWAIT, // EINVAL: Invalid argument
                    // WaitPidFlag::WSTOPPED, // EINVAL: Invalid argument
                    | WaitPidFlag::WUNTRACED, // parent  -> child -> parent
                ),
            );
            match result {
                Ok(wait_status) => {
                    // https://docs.rs/nix/latest/nix/sys/wait/enum.WaitStatus.html
                    match wait_status {
                        WaitStatus::Exited(pid, exit_code) => write_stdout(format!(
                            "child {}, pid {} is exited, exit code is {}",
                            child, pid, exit_code
                        )),
                        WaitStatus::Signaled(pid, signal, core_dump) => write_stdout(format!(
                            "child {}, pid {} is signaled {}. The signal generated a core dump? {}",
                            child, pid, signal, core_dump
                        )),
                        WaitStatus::Stopped(pid, signal) => write_stdout(format!(
                            "child {}, pid {} is stopped. Signal is {}.",
                            child, pid, signal
                        )),
                        WaitStatus::Continued(pid) => write_stdout(format!(
                            "Stopped child {}, pid {} has resumed execution after receiving a `SIGCONT` signal.",
                            child, pid
                        )),
                        WaitStatus::StillAlive => write_stdout(format!("child {} is still alive.", child)),
                    }
                }
                Err(errno) => write_stdout(format!("error: {}", errno)),
            }

            // After child process done, then parent process do its operation
            println!(
                "hello, I am parent(pid: {}) of child {}",
                process::id(),
                child,
            );
        }
    }
}

pub fn test_exec_wc() {
    println!("hello world: (pid: {})", process::id());

    match unsafe { nix_fork() } {
        Err(_) => println!("Fork failed"),
        Ok(NixChild) => {
            write_stdout(format!("hello, I am child: {}", process::id()));
            /*
             * let words = execvp(
             *    // "\0" terminated C type string bytes
             *    CStr::from_bytes_with_nul(b"wc\0").unwrap(),
             *    &[
             *        CStr::from_bytes_with_nul(b"-cl\0").unwrap(),
             *        CStr::from_bytes_with_nul(b"/Users/rody/VscodeProjects/ostep/Dockerfile\0")
             *            .unwrap(),
             *    ],
             * );
             */
            let words = execvp(
                // filename of program to be run
                &str_to_c_string("wc"),
                // args will be passed to the program
                &[
                    str_to_c_string("-cl"), // 안 먹히는듯?
                    str_to_c_string(
                        (env::current_dir().unwrap().as_path().display().to_string()
                            + "/Cargo.toml")
                            .as_str(),
                    ),
                ],
            );

            match words {
                Ok(infallible) => write_stdout(format!("infallible: {}", infallible)),
                Err(errno) => write_stdout(format!("errno: {}", errno)),
            }

            unsafe { libc::_exit(0) };
        }
        Ok(NixParent { child, .. }) => {
            let result = waitpid(child, None);
            match result {
                Ok(_) => write_stdout(format!(
                    "hello, I am parent(pid: {}) of child {}",
                    process::id(),
                    child,
                )),
                Err(errno) => write_stdout(format!("errorno is {}", errno)),
            }
        }
    }
}

pub fn test_exec_output_redirect() {
    println!("hello world: (pid: {})", process::id());
    // libc::fork() 사용
    match unsafe { nix_fork() } {
        Err(_) => println!("Fork failed"),
        Ok(NixChild) => {
            // child: redirect standard output to a file
            // close stdout
            match unistd::close(STDOUT_FILENO.into_raw_fd()) {
                Ok(()) => write_stdout("STDOUT closed".to_string()),
                Err(errno) => write_stdout(format!("Failed to close STDOUT, becuase of {}", errno)),
            }

            // open file where the output is redirected to
            unsafe {
                let pathbuf = pathbuf_of(&["tmp", "test_exec_output_redirect.output"]);
                // `STDOUT_FILENO`가 첫번째 사용 가능한 파일 디스크립터가 될 것이고, open 시 할당된다
                nix::libc::open(
                    str_to_c_string(pathbuf.display().to_string().as_str()).as_ptr(),
                    // https://docs.rs/nix/latest/nix/fcntl/struct.OFlag.html
                    OFlag::O_CREAT.bits() | OFlag::O_WRONLY.bits() | OFlag::O_TRUNC.bits(),
                    // https://docs.rs/nix/latest/nix/sys/stat/struct.Mode.html
                    // https://devanix.tistory.com/289
                    // S_IRUSR | S_IWUSR | S_IXUSR
                    Mode::S_IRWXU,
                );
            }

            // the output of this command will be redirected
            // to the file opened above
            // Why? 이런 리디렉션이 작동하는 이유는 OS가 file desciptor를 어떻게 관리하는지에 대한 가정 덕분
            // 1. 0부터 file descriptor를 찾기 시작.
            //    이 경우 `open()` 호출 시 `STDOUT_FILENO`가 사용 가능한 첫번째 파일 디스크립터가 되어 할당된다
            // 2. 그 다음 이어지는 자식 프로세스의 `printf()` 같은 루틴 통한 standard output 파일 디스크립터에 대한 쓰기는
            //    스크린이 아닌 새롭게 열린 파일로 라우팅 된다
            let result = execvp(
                &str_to_c_string("wc"),
                &[
                    str_to_c_string("-cl"),
                    str_to_c_string(pathbuf_of(&["Cargo.toml"]).display().to_string().as_str()),
                ],
            );

            match result {
                Ok(infallible) => write_stdout(format!("infallible: {}", infallible)),
                Err(errno) => write_stdout(format!("errno: {}", errno)),
            }

            unsafe { libc::_exit(0) };
        }
        // else
        Ok(NixParent { child, .. }) => {
            println!(
                "hello, I am parent(pid: {}) of child {}",
                process::id(),
                child,
            );
        }
    }
}

pub fn test_change_variable_from_child() {
    let mut x = 10;
    println!("hello world: (pid: {}). x is {}", process::id(), x);
    match unsafe { nix_fork() } {
        Err(_) => println!("Fork failed"),
        Ok(NixChild) => {
            x = 100;
            write_stdout(format!("hello, I am child: {}. x is {}", process::id(), x));

            unsafe { libc::_exit(0) };
        }
        // else
        Ok(NixParent { child, .. }) => {
            let result = waitpid(child, None);
            match result {
                Ok(wait_status) => println!("Child(pid: {}) exited", wait_status.pid().unwrap()),
                Err(errno) => println!("Errno is {}", errno),
            }

            println!(
                "hello, I am parent(pid: {}) of child {}. x is {}",
                process::id(),
                child,
                x // Not changed, so 10
            );
        }
    }
}

// Unsafe to use `println!` (or `unwrap`) in forked child process. See Safety.
fn write_stdout(message: String) {
    write(libc::STDOUT_FILENO, (message + "\n").as_bytes()).ok();
}

/// `CString` is a wrapper around a byte buffer that contains a null-terminated string,
/// while a `CStr` is a view into that buffer that doesn't own the underlying memory.
/// So can't create an owned `CStr` object directly
///
/// - /Users/rody/.rustup/toolchains/stable-aarch64-apple-darwin/lib/rustlib/src/rust/library/
///     - core/src/ffi/c_str.rs: `CStr` defined
///     - alloc/src/ffi/c_str.rs: `CString` defined
fn str_to_c_string(s: &str) -> CString {
    CString::new(s).unwrap()
}

fn pathbuf_of(paths: &[&str]) -> PathBuf {
    let mut current_dir_pathbuf = env::current_dir().unwrap();
    //  ^^^ mutable to add path `&str`s
    if paths.len() > 0 {
        for path in paths {
            current_dir_pathbuf.push(path);
        }
        current_dir_pathbuf
    } else {
        env::current_dir().unwrap()
    }
}
