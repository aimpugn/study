<?php

require_once dirname(__DIR__) . '/autoload.php';

error_reporting(E_ALL);
ini_set('display_errors', true);
ini_set('display_startup_errors', true);

$rn = "\r\n";
$r = "\r";
$n = "\n";

//$ch = curl_init("https://google.com");
////$ch = curl_init("https://rest2.domain-dev.co/users/getToken");
////$ch = curl_init("https://rest6.domain-dev.co/payments/imp_317587167758");
//curl_setopt($ch, CURLOPT_HEADER, true);
//curl_setopt($ch, CURLOPT_HTTPHEADER, [
//	'Authorization: 64f0edebde0978b0ac6e2dc30b0416e536f7c233',
////	'Content-Type: application/json; charset=UTF-8'
//]);
///*curl_setopt($ch, CURLOPT_POST, true);
//curl_setopt($ch, CURLOPT_POSTFIELDS, [
//	'apiKey' => 'apiKeyValue',
//	'apiSecret' => 'apiSecretValue'
//]);*/
//curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
//curl_setopt($ch, CURLOPT_FOLLOWLOCATION, 0);
//curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, 0);
//curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, 0);
//curl_setopt($ch, CURLOPT_TIMEOUT, 10);
//
//$contents = curl_exec($ch);
//curl_close($ch);
//if($contents === false) {
//	print_r('Curl error: ' . curl_error($ch));
//}

$contents = "GET / HTTP/1.1\r\nHost: www.example.com\r\nAccept: */*\r\n\r\n";
print_r($contents . PHP_EOL);

/*print_r("1. " . strpos($contents, "\n") . PHP_EOL);
print_r("2. " . strpos($contents, "\r\n") . PHP_EOL);
$lines1 = explode("\n", $contents);
print_r("3.1. " . strpos($lines1[0], "\r") . PHP_EOL);
print_r("3.2. " . strpos($lines1[0], "\n") . PHP_EOL);
print_r("3.3. " . strpos(trim($lines1[0]), "\r") . PHP_EOL);

$lines2 = array_map('trim', explode("\n", $contents));
print_r($lines2);
print_r(gettype($lines2[11]) . PHP_EOL);*/

function parse($rawText)
{
    $result = [];
    $rawText = trim($rawText);
    $headerAndBody = splitHeaderAndBody($rawText);

    if (!empty($headerAndBody)) {
        $result['headers'] = getHeaders($headerAndBody[0]);
        $result['body'] = !empty($headerAndBody[1]) ? trim($headerAndBody[1]) : '';
    }

    return $result;
}

function splitHeaderAndBody($rawText)
{
    $splitText = explode("\r\n\r\n", $rawText);
    if(count($splitText) > 0) {
        return $splitText;
    }

    $splitText = explode("\n\n", $rawText);
    if(count($splitText) > 0) {
        return $splitText;
    }

    return [];
}

function getHeaders($headerText)
{
    $headers = [];
    $headerLines = array_map('trim', explode("\n", $headerText));
    $lineCnt = count($headerLines);
    $cnt = 0;
    if ($lineCnt > 0) {
        $firstLine = $headerLines[$cnt++];
        if (strpos($firstLine, 'HTTP') === false) {
            return $headers;
        }

        $headers['http'] = $firstLine;
        while ($cnt < $lineCnt && ($line = $headerLines[$cnt++]) && ! empty($line)) {
            $splitLine = array_map('trim', explode(":", $line, 2));
            $headers[$splitLine[0]] = $splitLine[1];
        }
    }

    return $headers;
}

$parsedHttp = parse($contents);
print_r($parsedHttp);

print_r(parse("HTTP/1.1 200 OK\r\n\r\nsuccess"));

println('test');
//println(`test`);

$tmp1 = [];
$tmp2['message'] = $tmp1['yyyyy'];
println(var_export(is_string($tmp2['message']) && $tmp2['message'] == 'success', true));

println(var_export(null === 'success'));

$str1 = "teststring";
println('$str1[0]: ' . $str1[0]);
println('$str1[1]: ' . $str1[1]);

println(ucfirst('test'));

$strNumber1 = "123456";
$null = null;
$tmp3 = [];
println(var_export(is_string($null), true), "is_string null");
println(var_export(is_string($tmp3['undefined_key']), true), "is_string tmp3['undefined_key']");
println(var_export(is_string($strNumber1), true), "is_string $strNumber1");
println(var_export(is_string($undefined), true), "is_string undefined");
