# zend_parse_parameters

## specifier

- `a`: array (`zval*`)
- `A`: array or object (`zval*`)
- `b`: boolean (`zend_bool`)
- `C`: class (`zend_class_entry*`)
- `d`: double (double)
- `f`: function or array containing php method call info (returned as `zend_fcall_info` and `zend_fcall_info_cache`)
- `h`: array (returned as `HashTable*`)
- `H`: array or `HASH_OF(object)` (returned as `HashTable*`)
- `l`: long (long)
- `L`: long, limits out-of-range numbers to LONG_MAX/LONG_MIN (long)
- `o`: object of any type (`zval*`)
- `O`: object of specific type given by class entry (`zval*`, zend_class_entry)
- `p`: valid path (string without null bytes in the middle) and its length (char*, int)
- `r`: resource (`zval*`)
- `s`: string (with possible null bytes) and its length (char*, int)
- `z`: the actual zval (`zval*`)
- `Z`: the actual zval (`zval**`)
- `*`: variable arguments list (0 or more)
- `+`: variable arguments list (1 or more)

The following characters also have a meaning in the specifier string:
- `|`: indicates that the remaining parameters are optional, they should be initialized to default values by the extension since they will not be touched by the parsing function if they are not passed to it.
- `/`: use `SEPARATE_ZVAL_IF_NOT_REF()` on the parameter it follows
- `!`:
    - the parameter it follows can be of specified type or `NULL`.
    - If NULL is passed and the output for such type is a pointer, then the output pointer is set to a native `NULL` pointer.
    - For `b`, `l` and `d`, an extra argument of type `zend_bool*` must be passed after the corresponding `bool*`, `long*`or `double*` arguments, respectively.
    - A non-zero value will be written to the `zend_bool` iif a PHP `NULL` is passed.
