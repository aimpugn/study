# detekt

## [cli로 사용하기](https://detekt.dev/docs/gettingstarted/cli)

### install

```shell
brew install detekt
```

### options

- `--all-rules`

    ```text
    Activates all available (even unstable) rules.
    Default: false
    ```

- `--auto-correct`, `-ac`

    ```text
    Allow rules to auto correct code if they support it. The default rule
    sets do NOT support auto correcting and won't change any line in the
    users code base. However custom rules can be written to support auto
    correcting. The additional 'formatting' rule set, added with
    '--plugins', does support it and needs this flag.
    Default: false
    ```

- `--base-path`, `-bp`

    ```text
    Specifies a directory as the base path.Currently it impacts all file
    paths in the formatted reports. File paths in console output and txt
    report are not affected and remain as absolute paths.
    ```

- `--baseline`, `-b`

    ```text
    If a baseline xml file is passed in, only new code smells not in the
    baseline are printed in the console.
    ```

- `--build-upon-default-config`

    ```text
    Preconfigures detekt with a bunch of rules and some opinionated defaults
    for you. Allows additional provided configurations to override the
    defaults.
    Default: false
    ```

- `--classpath`, `-cp`

    ```text
    EXPERIMENTAL: Paths where to find user class files and depending jar
    files. Used for type resolution.
    ```

- `--config`, `-c`

    ```text
    Path to the config file (path/to/config.yml). Multiple configuration
    files can be specified with ',' or ';' as separator.
    ```

- `--config-resource`, `-cr`

    ```text
    Path to the config resource on detekt's classpath (path/to/config.yml).
    ```

- `--create-baseline`, `-cb`

    ```text
    Treats current analysis findings as a smell baseline for future detekt
    runs.
    Default: false
    ```

- `--debug`

    ```text
    Prints extra information about configurations and extensions.
    Default: false
    ```

- `--disable-default-rulesets`, `-dd`

    ```text
    Disables default rule sets.
    Default: false
    ```

- `--excludes`, `-ex`

    ```text
    Globbing patterns describing paths to exclude from the analysis.
    ```

- `--fail-fast`

    ```text
    DEPRECATED: please use '--build-upon-default-config' together with
    '--all-rules'. Same as 'build-upon-default-config' but explicitly
    running all available rules. With this setting only exit code 0 is
    returned when the analysis does not find a single code smell. Additional
    configuration files can override rule properties which includes turning
    off specific rules.
    Default: false
    ```

- `--generate-config`, `-gc`

    ```text
    Export default config. Path can be specified with --config option
    (default path: default-detekt-config.yml)
    Default: false
    ```

- `--help`, `-h`

    ```text
    Shows the usage.
    ```

- `--includes`, `-in`

    ```text
    Globbing patterns describing paths to include in the analysis. Useful in combination with 'excludes' patterns.
    ```

- `--input`, `-i`

    ```text
    Input paths to analyze. Multiple paths are separated by comma. If not
    specified the current working directory is used.
    ```

- `--jvm-target`

    ```text
    EXPERIMENTAL: Target version of the generated JVM bytecode that was
    generated during compilation and is now being used for type resolution
    (1.6, 1.8, 9, 10, 11, 12, 13, 14, 15, 16 or 17)
    Default: 1.8
    ```

- `--language-version`

    ```text
    EXPERIMENTAL: Compatibility mode for Kotlin language version X.Y,
    reports errors for all language features that came out later
    Possible Values: [1.0, 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9]
    ```

- `--max-issues`

    ```text
    Return exit code 0 only when found issues count does not exceed
    specified issues count.
    ```

- `--parallel`

    ```text
    Enables parallel compilation and analysis of source files. Do some
    benchmarks first before enabling this flag. Heuristics show performance
    benefits starting from 2000 lines of Kotlin code.
    Default: false
    ```

- `--plugins`, `-p`

    ```text
    Extra paths to plugin jars separated by ',' or ';'.
    ```

- `--report`, `-r`

    ```text
    Generates a report for given 'report-id' and stores it on given 'path'.
    Entry should consist of: [report-id:path]. Available 'report-id' values:
    'txt', 'xml', 'html', 'md', 'sarif'. These can also be used in
    combination with each other e.g. '-r txt:reports/detekt.txt -r
    xml:reports/detekt.xml'
    ```

- `--version`

    ```text
    Prints the detekt CLI version.
    Default: false
    ```
  
### run via cli

```shell
mkdir -p ~/Documents/reports/detekt/
detekt --excludes "**/interface/**" \
  -r html:/Users/rody/Documents/reports/detekt/detekt.221106191100.html
```
