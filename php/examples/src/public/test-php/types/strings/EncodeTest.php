<?php
require_once dirname(__DIR__) . '/autoload.php';

$tmp1 = '（주）프월';
$converted1 = iconv('utf-8', 'euc-kr', $tmp1);
println($converted1, 'utf-8 -> euc-kr');
$converted2 = iconv('EUC-KR', 'UTF-8', $converted1);
println($converted2, 'utf-8 -> euc-kr -> utf-8');
$converted3 = mb_convert_encoding($converted1, 'UTF-8', 'EUC-KR');
println($converted3, 'mb_convert_encoding to utf-8');

$converted4 = mb_convert_encoding($tmp1, 'EUC-KR');
println($converted4, 'mb_convert_encoding to euc-kr');
$converted5 = mb_convert_encoding($converted4, 'UTF-8', 'EUC-KR');
println($converted5, 'mb_convert_encoding to euc-kr -> utf-8');

println(mb_convert_kana($tmp1, 'a'), 'mb_convert_kana rnaskhc');

println(mb_strwidth($tmp1), 'mb_strwidth');
println(mb_strlen($tmp1), 'mb_strlen');
println(strlen($tmp1), 'strlen');


