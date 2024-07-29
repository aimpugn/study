# Map and Entry

## `Map::entry`

```rs
impl Map<String, Value> {
...
    /// Gets the given key's corresponding entry in the map for in-place
    /// manipulation.
    pub fn entry<S>(&mut self, key: S) -> Entry
    where
    // 제약 조건
        S: Into<String>, // `S` 타입이 `String`으로 변환될 수 있어야 함을 의미
        // 이를 통해 `entry` 메서드는 `String`뿐만 아니라 
        // `String`으로 변환 가능한 다른 타입들(`&str` 등)도 키로 받을 수 있다.

    {
        #[cfg(not(feature = "preserve_order"))] // 컴파일 타임에 기능 플래그에 따라 코드가 활성화되도록 한다.
        // "preserve_order" 기능이 비활성화된 경우
        use alloc::collections::btree_map::Entry as EntryImpl;
        #[cfg(feature = "preserve_order")] // 컴파일 타임에 기능 플래그에 따라 코드가 활성화되도록 한다.
        // "preserve_order" 기능이 활성화된 경우
        use indexmap::map::Entry as EntryImpl;

        match self.map.entry(key.into()) {
            EntryImpl::Vacant(vacant) => Entry::Vacant(VacantEntry { vacant }),
            EntryImpl::Occupied(occupied) => Entry::Occupied(OccupiedEntry { occupied }),
        }
    }
...
}
```

## `Entry::or_insert_with`

```rs

impl<'a> Entry<'a> {
...
    /// Ensures a value is in the entry by inserting the result of the default
    /// function if empty, and returns a mutable reference to the value in the
    /// entry.
    ///
    /// # Examples
    ///
    /// ```
    /// # use serde_json::json;
    /// #
    /// let mut map = serde_json::Map::new();
    /// map.entry("serde").or_insert_with(|| json!("hoho"));
    ///
    /// assert_eq!(map["serde"], "hoho".to_owned());
    /// ```
    pub fn or_insert_with<F>(self, default: F) -> &'a mut Value
    where
        F: FnOnce() -> Value, // `F`가 `FnOnce` 트레잇을 구현하는 타입, 즉 호출 가능한 타입이어야 한다
        // 이 클로저가 아무 인자 없이 호출되며, `Value` 타입을 반환해야 함을 의미한다.
    {
        match self {
            Entry::Vacant(entry) => entry.insert(default()),
            Entry::Occupied(entry) => entry.into_mut(),
        }
    }
...
}
```

> Rust에서 `FnOnce`, `FnMut`, `Fn`은 클로저 또는 함수 포인터 같은 호출 가능한 타입들이 구현해야 하는 트레잇들이다
