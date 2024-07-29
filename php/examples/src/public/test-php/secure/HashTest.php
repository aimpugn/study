<?php

require_once dirname(__DIR__) . '/autoload.php';

/*$tid = "someMID01012211151748493417";
$authToken = "NICETOKNA31E01D823DA3F10013A480C69BC86CF";
$mid = "someMID";
$amt = 100;
$ediDate = "20221117201312";
$merchantKey = 'dbW/d3tM72T5TrLeHAm1ph76MOmy1lFZSDH6KCTHlZrMcLHr3AJi9m04AOnv7Fh86ladh52h+wky+LMmP1PbLw==';*/

$tid = "someMID01012212022148153756";
$authToken = "NICETOKN68217145E4B7CB9FDA80570B96DD88EF";
$mid = "someMID";
$amt = 100;
$ediDate = "20221202214819";
$merchantKey = 'dbW/d3tM72T5TrLeHAm1ph76MOmy1lFZSDH6KCTHlZrMcLHr3AJi9m04AOnv7Fh86ladh52h+wky+LMmP1PbLw==';
$signData = bin2hex(hash('sha256', $authToken . $mid . $amt . $ediDate . $merchantKey, true));
println($signData, '$signData1');

$authToken = "NICETOKNA475D00A0A753EED265F7E579E5149B3";
$tid = "someMID01012302081855394959";
$mid = "someMID";
$amt = 100;
$ediDate = "20230208185519";
$merchantKey = 'dbW/d3tM72T5TrLeHAm1ph76MOmy1lFZSDH6KCTHlZrMcLHr3AJi9m04AOnv7Fh86ladh52h+wky+LMmP1PbLw==';
$signData = bin2hex(hash('sha256', $authToken . $mid . $amt . $ediDate . $merchantKey, true));
println($signData, '$signData2');
