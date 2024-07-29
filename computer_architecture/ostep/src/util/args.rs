use std::collections::HashMap;
use std::env;

// `pub(crate)`: only public within the crate, and cannot be re-exported outside
pub struct Config;

// Defining a Trait
// - https://doc.rust-lang.org/book/ch10-02-traits.html#defining-a-trait
pub trait ArgsMap {
    // https://doc.rust-lang.org/error-index.html#E0599
    fn parse(self: &Self) -> HashMap<String, String>;
    fn print_args_with_delimeter(args: &Vec<String>, delimeter: Option<&str>);
}

// Implementing a Trait on a Type
// - https://doc.rust-lang.org/book/ch10-02-traits.html#implementing-a-trait-on-a-type
impl ArgsMap for Config {
    // `&self`
    // - 없으면 외부에서 호출이 안 된다
    // - `&self` is sugar for `self: &Self`
    fn parse(&self) -> HashMap<String, String> {
        // HashMap: https://doc.rust-lang.org/std/collections/struct.HashMap.html
        // arguments 받기
        // - https://doc.rust-lang.org/book/ch12-01-accepting-command-line-arguments.html#saving-the-argument-values-in-variables
        // iterator -> collection
        // - https://doc.rust-lang.org/std/iter/trait.Iterator.html#method.collect
        let args_collect: Vec<String> = env::args().collect(); // === let args_collect = env::args().collect::<Vec<String>>();
        let args_len = args_collect.len();
        let mut map: HashMap<String, String> = HashMap::new();
        if args_len > 0 {
            for mut arg in args_collect {
                arg = arg.trim().to_string();
                /*
                   // temporary value dropped while borrowed [E0716]
                   // creates a temporary which is freed while still in use
                   // Note: consider using a `let` binding to create a longer lived value
                   // to_owned() 후 변수 할당하지 않고 바로 split 하면
                   // 아직 소유권이 정해지지 않은 상태이므로 위와 같은 에러 발생
                   args_split = arg.to_owned().split_once("=").unwrap_or(("", ""));
                */
                // 현재 블록에서 사용될 arg의 소유권을 새로운 변수에 할당
                let arg_owned = arg.to_owned();
                // 소유권이 할당된 변수를 split한다
                let args_split = arg_owned.split_once("=").unwrap_or(("", ""));

                if args_split.0.ne("") && args_split.1.ne("") {
                    map.insert(String::from(args_split.0), String::from(args_split.1));
                }
            }
        }

        // return 생략 가능
        map
    }

    #[allow(dead_code)] // suppress "warning: function `print_args_with_delimeter` is never used"
    fn print_args_with_delimeter(args: &Vec<String>, delimeter: Option<&str>) {
        // 함수에 기본값 사용하기
        // - https://stackoverflow.com/a/35369909
        let del = delimeter.unwrap_or(",").to_string();
        let mut line = String::from("");
        // 문자열 연결
        // - https://doc.rust-lang.org/std/string/struct.String.html#impl-Add%3C%26str%3E-for-String
        for arg in args {
            line.push_str(format!("{} {}", arg, del).as_str());
            // line = format!("{} {} {}", line, del, arg);
        }
        println!("{}", line)
    }
}
