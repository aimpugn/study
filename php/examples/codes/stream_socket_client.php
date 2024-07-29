<?php

function connect() {
    $fp = stream_socket_client("tcp://127.0.0.1:12345",$errno, $errstr, 30);
    if (is_resource($fp)) {
        fwrite($fp, "GET / HTTP/1.1\r\nHost: www.example.com\r\nAccept: */*\r\n\r\n");
    }
    $buffer = '';
    while (!feof($fp)) {
        $buffer .= fgets($fp, 1024);
    }
    fclose($fp);
}