<?php

class StreamSocketServer {
    private static $client;
    private static $socket;
    private static $pid;

    public static function run()
    {
        $cwd = getcwd();
        $file = __FILE__;
        $serverLogFile = "$cwd/app/tmp/logs/temp_server.txt";
        $cmd =<<<EOF
php $file --is_cli=true &>$serverLogFile &

EOF;

        shell_exec("$cmd");
        self::$pid = shell_exec("pgrep -f $file");
        self::printLog("server cmd: $cmd");
        self::printLog("server pid: " . self::$pid);
        sleep(1);
    }

    public static function listen()
    {
        $socket = stream_socket_server(
            "tcp://127.0.0.1:12345",
            $errno,
            $errstr
        );
        if (!$socket) {
            self::printServerLog("server side error? $errno, $errstr");
        } else {
            self::printServerLog("server side socket created");
            while ($conn = stream_socket_accept($socket, 30)) {
                self::printServerLog("client: " . stream_socket_get_name($conn, true));
                $recv = stream_socket_recvfrom($conn, 4096, 0, $peer);
                self::printServerLog("server side: recv $recv");
                fwrite($conn, 'The local time is ' . date('n/j/Y g:i a'));
                fclose($conn);
            }
            fclose($socket);
        }
    }

    public static function read()
    {
        $input = socket_read(self::$client, 1024);
        return $input;
    }

    public static function close()
    {
        socket_close(self::$socket);
    }

    public static function stopRunningProcess()
    {
        $pid = self::$pid;
        if ($pid && $pid > 1) {
            shell_exec("kill -15 $pid");
            self::$pid = null;
        }
    }

    public static function printLog($msg)
    {
        $cwd = getcwd();
        $pointer = fopen("$cwd/app/tmp/logs/temp.txt", 'a+');
        if (is_resource($pointer)) {
            fwrite($pointer, $msg . PHP_EOL);
            fclose($pointer);
        }
    }

    public static function printServerLog($msg)
    {
        $cwd = getcwd();
        $pointer = fopen("$cwd/app/tmp/logs/temp_server.txt", 'a+');
        if (is_resource($pointer)) {
            fwrite($pointer, $msg . PHP_EOL);
            fclose($pointer);
        }
    }
}