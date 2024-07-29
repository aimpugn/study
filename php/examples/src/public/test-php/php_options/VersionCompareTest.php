<?php
require_once dirname(__DIR__) . '/autoload.php';

println(version_compare('1.1.7', '1.1.8'), "version_compare('1.1.7', '1.1.8')");
println(version_compare('1.1.8', '1.1.8'), "version_compare('1.1.8', '1.1.8')");
println(version_compare('1.2.0', '1.1.8'), "version_compare('1.2.0', '1.1.8')");