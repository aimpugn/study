# CLI

## version change

### mac os

- [Switching between PHP versions when using Homebrew](https://localheinz.com/articles/2020/05/05/switching-between-php-versions-when-using-homebrew/)

```shell
brew list

brew unlink php@7.4

brew link php@8.1 --force --overwrite
```

## cgi-fcgi

```shell
sudo -u www-data \
  SCRIPT_NAME=/php-fpm-status \
  SCRIPT_FILENAME=/php-fpm-status \
  QUERY_STRING='json&full' \
  REQUEST_METHOD=GET \
  cgi-fcgi -bind -connect /run/php/php5.6-fpm.sock
```
