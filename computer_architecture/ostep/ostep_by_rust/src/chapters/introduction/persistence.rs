use std::env;
use std::fs::File;
use std::io::Write;
use std::path::Path;

pub fn write_hello_world() {
    let write_to_result = write_to(
        env::current_dir().unwrap().display().to_string().to_owned() + "/tmp/hello_world.txt",
        String::from("Hello! World\n"),
    );
    if write_to_result.is_err() {
        // ex) Write to error! No such file or directory (os error 2)
        println!(
            "Write to error! {}",
            write_to_result.err().unwrap().to_string()
        )
    }
}

fn write_to(path: String, contents: String) -> std::io::Result<()> {
    // 1. open
    let mut f = File::options().create(true).append(true).open(path)?;
    // 2. write
    let result = f.write_all(contents.as_bytes());
    // 3. close:
    // - Files are automatically closed when they go out of scope
    // -  Errors detected on closing are ignored by the implementation of `Drop`.
    //    Use the method `sync_all` if these errors must be manually handled.
    if result.is_ok() {
        Ok(())
    } else {
        Err(result.err().unwrap())
    }
}
