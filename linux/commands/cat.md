# cat

## Append to file

- https://stackoverflow.com/questions/17701989/how-do-i-append-text-to-a-file
- [Here Documents](https://tldp.org/LDP/abs/html/here-docs.html)

```shell
cat << EOF >> filename
This is text entered via the keyboard or via a script.
EOF
```

```shell
sudo sh -c 'cat << EOF >> filename
This is text entered via the keyboard.
EOF'
```

```shell
# to avoid cli sudo issue
tee -a filename << EOF
This is text entered via the keyboard or via a script.
EOF
```

