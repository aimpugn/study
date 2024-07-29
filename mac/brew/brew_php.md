# brew php

- [brew php](#brew-php)
    - [install](#install)
    - [php8.2](#php82)
    - [기타](#기타)

## install

```shell
brew tap shivammathur/php
brew install shivammathur/php/php@7.3

brew link --overwrite --force shivammathur/php/php@7.3
```

## php8.2

```bash
==> Caveats
To enable PHP in Apache add the following to httpd.conf and restart Apache:
    LoadModule php_module /opt/homebrew/opt/php@8.2/lib/httpd/modules/libphp.so

    <FilesMatch \.php$>
        SetHandler application/x-httpd-php
    </FilesMatch>

Finally, check DirectoryIndex includes index.php
    DirectoryIndex index.php index.html

The php.ini and php-fpm.ini file can be found in:
    /opt/homebrew/etc/php/8.2/

php@8.2 is keg-only, which means it was not symlinked into /opt/homebrew,
because this is an alternate version of another formula.

If you need to have php@8.2 first in your PATH, run:
  echo 'export PATH="/opt/homebrew/opt/php@8.2/bin:$PATH"' >> ~/.zshrc
  echo 'export PATH="/opt/homebrew/opt/php@8.2/sbin:$PATH"' >> ~/.zshrc

For compilers to find php@8.2 you may need to set:
  export LDFLAGS="-L/opt/homebrew/opt/php@8.2/lib"
  export CPPFLAGS="-I/opt/homebrew/opt/php@8.2/include"

To start php@8.2 now and restart at login:
  brew services start php@8.2
Or, if you don't want/need a background service you can just run:
  /opt/homebrew/opt/php@8.2/sbin/php-fpm --nodaemonize
```

## 기타

- [macOS 13.0 Ventura Apache Setup: Multiple PHP Versions](https://getgrav.org/blog/macos-ventura-apache-multiple-php-versions)
- [shivammathur/homebrew-php](https://github.com/shivammathur/homebrew-php)
