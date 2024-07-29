     __________________________________________________________________

Installing PHP
     __________________________________________________________________

     * General Installation Considerations
     * Installation on Unix systems
          + Apache 1.3.x on Unix systems
          + Apache 2.x on Unix systems
          + Lighttpd 1.4 on Unix systems
          + Sun, iPlanet and Netscape servers on Sun Solaris
          + CGI and command line setups
          + HP-UX specific installation notes
          + OpenBSD installation notes
          + Solaris specific installation tips
          + Debian GNU/Linux installation notes
     * Installation on Mac OS X
          + Using Packages
          + Using the bundled PHP
          + Compiling PHP on Mac OS X
     * Installation of PECL extensions
          + Introduction to PECL Installations
          + Downloading PECL extensions
          + Installing a PHP extension on Windows
          + Compiling shared PECL extensions with the pecl command
          + Compiling shared PECL extensions with phpize
          + php-config
          + Compiling PECL extensions statically into PHP
     * Problems?
          + Read the FAQ
          + Other problems
          + Bug reports
     * Runtime Configuration
          + The configuration file
          + .user.ini files
          + Where a configuration setting may be set
          + How to change configuration settings
     * Installation
     __________________________________________________________________

     __________________________________________________________________

Preface

   These installation instructions were generated from the HTML version of
   the  PHP  Manual  so  formatting and linking have been altered. See the
   online and updated version at: http://php.net/install.unix
     __________________________________________________________________

General Installation Considerations

   Before  starting  the  installation, first you need to know what do you
   want  to  use  PHP for. There are three main fields you can use PHP, as
   described in the What can PHP do? section:
     * Websites and web applications (server-side scripting)
     * Command line scripting
     * Desktop (GUI) applications

   For  the first and most common form, you need three things: PHP itself,
   a  web  server  and  a  web  browser.  You  probably already have a web
   browser,  and  depending  on  your operating system setup, you may also
   have  a  web server (e.g. Apache on Linux and MacOS X; IIS on Windows).
   You  may  also  rent webspace at a company. This way, you don't need to
   set  up anything on your own, only write your PHP scripts, upload it to
   the server you rent, and see the results in your browser.

   In  case  of  setting  up  the server and PHP on your own, you have two
   choices  for  the  method  of  connecting  PHP  to the server. For many
   servers  PHP  has  a  direct module interface (also called SAPI). These
   servers include Apache, Microsoft Internet Information Server, Netscape
   and  iPlanet  servers.  Many  other servers have support for ISAPI, the
   Microsoft  module  interface  (OmniHTTPd  for  example).  If PHP has no
   module  support  for your web server, you can always use it as a CGI or
   FastCGI  processor.  This  means  you set up your server to use the CGI
   executable of PHP to process all PHP file requests on the server.

   If  you are also interested to use PHP for command line scripting (e.g.
   write scripts autogenerating some images for you offline, or processing
   text  files  depending  on some arguments you pass to them), you always
   need  the  command  line  executable.  For  more  information, read the
   section  about writing command line PHP applications. In this case, you
   need no server and no browser.

   With  PHP you can also write desktop GUI applications using the PHP-GTK
   extension.  This  is  a  completely different approach than writing web
   pages,  as  you  do not output any HTML, but manage windows and objects
   within  them.  For  more  information about PHP-GTK, please » visit the
   site  dedicated  to  this  extension.  PHP-GTK  is  not included in the
   official PHP distribution.

   From  now on, this section deals with setting up PHP for web servers on
   Unix and Windows with server module interfaces and CGI executables. You
   will  also  find  information  on  the  command  line executable in the
   following sections.

   PHP  source  code  and binary distributions for Windows can be found at
   » http://www.php.net/downloads.php.   We  recommend  you  to  choose  a
   » mirror nearest to you for downloading the distributions.
     __________________________________________________________________
     __________________________________________________________________

Installation on Unix systems

Table of Contents

     * Apache 1.3.x on Unix systems
     * Apache 2.x on Unix systems
     * Lighttpd 1.4 on Unix systems
     * Sun, iPlanet and Netscape servers on Sun Solaris
     * CGI and command line setups
     * HP-UX specific installation notes
     * OpenBSD installation notes
     * Solaris specific installation tips
     * Debian GNU/Linux installation notes

   This  section  will  guide  you  through  the general configuration and
   installation  of  PHP  on  Unix  systems.  Be  sure  to investigate any
   sections  specific  to your platform or web server before you begin the
   process.

   As  our  manual  outlines  in  the  General Installation Considerations
   section,  we  are mainly dealing with web centric setups of PHP in this
   section,  although  we will cover setting up PHP for command line usage
   as well.

   There  are  several  ways  to install PHP for the Unix platform, either
   with  a  compile and configure process, or through various pre-packaged
   methods.  This  documentation  is  mainly focused around the process of
   compiling and configuring PHP. Many Unix like systems have some sort of
   package  installation  system. This can assist in setting up a standard
   configuration,  but  if  you  need  to have a different set of features
   (such as a secure server, or a different database driver), you may need
   to  build  PHP  and/or  your  web  server.  If  you are unfamiliar with
   building  and  compiling your own software, it is worth checking to see
   whether  somebody  has already built a packaged version of PHP with the
   features you need.

   Prerequisite knowledge and software for compiling:
     * Basic Unix skills (being able to operate "make" and a C compiler)
     * An ANSI C compiler
     * A web server
     * Any module specific components (such as GD, PDF libs, etc.)

   When  building  directly from Git sources or after custom modifications
   you might also need:
     * autoconf: 2.13+ (for PHP < 5.4.0), 2.59+ (for PHP >= 5.4.0)
     * automake: 1.4+
     * libtool: 1.4.x+ (except 1.4.2)
     * re2c: Version 0.13.4 or newer
     * flex: Version 2.5.4 (for PHP <= 5.2)
     * bison: Version 1.28 (preferred), 1.35, or 1.75

   The  initial  PHP  setup and configuration process is controlled by the
   use  of the command line options of the configure script. You could get
   a  list  of all available options along with short explanations running
   ./configure   --help.   Our  manual  documents  the  different  options
   separately.  You  will find the core options in the appendix, while the
   different  extension  specific  options  are  described on the reference
   pages.

   When  PHP  is  configured,  you  are  ready  to build the module and/or
   executables. The command make should take care of this. If it fails and
   you can't figure out why, see the Problems section.
     __________________________________________________________________

Apache 1.3.x on Unix systems

   This  section  contains  notes and hints specific to Apache installs of
   PHP on Unix platforms. We also have instructions and notes for Apache 2
   on a separate page.

   You  can select arguments to add to the configure on line 10 below from
   the  list of core configure options and from extension specific options
   described  at  the respective places in the manual. The version numbers
   have  been  omitted here, to ensure the instructions are not incorrect.
   You  will  need  to replace the 'xxx' here with the correct values from
   your files.

   Example #1 Installation Instructions (Apache Shared Module Version) for
   PHP
1.  gunzip apache_xxx.tar.gz
2.  tar -xvf apache_xxx.tar
3.  gunzip php-xxx.tar.gz
4.  tar -xvf php-xxx.tar
5.  cd apache_xxx
6.  ./configure --prefix=/www --enable-module=so
7.  make
8.  make install
9.  cd ../php-xxx

10. Now, configure your PHP.  This is where you customize your PHP
    with various options, like which extensions will be enabled.  Do a
    ./configure --help for a list of available options.  In our example
    we'll do a simple configure with Apache 1 and MySQL support.  Your
    path to apxs may differ from our example.

      ./configure --with-mysql --with-apxs=/www/bin/apxs

11. make
12. make install

    If you decide to change your configure options after installation,
    you only need to repeat the last three steps. You only need to
    restart apache for the new module to take effect. A recompile of
    Apache is not needed.

    Note that unless told otherwise, 'make install' will also install PEAR,
    various PHP tools such as phpize, install the PHP CLI, and more.

13. Setup your php.ini file:

      cp php.ini-development /usr/local/lib/php.ini

    You may edit your .ini file to set PHP options.  If you prefer your
    php.ini in another location, use --with-config-file-path=/some/path in
    step 10.

    If you instead choose php.ini-production, be certain to read the list
    of changes within, as they affect how PHP behaves.

14. Edit your httpd.conf to load the PHP module.  The path on the right hand
    side of the LoadModule statement must point to the path of the PHP
    module on your system.  The make install from above may have already
    added this for you, but be sure to check.

      LoadModule php5_module libexec/libphp5.so

15. And in the AddModule section of httpd.conf, somewhere under the
    ClearModuleList, add this:

      AddModule mod_php5.c

16. Tell Apache to parse certain extensions as PHP.  For example,
    let's have Apache parse the .php extension as PHP.  You could
    have any extension(s) parse as PHP by simply adding more, with
    each separated by a space.  We'll add .phtml to demonstrate.

      AddType application/x-httpd-php .php .phtml

    It's also common to setup the .phps extension to show highlighted PHP
    source, this can be done with:

      AddType application/x-httpd-php-source .phps

17. Use your normal procedure for starting the Apache server. (You must
    stop and restart the server, not just cause the server to reload by
    using a HUP or USR1 signal.)

   Alternatively, to install PHP as a static object:

   Example  #2  Installation  Instructions (Static Module Installation for
   Apache) for PHP
1.  gunzip -c apache_1.3.x.tar.gz | tar xf -
2.  cd apache_1.3.x
3.  ./configure
4.  cd ..

5.  gunzip -c php-5.x.y.tar.gz | tar xf -
6.  cd php-5.x.y
7.  ./configure --with-mysql --with-apache=../apache_1.3.x
8.  make
9.  make install

10. cd ../apache_1.3.x

