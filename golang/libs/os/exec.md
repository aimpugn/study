# exec

## `Command`

- 예제

    ```go
    cmd := exec.Command("tr", "a-z", "A-Z")
    cmd.Stdin = strings.NewReader("some input")
    var out strings.Builder
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("in all caps: %q\n", out.String())

    cmdWithEnv := exec.Command("prog")
    cmdWithEnv.Env = append(os.Environ(),
        "FOO=duplicate_value", // ignored
        "FOO=actual_value",    // this value is used
    )
    if err := cmdWithEnv.Run(); err != nil {
        log.Fatal(err)
    }
    ```

- `go env` 결과를 출력해보기

    ```go
    var out, stderr bytes.Buffer

    cmd := exec.Command("go", "env")
    cmd.Stderr = &stderr
    cmd.Stdout = &out

    if err := cmd.Run(); err != nil {
        fmt.Println("Error:", stderr.String())
        fmt.Println("Failed to execute command:", err)
        return
    }
    fmt.Println("Go list:\n", out.String())
    ```
