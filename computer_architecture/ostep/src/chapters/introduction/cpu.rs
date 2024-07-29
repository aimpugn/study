use std::{thread, time};

// make `item` public
pub fn print_and_sleep(str: String) {
    let mut cnt = 0;
    loop {
        cnt += 1;
        println!("[{}]{}", cnt, str);
        let ten_millis = time::Duration::from_secs(1);
        thread::sleep(ten_millis);
        if cnt == 10 {
            break;
        }
    }
}
