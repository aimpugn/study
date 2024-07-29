# nl

- [nl](#nl)
    - [nl?](#nl-1)
    - [DESCRIPTION](#description)
    - [EXAMPLES](#examples)
        - [Number all non-blank lines](#number-all-non-blank-lines)
        - [Number all lines including blank ones, with right justified line numbers with leading zeroes, starting at 2, with increment of 2 and a custom multi-character separator](#number-all-lines-including-blank-ones-with-right-justified-line-numbers-with-leading-zeroes-starting-at-2-with-increment-of-2-and-a-custom-multi-character-separator)
        - [Number lines matching regular expression for an i followed by either m or n](#number-lines-matching-regular-expression-for-an-i-followed-by-either-m-or-n)

## nl?

- line numbering filter

## DESCRIPTION

The `nl` utility reads lines from the named file, applies a configurable line numbering filter operation, and writes the result to the standard output.  
If file is a single dash(`-`) or absent, `nl` reads from the standard input.

The `nl` utility treats the text it reads in terms of logical pages.  
Unless specified otherwise, line numbering is reset at the start of each logical page.  
A logical page consists of a header, a body and a footer section; empty sections are valid.  
Different line numbering options are independently available for header, body and footer sections.

The starts of logical page sections are signalled by input lines containing nothing but one of the following sequences of delimiter characters:

```text
Line      Start of
\:\:\:    header
\:\:      body
\:        footer
```

If the input does not contain any logical page section signalling directives, the text being read is assumed to consist of a single logical page body.

## EXAMPLES

### Number all non-blank lines

```bash
$ echo -e "This is\n\n\na simple text" | nl
    1  This is


    2  a simple text
```

### Number all lines including blank ones, with right justified line numbers with leading zeroes, starting at 2, with increment of 2 and a custom multi-character separator

```bash
$ echo -e "This\nis\nan\n\n\nexample" | nl -ba -n rz -i2 -s "->" -v2
000002->This
000004->is
000006->an
000008->
000010->
000012->example
```

### Number lines matching regular expression for an i followed by either m or n

```bash
$ echo -e "This is\na simple text\nwith multiple\nlines" | nl -bp'i[mn]'
        This is
    1  a simple text
        with multiple
    2  lines
```
