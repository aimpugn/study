# security

## `find-generic-password`

> Find a generic password item.

```bash
find-generic-password [-a account] [-s service] [options...] [-g] [keychain...]
```

- `-a`: Match "account" string
- `-c`: Match "creator" (four-character code)
- `-C`: Match "type" (four-character code)
- `-D`: Match "kind" string
- `-G`: Match "value" string (generic attribute)
- `-j`: Match "comment" string
- `-l`: Match "label" string
- `-s`: Match "service" string
- `-g`: Display the password for the item found
- `-w`: Display only the password on stdout

If no keychains are specified to search, the default search list is used.

## `add-generic-password`

> Add a generic password item.

```bash
add-generic-password [-a account] [-s service] [-w password] [options...] [-A|-T appPath] [keychain]
```

- `-a`: Specify account name (required)
- `-c`: Specify item creator (optional four-character code)
- `-C`: Specify item type (optional four-character code)
- `-D`: Specify kind (default is "application password")
- `-G`: Specify generic attribute (optional)
- `-j`: Specify comment string (optional)
- `-l`: Specify label (if omitted, service name is used as default label)
- `-s`: Specify service name (required)
- `-p`: Specify password to be added (legacy option, equivalent to -w)
- `-w`: Specify password to be added
- `-X`: Specify password data to be added as a hexadecimal string
- `-A`: Allow any application to access this item without warning (insecure, not recommended!)
- `-T`: Specify an application which may access this item (multiple -T options are allowed)
- `-U`: Update item if it already exists (if omitted, the item cannot already exist)

By default, the application which creates an item is trusted to access its data without warning.
You can remove this default access by explicitly specifying an empty app pathname: -T ""

If no keychain is specified, the password is added to the default keychain.

Use of the -p or -w options is insecure. Specify -w as the last option to be prompted.
