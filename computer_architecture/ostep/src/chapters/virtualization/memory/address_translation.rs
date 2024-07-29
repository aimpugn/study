#[inline(never)]
pub fn func_to_increase() -> i32 {
    let mut x = 3000; // assing 3000 to x
    x = x + 3; // line of code we are interested in

    return x;
}
