<?php
# php ./resources/scripts/test_mysql_connection.php --host=docker.for.mac.localhost --port=33060 --username=devdb --password=
checkMysqlConnection(parseArgv($argv));

function parseArgv($argv)
{
    unset($argv[0]);
    $parsedArgv = [];
    println(implode(", ", $argv), "passed argv");
    foreach ($argv as $arg) {
        $arr = explode("=", $arg);
        if(count($arr) == 2 && trim($arr[0]) && trim($arr[1])) {
            $parsedArgv[trim($arr[0])] = trim($arr[1]);
        }
    }

    return $parsedArgv;
}

function checkMysqlConnection($parsedArgv)
{
    if (isValidArgv($parsedArgv)) {
        $db = new mysqli("{$parsedArgv['--host']}:{$parsedArgv['--port']}", $parsedArgv['--username'], $parsedArgv['--password']);
        if($db->connect_error){
            println($db->connect_error, "Error");
        } else {
            $res = $db->query("SELECT 1");
            $row = $res->fetch_assoc();
            $db->close();
            println(count($row) > 0 && (intval($row[1]) == 1) ? "success" : "failed", "result");
        }
    }
}

function isValidArgv($parsedArgv)
{
    $requiredArgvList = [
        '--host',
        '--port',
        '--username',
        '--password',
    ];
    $notPassed = [];
    foreach ($requiredArgvList as $requiredArgv) {
        if (empty($parsedArgv[$requiredArgv])) {
            $notPassed[] = $requiredArgv;
        }
    }
    if (!empty($notPassed)) {
        println(implode(", ", $notPassed) . " should be passed", "Check parsed argv");
        return false;
    }

    return true;
}

function println($msg, $title = '')
{
    if (!is_string($msg)) {
        $msg = print_r($msg, true);
    }
    $line = $msg;
    if (!empty($title)) {
        $line = "$title: $line";
    }
    echo $line . PHP_EOL;
}

exit;