11. ./configure --prefix=/www --activate-module=src/modules/php5/libphp5.a
    (The above line is correct! Yes, we know libphp5.a does not exist at this
    stage. It isn't supposed to. It will be created.)

12. make
    (you should now have an httpd binary which you can copy to your Apache bin d
ir if
    it is your first install then you need to "make install" as well)

13. cd ../php-5.x.y
14. cp php.ini-development /usr/local/lib/php.ini

15. You can edit /usr/local/lib/php.ini file to set PHP options.
    Edit your httpd.conf or srm.conf file and add:
    AddType application/x-httpd-php .php

   Depending  on  your  Apache  install  and  Unix variant, there are many
   possible  ways  to  stop and restart the server. Below are some typical
   lines   used  in  restarting  the  server,  for  different  apache/unix
   installations.  You  should  replace  /path/to/  with the path to these
   applications on your systems.

   Example #3 Example commands for restarting Apache
1. Several Linux and SysV variants:
/etc/rc.d/init.d/httpd restart

2. Using apachectl scripts:
/path/to/apachectl stop
/path/to/apachectl start

3. httpdctl and httpsdctl (Using OpenSSL), similar to apachectl:
/path/to/httpsdctl stop
/path/to/httpsdctl start

4. Using mod_ssl, or another SSL server, you may want to manually
stop and start:
/path/to/apachectl stop
/path/to/apachectl startssl

   The  locations of the apachectl and http(s)dctl binaries often vary. If
   your  system  has locate or whereis or which commands, these can assist
   you in finding your server control programs.

   Different examples of compiling PHP for apache are as follows:
./configure --with-apxs --with-pgsql

   This will create a libphp5.so shared library that is loaded into Apache
   using  a  LoadModule  line  in Apache's httpd.conf file. The PostgreSQL
   support is embedded into this library.

./configure --with-apxs --with-pgsql=shared

   This  will  create  a libphp5.so shared library for Apache, but it will
   also create a pgsql.so shared library that is loaded into PHP either by
   using  the  extension  directive  in  php.ini  file  or  by  loading it
   explicitly in a script using the dl() function.

./configure --with-apache=/path/to/apache_source --with-pgsql

   This  will  create  a  libmodphp5.a  library,  a  mod_php5.c  and  some
   accompanying files and copy this into the src/modules/php5 directory in
   the    Apache    source   tree.   Then   you   compile   Apache   using
   --activate-module=src/modules/php5/libphp5.a   and   the  Apache  build
   system  will  create  libphp5.a  and  link it statically into the httpd
   binary.  The  PostgreSQL  support  is included directly into this httpd
   binary, so the final result here is a single httpd binary that includes
   all of Apache and all of PHP.

./configure --with-apache=/path/to/apache_source --with-pgsql=shared

   Same as before, except instead of including PostgreSQL support directly
   into  the  final  httpd you will get a pgsql.so shared library that you
   can load into PHP from either the php.ini file or directly using dl().

   When  choosing  to build PHP in different ways, you should consider the
   advantages  and  drawbacks  of each method. Building as a shared object
   will  mean  that  you  can compile apache separately, and don't have to
   recompile  everything  as you add to, or change, PHP. Building PHP into
   apache  (static  method)  means  that PHP will load and run faster. For
   more information, see the Apache » web page on DSO support.

     Note:

     Apache's  default  httpd.conf  currently  ships  with a section that
     looks like this:

User nobody
Group "#-1"

     Unless  you  change  that  to "Group nogroup" or something like that
     ("Group  daemon"  is  also very common) PHP will not be able to open
     files.

     Note:

     Make  sure  you  specify  the  installed  version of apxs when using
     --with-apxs=/path/to/apxs  .  You must NOT use the apxs version that
     is  in  the apache sources but the one that is actually installed on
     your system.
     __________________________________________________________________
     __________________________________________________________________

Apache 2.x on Unix systems

   This  section  contains notes and hints specific to Apache 2.x installs
   of PHP on Unix systems.
   Warning

   We  do  not recommend using a threaded MPM in production with Apache 2.
   Use  the prefork MPM, which is the default MPM with Apache 2.0 and 2.2.
   For  information  on  why,  read the related FAQ entry on using Apache2
   with a threaded MPM

   The   » Apache  Documentation  is  the  most  authoritative  source  of
   information   on   the   Apache  2.x  server.  More  information  about
   installation options for Apache may be found there.

   The  most  recent  version  of  Apache HTTP Server may be obtained from
   » Apache  download  site,  and  a  fitting  PHP  version from the above
   mentioned  places.  This  quick  guide  covers  only  the basics to get
   started with Apache 2.x and PHP. For more information read the » Apache
   Documentation.  The  version  numbers have been omitted here, to ensure
   the  instructions are not incorrect. In the examples below, 'NN' should
   be replaced with the specific version of Apache being used.

   There  are  currently two versions of Apache 2.x - there's 2.0 and 2.2.
   While  there  are various reasons for choosing each, 2.2 is the current
   latest  version,  and  the  one  that is recommended, if that option is
   available  to  you. However, the instructions here will work for either
   2.0 or 2.2.
    1. Obtain  the  Apache HTTP server from the location listed above, and
       unpack it:
gzip -d httpd-2_x_NN.tar.gz
tar -xf httpd-2_x_NN.tar

    2. Likewise, obtain and unpack the PHP source:
gunzip php-NN.tar.gz
tar -xf php-NN.tar

    3. Build  and install Apache. Consult the Apache install documentation
       for more details on building Apache.
cd httpd-2_x_NN
./configure --enable-so
make
make install

    4. Now  you  have  Apache  2.x.NN  available under /usr/local/apache2,
       configured  with  loadable  module  support  and  the  standard MPM
       prefork.  To  test  the  installation use your normal procedure for
       starting the Apache server, e.g.:
/usr/local/apache2/bin/apachectl start

       and stop the server to go on with the configuration for PHP:
/usr/local/apache2/bin/apachectl stop

    5. Now,  configure and build PHP. This is where you customize PHP with
       various  options,  like  which  extensions  will  be  enabled.  Run
       ./configure  --help for a list of available options. In our example
       we'll do a simple configure with Apache 2 and MySQL support.
       If  you  built  Apache  from  source, as described above, the below
       example  will match your path for apxs, but if you installed Apache
       some other way, you'll need to adjust the path to apxs accordingly.
       Note that some distros may rename apxs to apxs2.
cd ../php-NN
./configure --with-apxs2=/usr/local/apache2/bin/apxs --with-mysql
make
make install

       If  you decide to change your configure options after installation,
       you'll  need to re-run the configure, make, and make install steps.
       You  only need to restart apache for the new module to take effect.
       A recompile of Apache is not needed.
       Note  that  unless told otherwise, 'make install' will also install
       PEAR,  various  PHP  tools such as phpize, install the PHP CLI, and
       more.
    6. Setup your php.ini
cp php.ini-development /usr/local/lib/php.ini

       You  may  edit  your  .ini  file  to set PHP options. If you prefer
       having php.ini in another location, use
       --with-config-file-path=/some/path in step 5.
       If  you  instead  choose php.ini-production, be certain to read the
       list of changes within, as they affect how PHP behaves.
    7. Edit  your httpd.conf to load the PHP module. The path on the right
       hand side of the LoadModule statement must point to the path of the
       PHP  module  on  your  system. The make install from above may have
       already added this for you, but be sure to check.
LoadModule php5_module modules/libphp5.so
    8. Tell  Apache to parse certain extensions as PHP. For example, let's
       have  Apache  parse  .php  files  as PHP. Instead of only using the
       Apache  AddType  directive,  we want to avoid potentially dangerous
       uploads  and  created  files  such  as  exploit.php.jpg  from being
       executed   as   PHP.   Using  this  example,  you  could  have  any
       extension(s)  parse as PHP by simply adding them. We'll add .php to
       demonstrate.
<FilesMatch \.php$>
    SetHandler application/x-httpd-php
</FilesMatch>
       Or,  if we wanted to allow .php, .php2, .php3, .php4, .php5, .php6,
       and  .phtml files to be executed as PHP, but nothing else, we'd use
       this:
<FilesMatch "\.ph(p[2-6]?|tml)$">
    SetHandler application/x-httpd-php
</FilesMatch>
       And  to  allow  .phps files to be handled by the php source filter,
       and displayed as syntax-highlighted source code, use this:
<FilesMatch "\.phps$">
    SetHandler application/x-httpd-php-source
</FilesMatch>
       mod_rewrite  may  be  used  To  allow any arbitrary .php file to be
       displayed  as  syntax-highlighted  source  code,  without having to
       rename or copy it to a .phps file:
RewriteEngine On
RewriteRule (.*\.php)s$ $1 [H=application/x-httpd-php-source]
       The  php source filter should not be enabled on production systems,
       where it may expose confidential or otherwise sensitive information
       embedded in source code.
    9. Use your normal procedure for starting the Apache server, e.g.:
/usr/local/apache2/bin/apachectl start

       OR
service httpd restart

   Following  the  steps  above you will have a running Apache2 web server
   with  support  for  PHP as a SAPI module. Of course there are many more
   configuration  options  available  Apache and PHP. For more information
   type ./configure --help in the corresponding source tree.

   Apache  may  be built multithreaded by selecting the worker MPM, rather
   than  the  standard  prefork MPM, when Apache is built. This is done by
   adding  the  following option to the argument passed to ./configure, in
   step 3 above:
   --with-mpm=worker

   This  should  not be undertaken without being aware of the consequences
   of  this  decision,  and  having  at  least a fair understanding of the
   implications.   The   Apache   documentation   regarding  » MPM-Modules
   discusses MPMs in a great deal more detail.

     Note:

     The Apache MultiViews FAQ discusses using multiviews with PHP.

     Note:

     To  build  a multithreaded version of Apache, the target system must
     support  threads.  In  this  case,  PHP  should  also  be built with
     experimental Zend Thread Safety (ZTS). Under this configuration, not
     all  extensions will be available. The recommended setup is to build
     Apache with the default prefork MPM-Module.
     __________________________________________________________________
     __________________________________________________________________

Lighttpd 1.4 on Unix systems

   This section contains notes and hints specific to Lighttpd 1.4 installs
   of PHP on Unix systems.

   Please  use  the  » Lighttpd  trac  to  learn  how  to install Lighttpd
   properly before continuing.

   Fastcgi  is  the preferred SAPI to connect PHP and Lighttpd. Fastcgi is
   automagically  enabled  in  php-cgi  in PHP 5.3, but for older versions
   configure  PHP  with  --enable-fastcgi. To confirm that PHP has fastcgi
   enabled,  php  -v should contain PHP 5.2.5 (cgi-fcgi) Before PHP 5.2.3,
   fastcgi was enabled on the php binary (there was no php-cgi).

Letting Lighttpd spawn php processes

   To  configure  Lighttpd  to connect to php and spawn fastcgi processes,
   edit  lighttpd.conf.  Sockets  are  preferred  to  connect  to  fastcgi
   processes on the local system.

   Example #1 Partial lighttpd.conf
server.modules += ( "mod_fastcgi" )

fastcgi.server = ( ".php" =>
  ((
    "socket" => "/tmp/php.socket",
    "bin-path" => "/usr/local/bin/php-cgi",
    "bin-environment" => (
      "PHP_FCGI_CHILDREN" => "16",
      "PHP_FCGI_MAX_REQUESTS" => "10000"
    ),
    "min-procs" => 1,
    "max-procs" => 1,
    "idle-timeout" => 20
  ))
)

   The  bin-path  directive  allows  lighttpd  to  spawn fastcgi processes
   dynamically. PHP will spawn children according to the PHP_FCGI_CHILDREN
   environment   variable.   The   "bin-environment"  directive  sets  the
   environment  for  the  spawned processes. PHP will kill a child process
   after  the  number  of  requests  specified by PHP_FCGI_MAX_REQUESTS is
   reached. The directives "min-procs" and "max-procs" should generally be
   avoided  with  PHP. PHP manages its own children and opcode caches like
   APC  will  only  share among children managed by PHP. If "min-procs" is
   set  to  something  greater  than 1, the total number of php responders
   will  be  multiplied PHP_FCGI_CHILDREN (2 min-procs * 16 children gives
   32 responders).

Spawning with spawn-fcgi

   Lighttpd  provides  a  program called spawn-fcgi to ease the process of
   spawning fastcgi processes easier.

Spawning php-cgi

   It  is  possible to spawn processes without spawn-fcgi, though a bit of
   heavy-lifting  is  required.  Setting the PHP_FCGI_CHILDREN environment
   var  controls  how  many  children  PHP  will  spawn to handle incoming
   requests.  Setting  PHP_FCGI_MAX_REQUESTS  will  determine how long (in
   requests)  each  child  will  live. Here's a simple bash script to help
   spawn php responders.

   Example #2 Spawning FastCGI Responders
#!/bin/sh

# Location of the php-cgi binary
PHP=/usr/local/bin/php-cgi

# PID File location
PHP_PID=/tmp/php.pid

# Binding to an address
#FCGI_BIND_ADDRESS=10.0.1.1:10000
# Binding to a domain socket
FCGI_BIND_ADDRESS=/tmp/php.sock

PHP_FCGI_CHILDREN=16
PHP_FCGI_MAX_REQUESTS=10000

env -i PHP_FCGI_CHILDREN=$PHP_FCGI_CHILDREN \
       PHP_FCGI_MAX_REQUESTS=$PHP_FCGI_MAX_REQUESTS \
       $PHP -b $FCGI_BIND_ADDRESS &

echo $! > "$PHP_PID"


Connecting to remote FCGI instances

   Fastcgi  instances  can be spawned on multiple remote machines in order
   to scale applications.

   Example #3 Connecting to remote php-fastcgi instances
fastcgi.server = ( ".php" =>
   (( "host" => "10.0.0.2", "port" => 1030 ),
    ( "host" => "10.0.0.3", "port" => 1030 ))
)
     __________________________________________________________________
     __________________________________________________________________

Sun, iPlanet and Netscape servers on Sun Solaris

   This  section  contains notes and hints specific to Sun Java System Web
   Server, Sun ONE Web Server, iPlanet and Netscape server installs of PHP
   on Sun Solaris.

   From  PHP  4.3.3  on  you  can use PHP scripts with the NSAPI module to
   generate   custom   directory  listings  and  error  pages.  Additional
   functions  for  Apache compatibility are also available. For support in
   current web servers read the note about subrequests.

   You  can  find  more  information about setting up PHP for the Netscape
   Enterprise Server (NES) here:
   » http://benoit.noss.free.fr/php/install-php4.html

   To  build  PHP  with  Sun JSWS/Sun ONE WS/iPlanet/Netscape web servers,
   enter  the  proper install directory for the --with-nsapi=[DIR] option.
   The  default directory is usually /opt/netscape/suitespot/. Please also
   read /php-xxx-version/sapi/nsapi/nsapi-readme.txt.

    1. Install  the following packages from »  http://www.sunfreeware.com/
       or another download site:
          + autoconf-2.13
          + automake-1.4
          + bison-1_25-sol26-sparc-local
          + flex-2_5_4a-sol26-sparc-local
          + gcc-2_95_2-sol26-sparc-local
          + gzip-1.2.4-sol26-sparc-local
          + m4-1_4-sol26-sparc-local
          + make-3_76_1-sol26-sparc-local
          + mysql-3.23.24-beta (if you want mysql support)
          + perl-5_005_03-sol26-sparc-local
          + tar-1.13 (GNU tar)
    2. Make    sure    your   path   includes   the   proper   directories
       PATH=.:/usr/local/bin:/usr/sbin:/usr/bin:/usr/ccs/bin  and  make it
       available to your system export PATH.
    3. gunzip  php-x.x.x.tar.gz  (if  you have a .gz dist, otherwise go to
       4).
    4. tar xvf php-x.x.x.tar
    5. Change to your extracted PHP directory: cd ../php-x.x.x
    6. For the following step, make sure /opt/netscape/suitespot/ is where
       your netscape server is installed. Otherwise, change to the correct
       path and run:
./configure --with-mysql=/usr/local/mysql \
--with-nsapi=/opt/netscape/suitespot/ \
--enable-libgcc
    7. Run make followed by make install.

   After  performing  the  base install and reading the appropriate readme
   file, you may need to perform some additional configuration steps.

Configuration Instructions for Sun/iPlanet/Netscape

   Firstly  you  may  need  to  add  some  paths  to  the  LD_LIBRARY_PATH
   environment  for  the server to find all the shared libs. This can best
   done in the start script for your web server. The start script is often
   located  in:  /path/to/server/https-servername/start. You may also need
   to    edit    the    configuration   files   that   are   located   in:
   /path/to/server/https-servername/config/.
    1. Add  the  following  line  to  mime.types  (you  can do that by the
       administration server):
type=magnus-internal/x-httpd-php exts=php

    2. Edit  magnus.conf  (for servers >= 6) or obj.conf (for servers < 6)
       and add the following, shlib will vary depending on your system, it
       will  be something like /opt/netscape/suitespot/bin/libphp4.so. You
       should place the following lines after mime types init.
Init fn="load-modules" funcs="php4_init,php4_execute,php4_auth_trans" shlib="/op
t/netscape/suitespot/bin/libphp4.so"
Init fn="php4_init" LateInit="yes" errorString="Failed to initialize PHP!" [php_
ini="/path/to/php.ini"]

       (PHP  >=  4.3.3)  The php_ini parameter is optional but with it you
       can place your php.ini in your web server config directory.
    3. Configure  the  default  object  in  obj.conf  (for  virtual server
       classes [version 6.0+] in their vserver.obj.conf):
<Object name="default">
.
.
.
.#NOTE this next line should happen after all 'ObjectType' and before all 'AddLo
g' lines
Service fn="php4_execute" type="magnus-internal/x-httpd-php" [inikey=value inike
y=value ...]
.
.
</Object>

       (PHP  >=  4.3.3)  As additional parameters you can add some special
       php.ini-values, for example you can set a
       docroot="/path/to/docroot"  specific to the context php4_execute is
       called.   For  boolean  ini-keys  please  use  0/1  as  value,  not
       "On","Off",...    (this    will    not    work   correctly),   e.g.
       zlib.output_compression=1 instead of zlib.output_compression="On"
    4. This  is only needed if you want to configure a directory that only
       consists of PHP scripts (same like a cgi-bin directory):
<Object name="x-httpd-php">
ObjectType fn="force-type" type="magnus-internal/x-httpd-php"
Service fn=php4_execute [inikey=value inikey=value ...]
</Object>

       After  that  you  can  configure  a directory in the Administration
       server  and  assign  it the style x-httpd-php. All files in it will
       get  executed  as  PHP.  This is nice to hide PHP usage by renaming
       files to .html.
    5. Setup of authentication: PHP authentication cannot be used with any
       other  authentication.  ALL  AUTHENTICATION  IS  PASSED TO YOUR PHP
       SCRIPT.  To configure PHP Authentication for the entire server, add
       the following line to your default object:
<Object name="default">
AuthTrans fn=php4_auth_trans
.
.
.
</Object>

    6. To use PHP Authentication on a single directory, add the following:
<Object ppath="d:\path\to\authenticated\dir\*">
AuthTrans fn=php4_auth_trans
</Object>

     Note:

     The  stacksize that PHP uses depends on the configuration of the web
     server.  If  you  get  crashes  with  very  large PHP scripts, it is
     recommended  to  raise  it  with  the  Admin  Server (in the section
     "MAGNUS EDITOR").

CGI environment and recommended modifications in php.ini

   Important  when  writing  PHP scripts is the fact that Sun JSWS/Sun ONE
   WS/iPlanet/Netscape  is a multithreaded web server. Because of that all
   requests  are  running  in the same process space (the space of the web
   server  itself) and this space has only one environment. If you want to
   get  CGI variables like PATH_INFO, HTTP_HOST etc. it is not the correct
   way  to  try  this  in  the  old PHP way with getenv() or a similar way
   (register  globals  to  environment,  $_ENV).  You  would  only get the
   environment of the running web server without any valid CGI variables!

     Note:

     Why are there (invalid) CGI variables in the environment?

     Answer:  This is because you started the web server process from the
     admin  server  which  runs the startup script of the web server, you
     wanted  to  start, as a CGI script (a CGI script inside of the admin
     server!).  This is why the environment of the started web server has
     some  CGI environment variables in it. You can test this by starting
     the  web  server not from the administration server. Use the command
     line  as root user and start it manually - you will see there are no
     CGI-like environment variables.

   Simply  change your scripts to get CGI variables in the correct way for
   PHP  4.x  by  using the superglobal $_SERVER. If you have older scripts
   which  use  $HTTP_HOST,  etc.,  you  should turn on register_globals in
   php.ini  and  change the variable order too (important: remove "E" from
   it, because you do not need the environment here):
variables_order = "GPCS"
register_globals = On

Special use for error pages or self-made directory listings (PHP >= 4.3.3)

   You  can  use  PHP  to  generate the error pages for "404 Not Found" or
   similar.  Add  the  following  line to the object in obj.conf for every
   error page you want to overwrite:
Error fn="php4_execute" code=XXX script="/path/to/script.php" [inikey=value inik
ey=value...]

   where  XXX  is  the  HTTP  error  code.  Please  delete any other Error
   directives  which  could  interfere  with yours. If you want to place a
   page  for  all  errors  that could exist, leave the code parameter out.
   Your script can get the HTTP status code with $_SERVER['ERROR_TYPE'].

   Another  possibility  is to generate self-made directory listings. Just
   create  a PHP script which displays a directory listing and replace the
   corresponding default Service line for type="magnus-internal/directory"
   in obj.conf with the following:
Service fn="php4_execute" type="magnus-internal/directory" script="/path/to/scri
pt.php" [inikey=value inikey=value...]

   For  both  error  and  directory  listing  pages  the  original URI and
   translated   URI   are   in  the  variables  $_SERVER['PATH_INFO']  and
   $_SERVER['PATH_TRANSLATED'].

Note about nsapi_virtual() and subrequests (PHP >= 4.3.3)

   The  NSAPI  module  now  supports  the nsapi_virtual() function (alias:
   virtual())  to make subrequests on the web server and insert the result
   in the web page. This function uses some undocumented features from the
   NSAPI  library.  On  Unix the module automatically looks for the needed
   functions  and  uses  them  if  available.  If  not, nsapi_virtual() is
   disabled.

     Note:

     But be warned: Support for nsapi_virtual() is EXPERIMENTAL!!!
     __________________________________________________________________
     __________________________________________________________________

CGI and command line setups

   By  default,  PHP  is built as both a CLI and CGI program, which can be
   used  for  CGI processing. If you are running a web server that PHP has
   module  support  for,  you  should  generally  go for that solution for
   performance  reasons.  However,  the  CGI  version enables users to run
   different PHP-enabled pages under different user-ids.
   Warning

   A   server   deployed   in   CGI  mode  is  open  to  several  possible
   vulnerabilities.  Please  read our CGI security section to learn how to
   defend yourself from such attacks.

Testing

   If  you  have  built  PHP  as a CGI program, you may test your build by
   typing make test. It is always a good idea to test your build. This way
   you  may  catch  a  problem  with PHP on your platform early instead of
   having to struggle with it later.

Using Variables

   Some  server  supplied  environment  variables  are  not defined in the
   current  » CGI/1.1  specification.  Only  the  following  variables are
   defined     there:     AUTH_TYPE,     CONTENT_LENGTH,     CONTENT_TYPE,
   GATEWAY_INTERFACE,     PATH_INFO,     PATH_TRANSLATED,    QUERY_STRING,
   REMOTE_ADDR,  REMOTE_HOST,  REMOTE_IDENT,  REMOTE_USER, REQUEST_METHOD,
   SCRIPT_NAME,    SERVER_NAME,    SERVER_PORT,    SERVER_PROTOCOL,    and
   SERVER_SOFTWARE.   Everything   else   should  be  treated  as  'vendor
   extensions'.
     __________________________________________________________________
     __________________________________________________________________

HP-UX specific installation notes

   This  section  contains  notes  and hints specific to installing PHP on
   HP-UX systems.

   There  are two main options for installing PHP on HP-UX systems. Either
   compile it, or install a pre-compiled binary.

   Official      pre-compiled      packages      are     located     here:
   » http://software.hp.com/

   Until  this  manual  section  is  rewritten,  the  documentation  about
   compiling  PHP  (and  related  extensions)  on  HP-UX  systems has been
   removed.  For  now,  consider  reading the following external resource:
   » Building Apache and PHP on HP-UX 11.11
     __________________________________________________________________
     __________________________________________________________________

OpenBSD installation notes

   This  section  contains  notes  and hints specific to installing PHP on
   » OpenBSD 3.6.

Using Binary Packages

   Using  binary packages to install PHP on OpenBSD is the recommended and
   simplest  method.  The core package has been separated from the various
   modules,  and  each can be installed and removed independently from the
   others.  The  files  you need can be found on your OpenBSD CD or on the
   FTP site.

   The  main  package  you  need  to install is php4-core-4.3.8.tgz, which
   contains  the  basic engine (plus gettext and iconv). Next, take a look
   at    the    module   packages,   such   as   php4-mysql-4.3.8.tgz   or
   php4-imap-4.3.8.tgz.  You need to use the phpxs command to activate and
   deactivate these modules in your php.ini.

   Example #1 OpenBSD Package Install Example
# pkg_add php4-core-4.3.8.tgz
# /usr/local/sbin/phpxs -s
# cp /usr/local/share/doc/php4/php.ini-recommended /var/www/conf/php.ini
  (add in mysql)
# pkg_add php4-mysql-4.3.8.tgz
# /usr/local/sbin/phpxs -a mysql
  (add in imap)
# pkg_add php4-imap-4.3.8.tgz
# /usr/local/sbin/phpxs -a imap
  (remove mysql as a test)
# pkg_delete php4-mysql-4.3.8
# /usr/local/sbin/phpxs -r mysql
  (install the PEAR libraries)
# pkg_add php4-pear-4.3.8.tgz

   Read  the  » packages(7)  manual page for more information about binary
   packages on OpenBSD.

Using Ports

   You  can  also  compile  up  PHP  from  source  using the » ports tree.
   However,  this is only recommended for users familiar with OpenBSD. The
   PHP  4 port is split into two sub-directories: core and extensions. The
   extensions  directory  generates  sub-packages for all of the supported
   PHP  modules.  If  you  find  you  do  not want to create some of these
   modules,  use  the  no_* FLAVOR. For example, to skip building the imap
   module, set the FLAVOR to no_imap.

Common Problems

     * The default install of Apache runs inside a » chroot(2) jail, which
       will  restrict  PHP  scripts to accessing files under /var/www. You
       will  therefore  need  to  create  a /var/www/tmp directory for PHP
       session  files to be stored, or use an alternative session backend.
       In  addition, database sockets need to be placed inside the jail or
       listen  on  the  localhost interface. If you use network functions,
       some  files  from  /etc  such as /etc/resolv.conf and /etc/services
       will  need  to be moved into /var/www/etc. The OpenBSD PEAR package
       automatically  installs  into the correct chroot directories, so no
       special  modification  is  needed  there.  More  information on the
       OpenBSD Apache is available in the » OpenBSD FAQ.
     * The  OpenBSD 3.6 package for the » gd extension requires XFree86 to
       be  installed.  If you do not wish to use some of the font features
       that  require  X11,  install  the  php4-gd-4.3.8-no_x11.tgz package
       instead.

Older Releases

   Older  releases  of  OpenBSD  used  the  FLAVORS system to compile up a
   statically  linked  PHP.  Since  it is hard to generate binary packages
   using  this  method,  it  is  now deprecated. You can still use the old
   stable ports trees if you wish, but they are unsupported by the OpenBSD
   team.  If  you have any comments about this, the current maintainer for
   the port is Anil Madhavapeddy (avsm at openbsd dot org).
     __________________________________________________________________
     __________________________________________________________________

Solaris specific installation tips

   This  section  contains  notes  and hints specific to installing PHP on
   Solaris systems.

Required software

   Solaris  installs  often lack C compilers and their related tools. Read
   this  FAQ  for  information on why using GNU versions for some of these
   tools is necessary.

   For unpacking the PHP distribution you need
     * tar
     * gzip or
     * bzip2

   For compiling PHP you need
     * gcc (recommended, other C compilers may work)
     * make
     * GNU sed

   For building extra extensions or hacking the code of PHP you might also
   need
     * flex (up to PHP 5.2)
     * re2c
     * bison
     * m4
     * autoconf
     * automake

   In  addition,  you  will  need  to  install  (and possibly compile) any
   additional  software  specific to your configuration, such as Oracle or
   MySQL.

Using Packages

   You can simplify the Solaris install process by using pkgadd to install
   most  of  your  needed components. The Image Packaging System (IPS) for
   Solaris  11  Express  also contains most of the required components for
   installation using the pkg command.
     __________________________________________________________________
     __________________________________________________________________

Debian GNU/Linux installation notes

   This  section  contains  notes  and hints specific to installing PHP on
   » Debian GNU/Linux.
   Warning

   Unofficial  builds  from third-parties are not supported here. Any bugs
   should  be  reported  to  the Debian team unless they can be reproduced
   using the latest builds from our » download area.

   While  the  instructions  for  building  PHP on Unix apply to Debian as
   well, this manual page contains specific information for other options,
   such as using either the apt-get or aptitude commands. This manual page
   uses these two commands interchangeably.

Using APT

   First,   note   that   other  related  packages  may  be  desired  like
   libapache2-mod-php5 to integrate with Apache 2, and php-pear for PEAR.

   Second,  before  installing  a package, it's wise to ensure the package
   list  is  up  to  date.  Typically, this is done by running the command
   apt-get update.

   Example #1 Debian Install Example with Apache 2
# apt-get install php5-common libapache2-mod-php5 php5-cli

   APT will automatically install the PHP 5 module for Apache 2 and all of
   its  dependencies,  and then activate it. Apache should be restarted in
   order for the changes take place. For example:

   Example #2 Stopping and starting Apache once PHP is installed
# /etc/init.d/apache2 stop
# /etc/init.d/apache2 start

Better control of configuration

   In  the  last  section,  PHP was installed with only core modules. It's
   very  likely  that  additional  modules will be desired, such as MySQL,
   cURL, GD, etc. These may also be installed via the apt-get command.

   Example #3 Methods for listing additional PHP 5 packages
# apt-cache search php5
# aptitude search php5
# aptitude search php5 |grep -i mysql

   The examples will show a lot of packages including several PHP specific
   ones  like  php5-cgi, php5-cli and php5-dev. Determine which are needed
   and  install  them  like any other with either apt-get or aptitude. And
   because  Debian  performs  dependency checks, it'll prompt for those so
   for example to install MySQL and cURL:

   Example #4 Install PHP with MySQL, cURL
# apt-get install php5-mysql php5-curl

   APT  will  automatically  add  the  appropriate  lines to the different
   php.ini      related      files     like     /etc/php5/apache2/php.ini,
   /etc/php5/conf.d/pdo.ini,  etc. and depending on the extension will add
   entries similar to extension=foo.so. However, restarting the web server
   (like Apache) is required before these changes take affect.

Common Problems

     * If  the  PHP  scripts are not parsing via the web server, then it's
       likely  that  PHP  was  not added to the web server's configuration
       file,  which on Debian may be /etc/apache2/apache2.conf or similar.
       See the Debian manual for further details.
     * If  an  extension  was  seemingly  installed  yet the functions are
       undefined,  be  sure  that the appropriate ini file is being loaded
       and/or the web server was restarted after installation.
     * There are two basic commands for installing packages on Debian (and
       other  linux  variants):  apt-get and aptitude. However, explaining
       the subtle differences between these commands goes beyond the scope
       of this manual.
     __________________________________________________________________
     __________________________________________________________________
     __________________________________________________________________

Installation on Mac OS X

Table of Contents

     * Using Packages
     * Using the bundled PHP
     * Compiling PHP on Mac OS X

   This section contains notes and hints specific to installing PHP on Mac
   OS  X.  PHP  is bundled with Macs, and compiling is similar to the Unix
   installation guide.
     __________________________________________________________________

Using Packages

   There  are  a few pre-packaged and pre-compiled versions of PHP for Mac
   OS  X. This can help in setting up a standard configuration, but if you
   need to have a different set of features (such as a secure server, or a
   different  database  driver), you may need to build PHP and/or your web
   server yourself. If you are unfamiliar with building and compiling your
   own  software, it's worth checking whether somebody has already built a
   packaged version of PHP with the features you need.

   The  following resources offer easy to install packages and precompiled
   binaries for PHP on Mac OS:

     * MacPorts: » http://www.macports.org/
     * Entropy: » http://www.entropy.ch/software/macosx/php/
     * Fink: » http://www.finkproject.org/
     * Homebrew: » http://github.com/mxcl/homebrew
     __________________________________________________________________
     __________________________________________________________________

Using the bundled PHP

   PHP has come standard with Macs since OS X version 10.0.0. Enabling PHP
   with  the  default  web server requires uncommenting a few lines in the
   Apache  configuration  file  httpd.conf  whereas the CGI and/or CLI are
   enabled by default (easily accessible via the Terminal program).

   Enabling  PHP using the instructions below is meant for quickly setting
   up  a  local development environment. It's highly recommended to always
   upgrade  PHP  to  the  newest  version.  Like most live software, newer
   versions  are  created to fix bugs and add features and PHP being is no
   different.  See the appropriate MAC OS X installation documentation for
   further  details.  The  following  instructions  are  geared  towards a
   beginner with details provided for getting a default setup to work. All
   users are encouraged to compile, or install a new packaged version.

   The  standard  installation  type  is  using  mod_php, and enabling the
   bundled  mod_php on Mac OS X for the Apache web server (the default web
   server,  that  is  accessible  via  System  Preferences)  involves  the
   following steps:

    1. Locate  and  open  the  Apache  configuration file. By default, the
       location   is  as  follows:  /private/etc/apache2/httpd.conf  Using
       Finder  or  Spotlight  to  find this file may prove difficult as by
       default it's private and owned by the root user.

     Note:  One  way to open this is by using a Unix based text editor in
     the  Terminal,  for  example  nano, and because the file is owned by
     root  we'll use the sudo command to open it (as root) so for example
     type  the  following  into  the Terminal Application (after, it will
     prompt  for  a  password): sudo nano /private/etc/apache2/httpd.conf
     Noteworthy  nano  commands:  ^w  (search),  ^o (save), and ^x (exit)
     where ^ represents the Ctrl key.

     Note:  Versions  of  Mac  OS X prior to 10.5 were bundled with older
     versions  of  PHP and Apache. As such, the Apache configuration file
     on legacy machines may be /etc/httpd/httpd.conf.
    2. With  a  text  editor, uncomment the lines (by removing the #) that
       look  similar  to  the  following  (these  two  lines are often not
       together, locate them both in the file):
# LoadModule php5_module libexec/httpd/libphp5.so

# AddModule mod_php5.c

       Notice  the  location/path.  When  building  PHP in the future, the
       above files should be replaced or commented out.
    3. Be  sure  the  desired extensions will parse as PHP (examples: .php
       .html and .inc)
       Due  to  the following statement already existing in httpd.conf (as
       of   Mac  Panther),  once  PHP  is  enabled  the  .php  files  will
       automatically parse as PHP.
<IfModule mod_php5.c>
    # If php is turned on, we respect .php and .phps files.
    AddType application/x-httpd-php .php
    AddType application/x-httpd-php-source .phps

    # Since most users will want index.php to work we
    # also automatically enable index.php
    <IfModule mod_dir.c>
        DirectoryIndex index.html index.php
    </IfModule>
</IfModule>

     Note:
     Before  OS  X  10.5 (Leopard), PHP 4 was bundled instead of PHP 5 in
     which  case  the above instructions will differ slightly by changing
     5's to 4's.
    4. Be  sure  the  DirectoryIndex  loads the desired default index file
       This  is also set in httpd.conf. Typically index.php and index.html
       are  used. By default index.php is enabled because it's also in the
       PHP check shown above. Adjust accordingly.
    5. Set  the  php.ini  location  or  use  the default A typical default
       location  on  Mac  OS  X  is  /usr/local/php/php.ini  and a call to
       phpinfo()  will  reveal this information. If a php.ini is not used,
       PHP  will  use  all  default  values.  See  also the related FAQ on
       finding php.ini.
    6. Locate  or  set the DocumentRoot This is the root directory for all
       the  web  files.  Files  in  this directory are served from the web
       server so the PHP files will parse as PHP before outputting them to
       the browser. A typical default path is /Library/WebServer/Documents
       but  this  can be set to anything in httpd.conf. Alternatively, the
       default      DocumentRoot      for      individual     users     is
       /Users/yourusername/Sites
    7. Create a phpinfo() file
       The phpinfo() function will display information about PHP. Consider
       creating a file in the DocumentRoot with the following PHP code:
       <?php phpinfo(); ?>
    8. Restart  Apache,  and  load  the PHP file created above To restart,
       either  execute  sudo apachectl graceful in the shell or stop/start
       the "Personal Web Server" option in the OS X System Preferences. By
       default,  loading  local files in the browser will have an URL like
       so: http://localhost/info.php Or using the DocumentRoot in the user
       directory  is  another  option  and  would  end  up  looking  like:
       http://localhost/~yourusername/info.php

   The  CLI  (or  CGI  in  older  versions) is appropriately named php and
   likely  exists  as /usr/bin/php. Open up the terminal, read the command
   line  section  of  the  PHP manual, and execute php -v to check the PHP
   version  of  this PHP binary. A call to phpinfo() will also reveal this
   information.
     __________________________________________________________________
     __________________________________________________________________

Compiling PHP on Mac OS X

   Use the Unix installation guide to compile PHP on Mac OS X.
     __________________________________________________________________
     __________________________________________________________________
     __________________________________________________________________

Installation of PECL extensions

Table of Contents

     * Introduction to PECL Installations
     * Downloading PECL extensions
     * Installing a PHP extension on Windows
     * Compiling shared PECL extensions with the pecl command
     * Compiling shared PECL extensions with phpize
     * php-config
     * Compiling PECL extensions statically into PHP
     __________________________________________________________________

Introduction to PECL Installations

   » PECL is a repository of PHP extensions that are made available to you
   via the » PEAR packaging system. This section of the manual is intended
   to demonstrate how to obtain and install PECL extensions.

   These  instructions  assume  /your/phpsrcdir/  is  the  path to the PHP
   source  distribution,  and  that  extname  is  the  name  of  the  PECL
   extension.   Adjust  accordingly.  These  instructions  also  assume  a
   familiarity with the » pear command. The information in the PEAR manual
   for the pear command also applies to the pecl command.

   To  be useful, a shared extension must be built, installed, and loaded.
   The  methods  described  below provide you with various instructions on
   how  to build and install the extensions, but they do not automatically
   load  them.  Extensions can be loaded by adding an extension directive.
   To this php.ini file, or through the use of the dl() function.

   When  building  PHP modules, it's important to have known-good versions
   of  the  required  tools  (autoconf,  automake,  libtool, etc.) See the
   » Anonymous  Git  Instructions  for  details on the required tools, and
   required versions.
     __________________________________________________________________
     __________________________________________________________________

Downloading PECL extensions

   There are several options for downloading PECL extensions, such as:
     * The  pecl  install  extname  command  downloads the extensions code
       automatically,  so  in  this  case  there is no need for a separate
       download.
     * » http://pecl.php.net/ The PECL web site contains information about
       the  different  extensions  that are offered by the PHP Development
       Team.  The  information available here includes: ChangeLog, release
       notes, requirements and other similar details.
     * pecl  download extname PECL extensions that have releases listed on
       the PECL web site are available for download and installation using
       the » pecl command. Specific revisions may also be specified.
     * SVN  Most  PECL extensions also reside in SVN. A web-based view may
       be  seen at » http://svn.php.net/viewvc/pecl/. To download straight
       from SVN, the following sequence of commands may be used:
       $   svn  checkout  http://svn.php.net/repository/pecl/extname/trunk
       extname
     * Windows  downloads  At  this  time the PHP project does not compile
       Windows binaries for PECL extensions. However, to compile PHP under
       Windows see the chapter titled building PHP for Windows.
     __________________________________________________________________
     __________________________________________________________________

Installing a PHP extension on Windows

   On  Windows,  you have two ways to load a PHP extension: either compile
   it  into  PHP, or load the DLL. Loading a pre-compiled extension is the
   easiest and preferred way.

   To load an extension, you need to have it available as a ".dll" file on
   your  system.  All  the  extensions  are automatically and periodically
   compiled by the PHP Group (see next section for the download).

   To  compile an extension into PHP, please refer to building from source
   documentation.

   To  compile  a  standalone  extension (aka a DLL file), please refer to
   building  from  source  documentation.  If  the  DLL  file is available
   neither with your PHP distribution nor in PECL, you may have to compile
   it before you can start using the extension.

Where to find an extension?

   PHP   extensions   are  usually  called  "php_*.dll"  (where  the  star
   represents  the  name  of the extension) and they are located under the
   "PHP\ext" ("PHP\extensions" in PHP 4) folder.

   PHP   ships  with  the  extensions  most  useful  to  the  majority  of
   developers. They are called "core" extensions.

   However,  if you need functionality not provided by any core extension,
   you  may still be able to find one in PECL. The PHP Extension Community
   Library  (PECL)  is  a  repository  for  PHP  Extensions,  providing  a
   directory   of   all   known  extensions  and  hosting  facilities  for
   downloading and development of PHP extensions.

   If you have developed an extension for your own uses, you might want to
   think  about  hosting it on PECL so that others with the same needs can
   benefit from your time. A nice side effect is that you give them a good
   chance  to  give you feedback, (hopefully) thanks, bug reports and even
   fixes/patches.  Before  you  submit your extension for hosting on PECL,
   please read http://pecl.php.net/package-new.php.

Which extension to download?

   Many times, you will find several versions of each DLL:
     * Different  version  numbers  (at least the first two numbers should
       match)
     * Different thread safety settings
     * Different processor architecture (x86, x64, ...)
     * Different debugging settings
     * etc.

   You  should  keep in mind that your extension settings should match all
   the  settings  of  the  PHP executable you are using. The following PHP
   script will tell you all about your PHP settings:

   Example #1 phpinfo() call
   <?php
   phpinfo();
   ?>

   Or from the command line, run:
drive:\\path\to\php\executable\php.exe -i

Loading an extension

   The  most  common  way to load a PHP extension is to include it in your
   php.ini  configuration  file.  Please  note  that  many  extensions are
   already  present  in  your php.ini and that you only need to remove the
   semicolon to activate them.
;extension=php_extname.dll

extension=php_extname.dll

   However,  some  web  servers  are confusing because they do not use the
   php.ini  located  alongside your PHP executable. To find out where your
   actual php.ini resides, look for its path in phpinfo():
Configuration File (php.ini) Path  C:\WINDOWS

Loaded Configuration File   C:\Program Files\PHP\5.2\php.ini

   After activating an extension, save php.ini, restart the web server and
   check  phpinfo()  again.  The  new  extension  should  now have its own
   section.

Resolving problems

   If  the  extension  does not appear in phpinfo(), you should check your
   logs to learn where the problem comes from.

   If you are using PHP from the command line (CLI), the extension loading
   error can be read directly on screen.

   If  you are using PHP with a web server, the location and format of the
   logs  vary  depending  on  your  software.  Please read your web server
   documentation  to  locate  the logs, as it does not have anything to do
   with PHP itself.

   Common  problems  are  the  location  of  the  DLL,  the value of the "
   extension_dir"   setting   inside   php.ini  and  compile-time  setting
   mismatches.

   If  the  problem  lies in a compile-time setting mismatch, you probably
   didn't download the right DLL. Try downloading again the extension with
   the right settings. Again, phpinfo() can be of great help.
     __________________________________________________________________
     __________________________________________________________________

Compiling shared PECL extensions with the pecl command

   PECL  makes  it  easy to create shared PHP extensions. Using the » pecl
   command, do the following:

   $ pecl install extname

   This  will  download  the  source  for  extname,  compile,  and install
   extname.so  into  your extension_dir. extname.so may then be loaded via
   php.ini

   By  default, the pecl command will not install packages that are marked
   with  the alpha or beta state. If no stable packages are available, you
   may install a beta package using the following command:

   $ pecl install extname-beta

   You may also install a specific version using this variant:

   $ pecl install extname-0.1

     Note:

     After  enabling the extension in php.ini, restarting the web service
     is required for the changes to be picked up.
     __________________________________________________________________
     __________________________________________________________________

Compiling shared PECL extensions with phpize

   Sometimes,  using  the  pecl  installer is not an option. This could be
   because  you're behind a firewall, or it could be because the extension
   you want to install is not available as a PECL compatible package, such
   as  unreleased  extensions  from  SVN.  If  you  need  to build such an
   extension, you can use the lower-level build tools to perform the build
   manually.

   The  phpize  command is used to prepare the build environment for a PHP
   extension. In the following sample, the sources for an extension are in
   a directory named extname:

$ cd extname
$ phpize
$ ./configure
$ make
# make install

   A  successful  install will have created extname.so and put it into the
   PHP  extensions directory. You'll need to and adjust php.ini and add an
   extension=extname.so line before you can use the extension.

   If  the  system is missing the phpize command, and precompiled packages
   (like  RPM's)  are  used, be sure to also install the appropriate devel
   version  of  the  PHP  package as they often include the phpize command
   along   with  the  appropriate  header  files  to  build  PHP  and  its
   extensions.

   Execute phpize --help to display additional usage information.
     __________________________________________________________________
     __________________________________________________________________

php-config

   php-config is a simple shell script for obtaining information about the
   installed PHP configuration.

   When compiling extensions, if you have multiple PHP versions installed,
   you may specify for which installation you'd like to build by using the
   --with-php-config  option  during configuration, specifying the path of
   the respective php-config script.

   The  list of command line options provided by the php-config script can
   be queried anytime by running php-config with the -h switch:
Usage: /usr/local/bin/php-config [OPTION]
Options:
  --prefix            [...]
  --includes          [...]
  --ldflags           [...]
  --libs              [...]
  --extension-dir     [...]
  --include-dir       [...]
  --php-binary        [...]
  --php-sapis         [...]
  --configure-options [...]
  --version           [...]
  --vernum            [...]

   CAPTION: Command line options

   Option Description
   --prefix Directory prefix where PHP is installed, e.g. /usr/local
   --includes List of -I options with all include files
   --ldflags LD Flags which PHP was compiled with
   --libs Extra libraries which PHP was compiled with
   --extension-dir Directory where extensions are searched by default
   --include-dir Directory prefix where header files are installed by
   default
   --php-binary Full path to php CLI or CGI binary
   --php-sapis Show all SAPI modules available
   --configure-options  Configure  options  to  recreate  configuration of
   current PHP installation
   --version PHP version
   --vernum PHP version as integer
     __________________________________________________________________
     __________________________________________________________________

Compiling PECL extensions statically into PHP

   You  might find that you need to build a PECL extension statically into
   your  PHP binary. To do this, you'll need to place the extension source
   under  the  php-src/ext/  directory  and  tell  the PHP build system to
   regenerate its configure script.

$ cd /your/phpsrcdir/ext
$ pecl download extname
$ gzip -d < extname.tgz | tar -xvf -
$ mv extname-x.x.x extname

   This will result in the following directory:

   /your/phpsrcdir/ext/extname

   From  here,  force  PHP to rebuild the configure script, and then build
   PHP as normal:

   $ cd /your/phpsrcdir
   $ rm configure
   $ ./buildconf --force
   $ ./configure --help
   $ ./configure --with-extname --enable-someotherext --with-foobar
   $ make
   $ make install

     Note:  To  run  the  'buildconf'  script  you need autoconf 2.13 and
     automake  1.4+  (newer  versions  of  autoconf may work, but are not
     supported).

   Whether  --enable-extname  or  --with-extname  is  used  depends on the
   extension.  Typically  an  extension  that  does  not  require external
   libraries uses --enable. To be sure, run the following after buildconf:

   $ ./configure --help | grep extname
     __________________________________________________________________
     __________________________________________________________________
     __________________________________________________________________

Problems?

Table of Contents

     * Read the FAQ
     * Other problems
     * Bug reports
     __________________________________________________________________

Read the FAQ

   Some  problems  are  more  common than others. The most common ones are
   listed in the PHP FAQ, part of this manual.
     __________________________________________________________________
     __________________________________________________________________

Other problems

   If  you  are  still stuck, someone on the PHP installation mailing list
   may  be  able  to  help you. You should check out the archive first, in
   case  someone already answered someone else who had the same problem as
   you.   The   archives   are   available   from   the  support  page  on
   » http://www.php.net/support.php.  To subscribe to the PHP installation
   mailing list, send an empty mail to
   » php-install-subscribe@lists.php.net.  The  mailing  list  address  is
   » php-install@lists.php.net.

   If  you  want to get help on the mailing list, please try to be precise
   and  give the necessary details about your environment (which operating
   system,  what  PHP  version, what web server, if you are running PHP as
   CGI or a server module, safe mode, etc.), and preferably enough code to
   make others able to reproduce and test your problem.
     __________________________________________________________________
     __________________________________________________________________

Bug reports

   If  you  think  you  have found a bug in PHP, please report it. The PHP
   developers  probably  don't  know  about  it, and unless you report it,
   chances  are  it  won't  be  fixed.  You  can  report  bugs  using  the
   bug-tracking  system  at » http://bugs.php.net/. Please do not send bug
   reports  in  mailing  list  or personal letters. The bug system is also
   suitable to submit feature requests.

   Read  the  » How  to  report  a  bug document before submitting any bug
   reports!
     __________________________________________________________________
     __________________________________________________________________
     __________________________________________________________________

Runtime Configuration

Table of Contents

     * The configuration file
     * .user.ini files
     * Where a configuration setting may be set
     * How to change configuration settings
     __________________________________________________________________

The configuration file

   The  configuration  file  (php.ini) is read when PHP starts up. For the
   server  module  versions  of  PHP,  this happens only once when the web
   server  is  started.  For the CGI and CLI versions, it happens on every
   invocation.

   php.ini is searched for in these locations (in order):
     * SAPI  module specific location (PHPIniDir directive in Apache 2, -c
       command  line  option  in  CGI and CLI, php_ini parameter in NSAPI,
       PHP_INI_PATH environment variable in THTTPD)
     * The  PHPRC environment variable. Before PHP 5.2.0, this was checked
       after the registry key mentioned below.
     * As  of  PHP  5.2.0, the location of the php.ini file can be set for
       different versions of PHP. The following registry keys are examined
       in          order:         [HKEY_LOCAL_MACHINE\SOFTWARE\PHP\x.y.z],
       [HKEY_LOCAL_MACHINE\SOFTWARE\PHP\x.y] and
       [HKEY_LOCAL_MACHINE\SOFTWARE\PHP\x],  where x, y and z mean the PHP
       major,  minor  and  release  versions.  If  there  is  a  value for
       IniFilePath  in any of these keys, the first one found will be used
       as the location of the php.ini (Windows only).
     * [HKEY_LOCAL_MACHINE\SOFTWARE\PHP],  value  of  IniFilePath (Windows
       only).
     * Current working directory (except CLI).
     * The  web server's directory (for SAPI modules), or directory of PHP
       (otherwise in Windows).
     * Windows  directory  (C:\windows  or  C:\winnt)  (for  Windows),  or
       --with-config-file-path compile time option.

   If php-SAPI.ini exists (where SAPI is the SAPI in use, so, for example,
   php-cli.ini or php-apache.ini), it is used instead of php.ini. The SAPI
   name can be determined with php_sapi_name().

     Note:

     The  Apache  web  server  changes  the directory to root at startup,
     causing  PHP  to attempt to read php.ini from the root filesystem if
     it exists.

   The  php.ini  directives  handled  by  extensions are documented on the
   respective  pages  of  the  extensions  themselves.  A list of the core
   directives  is  available  in  the appendix. Not all PHP directives are
   necessarily   documented  in  this  manual:  for  a  complete  list  of
   directives  available  in  your  PHP  version,  please  read  your well
   commented  php.ini  file.  Alternatively,  you  may  find  » the latest
   php.ini from Git helpful too.

   Example #1 php.ini example
; any text on a line after an unquoted semicolon (;) is ignored
[php] ; section markers (text within square brackets) are also ignored
; Boolean values can be set to either:
;    true, on, yes
; or false, off, no, none
register_globals = off
track_errors = yes

; you can enclose strings in double-quotes
include_path = ".:/usr/local/lib/php"

; backslashes are treated the same as any other character
include_path = ".;c:\php\lib"

   Since  PHP  5.1.0,  it  is possible to refer to existing .ini variables
   from   within  .ini  files.  Example:  open_basedir  =  ${open_basedir}
   ":/new/dir".
     __________________________________________________________________
     __________________________________________________________________

.user.ini files

   Since  PHP 5.3.0, PHP includes support for .htaccess-style INI files on
   a   per-directory   basis.  These  files  are  processed  only  by  the
   CGI/FastCGI  SAPI.  This  functionality  obsoletes  the  PECL htscanner
   extension.  If  you  are using Apache, use .htaccess files for the same
   effect.

   In  addition  to the main php.ini file, PHP scans for INI files in each
   directory,  starting  with the directory of the requested PHP file, and
   working   its   way  up  to  the  current  document  root  (as  set  in
   $_SERVER['DOCUMENT_ROOT']).  In  case  the  PHP  file  is  outside  the
   document root, only its directory is scanned.

   Only  INI  settings with the modes PHP_INI_PERDIR and PHP_INI_USER will
   be recognized in .user.ini-style INI files.

   Two   new  INI  directives,  user_ini.filename  and  user_ini.cache_ttl
   control the use of user INI files.

   user_ini.filename  sets  the  name  of  the  file PHP looks for in each
   directory;  if  set  to  an  empty string, PHP doesn't scan at all. The
   default is .user.ini.

   user_ini.cache_ttl  controls  how often user INI files are re-read. The
   default is 300 seconds (5 minutes).
     __________________________________________________________________
     __________________________________________________________________

Where a configuration setting may be set

   These  modes determine when and where a PHP directive may or may not be
   set, and each directive within the manual refers to one of these modes.
   For  example,  some  settings  may  be  set  within  a PHP script using
   ini_set(), whereas others may require php.ini or httpd.conf.

   For  example,  the output_buffering setting is PHP_INI_PERDIR therefore
   it  may  not  be  set  using  ini_set().  However,  the  display_errors
   directive  is  PHP_INI_ALL  therefore it may be set anywhere, including
   with ini_set().

   CAPTION: Definition of PHP_INI_* modes

   Mode Meaning
   PHP_INI_USER Entry can be set in user scripts (like with ini_set()) or
   in the Windows registry. Since PHP 5.3, entry can be set in .user.ini
   PHP_INI_PERDIR  Entry  can  be set in php.ini, .htaccess, httpd.conf or
   .user.ini (since PHP 5.3)
   PHP_INI_SYSTEM Entry can be set in php.ini or httpd.conf
   PHP_INI_ALL Entry can be set anywhere
     __________________________________________________________________
     __________________________________________________________________

How to change configuration settings

Running PHP as an Apache module

   When   using  PHP  as  an  Apache  module,  you  can  also  change  the
   configuration  settings  using directives in Apache configuration files
   (e.g.  httpd.conf)  and  .htaccess  files. You will need "AllowOverride
   Options" or "AllowOverride All" privileges to do so.

   There  are  several  Apache directives that allow you to change the PHP
   configuration from within the Apache configuration files. For a listing
   of which directives are PHP_INI_ALL, PHP_INI_PERDIR, or PHP_INI_SYSTEM,
   have a look at the List of php.ini directives appendix.

   php_value name value
          Sets the value of the specified directive. Can be used only with
          PHP_INI_ALL  and  PHP_INI_PERDIR  type  directives.  To  clear a
          previously set value use none as the value.

     Note:  Don't  use  php_value  to  set  boolean values. php_flag (see
     below) should be used instead.

   php_flag name on|off
          Used  to set a boolean configuration directive. Can be used only
          with PHP_INI_ALL and PHP_INI_PERDIR type directives.

   php_admin_value name value
          Sets  the value of the specified directive. This can not be used
          in  .htaccess files. Any directive type set with php_admin_value
          can  not  be  overridden  by  .htaccess or ini_set(). To clear a
          previously set value use none as the value.

   php_admin_flag name on|off
          Used  to  set a boolean configuration directive. This can not be
          used   in   .htaccess   files.   Any  directive  type  set  with
          php_admin_flag can not be overridden by .htaccess or ini_set().

   Example #1 Apache configuration example
<IfModule mod_php5.c>
  php_value include_path ".:/usr/local/lib/php"
  php_admin_flag engine on
</IfModule>
<IfModule mod_php4.c>
  php_value include_path ".:/usr/local/lib/php"
  php_admin_flag engine on
</IfModule>

   Caution

   PHP  constants  do not exist outside of PHP. For example, in httpd.conf
   you  can  not  use  PHP  constants such as E_ALL or E_NOTICE to set the
   error_reporting  directive  as  they  will  have  no  meaning  and will
   evaluate  to  0.  Use  the  associated  bitmask  values  instead. These
   constants can be used in php.ini

Changing PHP configuration via the Windows registry

   When  running  PHP on Windows, the configuration values can be modified
   on  a per-directory basis using the Windows registry. The configuration
   values  are  stored in the registry key HKLM\SOFTWARE\PHP\Per Directory
   Values,  in  the sub-keys corresponding to the path names. For example,
   configuration  values  for  the  directory  c:\inetpub\wwwroot would be
   stored      in      the     key     HKLM\SOFTWARE\PHP\Per     Directory
   Values\c\inetpub\wwwroot.  The  settings  for  the  directory  would be
   active  for  any script running from this directory or any subdirectory
   of  it.  The  values  under  the  key  should  have the name of the PHP
   configuration  directive  and  the  string  value. PHP constants in the
   values are not parsed. However, only configuration values changeable in
   PHP_INI_USER can be set this way, PHP_INI_PERDIR values can not.

Other interfaces to PHP

   Regardless of how you run PHP, you can change certain values at runtime
   of  your  scripts  through  ini_set().  See  the  documentation  on the
   ini_set() page for more information.

   If  you  are interested in a complete list of configuration settings on
   your  system  with  their current values, you can execute the phpinfo()
   function, and review the resulting page. You can also access the values
   of  individual  configuration  directives at runtime using ini_get() or
   get_cfg_var().
     __________________________________________________________________
     __________________________________________________________________
     __________________________________________________________________

Installation

   This  section  holds common questions about the way to install PHP. PHP
   is available for almost any OS (except maybe for MacOS before OSX), and
   almost any web server.

   To install PHP, follow the instructions in Installing PHP.
    1. Why  shouldn't  I  use  Apache2 with a threaded MPM in a production
       environment?
    2. Unix/Windows: Where should my php.ini file be located?
    3. Unix:  I installed PHP, but every time I load a document, I get the
       message 'Document Contains No Data'! What's going on here?
    4. Unix:  I  installed PHP using RPMS, but Apache isn't processing the
       PHP pages! What's going on here?
    5. Unix:  I  patched  Apache  with the FrontPage extensions patch, and
       suddenly  PHP  stopped working. Is PHP incompatible with the Apache
       FrontPage extensions?
    6. Unix/Windows:  I have installed PHP, but when I try to access a PHP
       script file via my browser, I get a blank screen.
    7. Unix/Windows:  I  have  installed PHP, but when try to access a PHP
       script file via my browser, I get a server 500 error.
    8. Some  operating  systems:  I have installed PHP without errors, but
       when  I  try  to  start  Apache  I  get  undefined  symbol  errors:
       [mybox:user   /src/php5]   root#  apachectl  configtest  apachectl:
       /usr/local/apache/bin/httpd     Undefined     symbols:    _compress
       _uncompress
    9. Windows:  I  have  installed  PHP,  but  when I try to access a PHP
       script  file  via  my  browser,  I  get  the  error: cgi error: The
       specified  CGI  application  misbehaved by not returning a complete
       set of HTTP headers. The headers it did return are:
   10. Windows:  I've  followed  all the instructions, but still can't get
       PHP and IIS to work together!
   11. When  running  PHP as CGI with IIS, OmniHTTPD or Xitami, I get
       the  following  error:  Security  Alert! PHP CGI cannot be accessed
       directly..
   12. How  do I know if my php.ini is being found and read? It seems like
       it isn't as my changes aren't being implemented.
   13. How do I add my PHP directory to the PATH on Windows?
   14. How do I make the php.ini file available to PHP on windows?
   15. Is  it  possible  to  use  Apache  content  negotiation (MultiViews
       option) with PHP?
   16. Is PHP limited to process GET and POST request methods only?

   Why shouldn't I use Apache2 with a threaded MPM in a production
          environment?
          PHP  is glue. It is the glue used to build cool web applications
          by sticking dozens of 3rd-party libraries together and making it
          all  appear as one coherent entity through an intuitive and easy
          to  learn  language  interface. The flexibility and power of PHP
          relies  on  the  stability  and  robustness  of  the  underlying
          platform.  It  needs  a  working  OS,  a  working web server and
          working  3rd-party libraries to glue together. When any of these
          stop  working  PHP  needs  ways to identify the problems and fix
          them  quickly.  When  you  make  the  underlying  framework more
          complex  by  not  having  completely separate execution threads,
          completely  separate  memory  segments  and a strong sandbox for
          each  request to play in, further weaknesses are introduced into
          PHP's system.

          If   you  want  to  use  a  threaded  MPM,  look  at  a  FastCGI
          configuration where PHP is running in its own memory space.

   Unix/Windows: Where should my php.ini file be located?
          By  default  on  Unix  it  should  be in /usr/local/lib which is
          <install-path>/lib.  Most  people  will  want  to change this at
          compile-time  with  the --with-config-file-path flag. You would,
          for example, set it with something like:

--with-config-file-path=/etc

          And   then   you   would   copy   php.ini-development  from  the
          distribution  to  /etc/php.ini  and  edit  it  to make any local
          changes you want.

--with-config-file-scan-dir=PATH

          On  Windows the default path for the php.ini file is the Windows
          directory.  If  you're  using  the  Apache webserver, php.ini is
          first searched in the Apaches install directory, e.g. c:\program
          files\apache  group\apache.  This  way  you  can  have different
          php.ini  files  for  different  versions  of  Apache on the same
          machine.

          See also the chapter about the configuration file.

   Unix: I installed PHP, but every time I load a document, I get the
          message 'Document Contains No Data'! What's going on here?
          This  probably means that PHP is having some sort of problem and
          is core-dumping. Look in your server error log to see if this is
          the  case,  and  then  try to reproduce the problem with a small
          test case. If you know how to use 'gdb', it is very helpful when
          you  can  provide  a  backtrace with your bug report to help the
          developers  pinpoint  the  problem.  If  you are using PHP as an
          Apache module try something like:

          + Stop your httpd processes
          + gdb httpd
          + Stop your httpd processes
          + > run -X -f /path/to/httpd.conf
          + Then fetch the URL causing the problem with your browser
          + > run -X -f /path/to/httpd.conf
          + If  you are getting a core dump, gdb should inform you of this
            now
          + type: bt
          + You  should  include  your  backtrace in your bug report. This
            should be submitted to » http://bugs.php.net/

          If   your   script   uses   the   regular  expression  functions
          (preg_match()  and  friends),  you  should  make  sure  that you
          compiled  PHP  and  Apache  with  the  same  regular  expression
          package.  This  should  happen automatically with PHP and Apache
          1.3.x

   Unix: I installed PHP using RPMS, but Apache isn't processing the PHP
          pages! What's going on here?
          Assuming  you  installed  both Apache and PHP from RPM packages,
          you  need to uncomment or add some or all of the following lines
          in your httpd.conf file:

# Extra Modules
AddModule mod_php.c
AddModule mod_perl.c

# Extra Modules
LoadModule php_module         modules/mod_php.so
LoadModule php5_module        modules/libphp5.so
LoadModule perl_module        modules/libperl.so

          And add:

AddType application/x-httpd-php .php

          ...  to  the  global  properties,  or  to  the properties of the
          VirtualDomain you want to have PHP support added to.

   Unix: I patched Apache with the FrontPage extensions patch, and
          suddenly PHP stopped working. Is PHP incompatible with the
          Apache FrontPage extensions?
          No, PHP works fine with the FrontPage extensions. The problem is
          that  the  FrontPage  patch  modifies several Apache structures,
          that  PHP relies on. Recompiling PHP (using 'make clean ; make')
          after the FP patch is applied would solve the problem.

   Unix/Windows: I have installed PHP, but when I try to access a PHP
          script file via my browser, I get a blank screen.
          Do a 'view source' in the web browser and you will probably find
          that  you can see the source code of your PHP script. This means
          that  the  web  server  did  not  send  the  script  to  PHP for
          interpretation. Something is wrong with the server configuration
          -   double  check  the  server  configuration  against  the  PHP
          installation instructions.

   Unix/Windows: I have installed PHP, but when try to access a PHP script
          file via my browser, I get a server 500 error.
          Something went wrong when the server tried to run PHP. To get to
          see  a  sensible error message, from the command line, change to
          the directory containing the PHP executable (php.exe on Windows)
          and run php -i. If PHP has any problems running, then a suitable
          error message will be displayed which will give you a clue as to
          what  needs  to  be  done next. If you get a screen full of HTML
          codes  (the  output  of  the  phpinfo()  function)  then  PHP is
          working,  and  your  problem  may  be  related  to  your  server
          configuration which you should double check.

   Some operating systems: I have installed PHP without errors, but when I
          try to start Apache I get undefined symbol errors:

[mybox:user /src/php5] root# apachectl configtest
 apachectl: /usr/local/apache/bin/httpd Undefined symbols:
  _compress
  _uncompress

          This  has  actually  nothing  to do with PHP, but with the MySQL
          client libraries. Some need --with-zlib , others do not. This is
          also covered in the MySQL FAQ.

   Windows: I have installed PHP, but when I try to access a PHP script
          file via my browser, I get the error:

cgi error:
 The specified CGI application misbehaved by not
 returning a complete set of HTTP headers.
 The headers it did return are:

          This  error  message means that PHP failed to output anything at
          all.  To  get  to see a sensible error message, from the command
          line,  change  to  the  directory  containing the PHP executable
          (php.exe  on  Windows)  and  run php -i. If PHP has any problems
          running,  then  a suitable error message will be displayed which
          will  give  you  a clue as to what needs to be done next. If you
          get  a  screen  full  of HTML codes (the output of the phpinfo()
          function) then PHP is working.

          Once  PHP  is  working  at  the  command line, try accessing the
          script via the browser again. If it still fails then it could be
          one of the following:

          + File  permissions  on  your  PHP  script, php.exe, php5ts.dll,
            php.ini  or any PHP extensions you are trying to load are such
            that  the  anonymous  internet  user ISUR_<machinename> cannot
            access them.
          + The  script  file  does not exist (or possibly isn't where you
            think  it  is  relative to your web root directory). Note that
            for  IIS  you  can  trap this error by ticking the 'check file
            exists'  box  when  setting  up  the  script  mappings  in the
            Internet  Services  Manager.  If  a script file does not exist
            then the server will return a 404 error instead. There is also
            the  additional  benefit  that  IIS will do any authentication
            required  for  you  based  on the NTLanMan permissions on your
            script file.

   Windows: I've followed all the instructions, but still can't get PHP
          and IIS to work together!
          Make  sure any user who needs to run a PHP script has the rights
          to run php.exe! IIS uses an anonymous user which is added at the
          time  IIS is installed. This user needs rights to php.exe. Also,
          any authenticated user will also need rights to execute php.exe.
          And  for  IIS4  you need to tell it that PHP is a script engine.
          Also, you will want to read this faq.

   When running PHP as CGI with IIS, OmniHTTPD or Xitami, I get the
          following error: Security Alert! PHP CGI cannot be accessed
          directly..
          You  must set the cgi.force_redirect directive to 0. It defaults
          to  1  so  be sure the directive isn't commented out (with a ;).
          Like all directives, this is set in php.ini

          Because  the  default  is 1, it's critical that you're 100% sure
          that  the  correct php.ini file is being read. Read this faq for
          details.

   How do I know if my php.ini is being found and read? It seems like it
          isn't as my changes aren't being implemented.
          To  be  sure  your  php.ini is being read by PHP, make a call to
          phpinfo().  Near  the  top,  there  will  be  a  listing  called
          Configuration  File  (php.ini).  This will tell you where PHP is
          looking  for php.ini and whether or not it's being read. If just
          a  directory  PATH  exists,  then  it's  not being read, and you
          should  put  your  php.ini  in  that  directory.  If  php.ini is
          included within the PATH, it is being read.

          If  php.ini  is  being  read and you're running PHP as a module,
          then  be sure to restart your web server after making changes to
          php.ini

          See also php_ini_loaded_file().

   How do I add my PHP directory to the PATH on Windows?
          On Windows NT+ and Windows Server 2000+:

          + Go  to  Control  Panel  and  open  the  System  icon (Start ->
            Settings  -> Control Panel -> System, or just Start -> Control
            Panel -> System for Windows XP/2003+)
          + Go to the Advanced tab
          + Click on the 'Environment Variables' button
          + Look into the 'System Variables' pane
          + Find the Path entry (you may need to scroll to find it)
          + Double click on the Path entry
          + Enter  your  PHP  directory  at  the end, including ';' before
            (e.g. ;C:\php)
          + Press OK

          On Windows 98/Me you need to edit the autoexec.bat file:

          + Open the Notepad (Start -> Run and enter notepad)
          + Open the C:\autoexec.bat file
          + Locate  the line with PATH=C:\WINDOWS;C:\WINDOWS\COMMAND;.....
            and add: ;C:\php to the end of the line
          + Save the file and restart your computer

     Note:  Be  sure  to reboot after following the steps above to ensure
     that the PATH changes are applied.

          The  PHP  manual  used  to promote the copying of files into the
          Windows   system  directory,  this  is  because  this  directory
          (C:\Windows,  C:\WINNT, etc.) is by default in the systems PATH.
          Copying  files  into the Windows system directory has long since
          been deprecated and may cause problems.

   How do I make the php.ini file available to PHP on windows?
          There  are  several ways of doing this. If you are using Apache,
          read  their installation specific instructions (Apache 1, Apache
          2), otherwise you must set the PHPRC environment variable:

          On Windows NT, 2000, XP and 2003:

          + Go  to  Control  Panel  and  open  the  System  icon (Start ->
            Settings  -> Control Panel -> System, or just Start -> Control
            Panel -> System for Windows XP/2003)
          + Go to the Advanced tab
          + Click on the 'Environment Variables' button
          + Look into the 'System variables' pane
          + Click  on 'New' and enter 'PHPRC' as the variable name and the
            directory where php.ini is located as the variable value (e.g.
            C:\php)
          + Press OK and restart your computer

          On Windows 98/Me you need to edit the autoexec.bat file:

          + Open the Notepad (Start -> Run and enter notepad)
          + Open the C:\autoexec.bat file
          + Add  a  new  line  to  the  end  of the file: set PHPRC=C:\php
            (replace  C:\php with the directory where php.ini is located).
            Please note that the path cannot contain spaces. For instance,
            if  you  have installed PHP in C:\Program Files\PHP, you would
            enter C:\PROGRA~1\PHP instead.
          + Save the file and restart your computer

   Is it possible to use Apache content negotiation (MultiViews option)
          with PHP?
          If  links  to  PHP  files  include  extension,  everything works
          perfect.  This  FAQ is only for the case when links to PHP files
          don't  include extension and you want to use content negotiation
          to  choose  PHP  files from URL with no extension. In this case,
          replace the line AddType application/x-httpd-php .php with:

AddHandler php5-script php
AddType text/html php

          This  solution  doesn't  work for Apache 1 as PHP module doesn't
          catch php-script.

   Is PHP limited to process GET and POST request methods only?
          No,  it  is possible to handle any request method, e.g. CONNECT.
          Proper  response  status  can be sent with header(). If only GET
          and POST methods should be handled, it can be achieved with this
          Apache configuration:

<LimitExcept GET POST>
Deny from all
</LimitExcept>