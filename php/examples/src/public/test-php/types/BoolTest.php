<?php
require_once dirname(__DIR__) . '/autoload.php';
error_reporting(E_ALL);
ini_set('display_errors', TRUE);
ini_set('display_startup_errors', TRUE);

$arr1 = [

];

$escrow_confirmed = filter_var($arr1['escrow_confirmed'], FILTER_VALIDATE_BOOLEAN);

var_dump($escrow_confirmed);

function emptyx($v) {
    return !isset($v) || $v != true;
}

unset($undef);
$zero = 0;
$one = 1;
$true = true;
$false = false;
$null = null;
$arr = [];
println(var_export(isset($null), true), 'isset($null)');
println(var_export(isset($undef), true), 'isset($undef)');
println(var_export(isset($zero), true), 'isset($zero)');
println(var_export(isset($one), true), 'isset($one)');
println(var_export(isset($true), true), 'isset($true)');
println(var_export(isset($false), true), 'isset($false)');
println(var_export(isset($arr), true), 'isset($arr)');
println(var_export(isset($arr['not_exists']), true), 'isset($arr[not_exists])');
println();
println(var_export(!isset($null), true), '!isset($null)');
println(var_export(!isset($undef), true), '!isset($undef)');
println(var_export(!isset($zero), true), '!isset($zero)');
println(var_export(!isset($one), true), '!isset($one)');
println(var_export(!isset($true), true), '!isset($true)');
println(var_export(!isset($false), true), '!isset($false)');
println(var_export(!isset($arr), true), '!isset($arr)');
println(var_export(!isset($arr['not_exists']), true), '!isset($arr[not_exists])');
println();
println(var_export(empty($null), true), 'empty($null)');
println(var_export(empty($undef), true), 'empty($undef)');
println(var_export(empty($zero), true), 'empty($zero)');
println(var_export(empty($one), true), 'empty($one)');
println(var_export(empty($true), true), 'empty($true)');
println(var_export(empty($false), true), 'empty($false)');
println(var_export(empty($arr), true), 'empty($arr)');
println(var_export(empty($arr['not_exists']), true), 'empty($arr[not_exists])');
println();
println(var_export(!empty($null), true), '!empty(null)');
println(var_export(!empty($undef), true), '!empty(undefined)');
println(var_export(!empty($zero), true), '!empty(0)');
println(var_export(!empty($one), true), '!empty(1)');
println(var_export(!empty($true), true), '!empty(true)');
println(var_export(!empty($false), true), '!empty(false)');
println(var_export(!empty($arr), true), '!empty([])');
println(var_export(!empty($arr['not_exist_key']), true), '!empty(arr[not_exist_key])');
println();
println(var_export(isset($null) && $null == true, true), 'isset && true (null)');
println(var_export(isset($undef) && $undef == true, true), 'isset && true (undefined)');
println(var_export(isset($zero) && $zero == true, true), 'isset && true (0)');
println(var_export(isset($one) && $one == true, true), 'isset && true (1)');
println(var_export(isset($true) && $true == true, true), 'isset && true (true)');
println(var_export(isset($true) && $false == true, true), 'isset && true (false)');
println(var_export(isset($arr) && $arr == true, true), 'isset && true ([])');
println(var_export(isset($arr['not_exist_key']) && $arr['not_exist_key'] == true, true), 'isset && true arr[not_exist_key]');
println();
$emptyString = '';
println(var_export(empty(trim($emptyString)), true), 'empty(trim($emptyString))');
println(var_export(empty(trim($true)), true), 'empty(trim($true))');
println(var_export(empty(trim($undef)), true), 'empty(trim($undef))');
println(var_export(empty(trim($null)), true), 'empty(trim($null))');
println();
println(trim($true), 'trim($true)');
println(gettype(trim($true)), 'gettype(trim($true))');
println();
$arr2 = [
    'true' => true
];
println(var_export(!empty($arr2['true']), true), '!empty($arr2[true])');

$null = NULL;
$false = FALSE;
$true = TRUE;

// these operations:
println(isset($var), 'isset($var)');
println(empty($var), 'empty($var)');
println(array_key_exists('var', get_defined_vars()), "array_key_exists('var', get_defined_vars())");// "defined" below
println(var_export(!empty("0"), true), '! empty("0")');
println(var_export(!empty(0), true), '! empty(0)');
println(var_export(!empty(false), true), '! empty(false)');
println(var_export(!empty("1"), true), '! empty("1")');
println(var_export(!empty("1"), true), '! empty("1")');

