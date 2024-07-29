# jq examples

- [jq examples](#jq-examples)
    - [인자로 json 전달하여 출력하기](#인자로-json-전달하여-출력하기)
    - [배열로 만들기](#배열로-만들기)
        - [나열되는 문자열들을 배열로 만들기](#나열되는-문자열들을-배열로-만들기)
    - [length 측정하기](#length-측정하기)
    - [오브젝트로 만들기](#오브젝트로-만들기)
    - [csv 형태로 출력하기](#csv-형태로-출력하기)
    - [csv를 JSON 형태로 출력하기](#csv를-json-형태로-출력하기)

## 인자로 json 전달하여 출력하기

## 배열로 만들기

```bash
defaults export com.apple.symbolichotkeys - | plutil -convert json -o - - |
  jq '[.AppleSymbolicHotKeys["61"], .AppleSymbolicHotKeys["60"]]'
```

### 나열되는 문자열들을 배열로 만들기

```bash
cat mapper.json | jq '.[] | .merchant_uid'

"a"
"b"
"c"
"d"
...
```

이를 아래처럼 만들고 싶다면

```json
[
    "a",
    "b",
    "c",
    "d"
]
```

아래처럼 `[]` 로 감싼다.

```shell
cat mapper.json | jq '[.[] | .merchant_uid]'
```

## length 측정하기

> count json array using jq

To count the number of elements in a JSON array using `jq`, you can use the `length` filter.

Assuming you have a JSON file data.json with the following content:

```json
[
  {"name": "apple", "color": "red"},
  {"name": "banana", "color": "yellow"},
  {"name": "orange", "color": "orange"}
]
```

You can count the number of elements in the array like this:

```bash
jq 'length' data.json
# 3
```

특정 조건에 해당하는 요소의 수를 세고 싶다면, `select` 필터를 사용합니다.

color 속성이 red인 오브젝트의 수를 세고 싶은 경우:

```bash
# - `map(select(.color == "red"))`: creates a new array with only the objects where the color is "red".
# - `length`: then counts the number of elements in that filtered array.
jq 'map(select(.color == "red")) | length' data.json
# 1

jq 'map(select(.color == "red" or .color == "orange")) | length' data.json
# 2
```

These examples assume you're working with a JSON array as input. If your input is a stream of JSON objects
(one object per line), you can use the -n option with inputs to handle it: [2]

```bash
cat data.json | jq -n 'inputs | map(select(.color == "red")) | length'
```

This will count the number of objects where color is "red" from the stream of JSON objects.

## 오브젝트로 만들기

## csv 형태로 출력하기

`some_file.json`이라는 json 파일에 아래와 같은 데이터가 있다고 가정합니다.

```json
[
    {
        "some_key": {
            "col1": "1234",
            "col2": "USD",
            "col3": "some_id1",
            "col4": {
                "nested_col1": "nested_value"
            },
            "col5": [
                {
                    "nested_obj_id": "nested_obj_id_1"
                }
            ]
        }
    }
]
```

```sh
jq -r '.[] | ."some_key" | [.col1, .col2, .col3, .col5[].nested_obj_id] | @csv' some_file.json
# "1234","USD","some_id1","nested_obj_id_1"
```

## csv를 JSON 형태로 출력하기

```sh
❯ echo '"1234","USD","some_id1","nested_obj_id_1"' | jq -R 'split("\n")'
[
  "\"1234\",\"USD\",\"some_id1\",\"nested_obj_id_1\""
]
```

```sh
❯ echo '"1234","USD","some_id1","nested_obj_id_1"' | jq -R 'split("\n") | map(split(","))'
[
  [
    "\"1234\"",
    "\"USD\"",
    "\"some_id1\"",
    "\"nested_obj_id_1\""
  ]
]
```

```sh
❯ echo '"1234","USD","some_id1","nested_obj_id_1"' | jq -R 'split("\n") | map(split(",")) | map(map(gsub("\""; "")))'
[
  [
    "1234",
    "USD",
    "some_id1",
    "nested_obj_id_1"
  ]
]

# 또는
❯ echo '"1234","USD","some_id1","nested_obj_id_1"' | jq -R 'split("\n") | map(split(",")) | .[] | map(gsub("\""; ""))'
[
  "1234",
  "USD",
  "some_id1",
  "nested_obj_id_1"
]
```

```sh
❯ echo '"1234","USD","some_id1","nested_obj_id_1"' | jq -R 'split("\n") | map(split(",")) | map(map(gsub("\""; ""))) | map({"col1": .[0], "col2": .[1], "col3": .[2]})'
[
  {
    "col1": "1234",
    "col2": "USD",
    "col3": "some_id1"
  }
]
```

```sh
❯ echo '"1234","USD","some_id1","nested_obj_id_1"' | jq -R 'split("\n") | map(split(",")) | map(map(gsub("\""; ""))) | map({"col1": .[0], "col2": .[1], "col3": .[2]}) | {data: .}'
{
  "data": [
    {
      "col1": "1234",
      "col2": "USD",
      "col3": "some_id1"
    }
  ]
}
```

Here's a breakdown of the command:

1. cat input.csv reads the CSV file and pipes the content to jq.
2. -R option tells jq to read the input as raw strings.
3. -s option tells jq to read the input as a stream of objects.
4. split("\n") splits the input by newline characters, creating an array of lines.
5. map(split(",")) splits each line by commas, creating an array of fields for each line.
6. map(map(gsub("\""; ""; .))) removes the double quotes from each field.
7. map({"col1": .[0], "col2": .[1], "col3": .[2]}) creates an object for each line, using the field values as the values for the corresponding keys.
8. {data: .} wraps the array of objects in a top-level data key.

To use this command, save your CSV data in a file (e.g., input.csv), and then run the command:

zsh
cat input.csv | jq -R -s '
  split("\n") |
  map(split(",")) |
  map(map(gsub("\""; ""; .))) |
  map({"col1": .[0], "col2": .[1], "col3": .[2]}) |
  {data: .}
'

The output will be a JSON object with a data key containing an array of objects, where each object represents a row from the CSV file.

json
{
  "data": [
    {
      "col1": "123",
      "col2": "USD",
      "col3": "some_id"
    },
    {
      "col1": "1234",
      "col2": "USD",
      "col3": "some_id2"
    }
  ]
}

Note: This command assumes that your CSV file has a consistent number of columns for each row. If the number of columns varies, you may need to
modify the command accordingly.

1 <https://stackoverflow.com/questions/49228126>
