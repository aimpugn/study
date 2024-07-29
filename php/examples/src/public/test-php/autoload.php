<?php
require_once dirname(dirname(dirname(__DIR__))) . '/vendor/autoload.php';

error_reporting(E_ALL);
ini_set('display_errors', TRUE);
ini_set('display_startup_errors', TRUE);


function println($msg = '', $title = '')
{
    $isWebRequest = php_sapi_name() === 'fpm-fcgi';
	$line = '';
    $msg = is_bool($msg) ? var_export($msg, true) : $msg;
    if($isWebRequest) {
        $msg = is_array($msg) || is_object($msg)? json_encode($msg, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE) : $msg;
        $line .= '<div>';
        if(empty($title)) {
            $title = 'None Title';
        }
        $line .= "<h4>[$title]</h4>";
        $line .= "<pre>$msg</pre>";
        $line .= '</div>';
        print_r($line);
    } else {
        $msg = is_array($msg) || is_object($msg)? print_r($msg, true) : $msg;
        if(!empty($title)) {
            $line .= "[$title] ";
        }
        $line .= $msg;
        print_r($line . PHP_EOL);
    }
}