<?php
require_once dirname(__DIR__) . '/autoload.php';

$filename = __DIR__ . '/test';
//clearstatcache();
# 2분 정도 캐시 된다
println($filename);
println('=========================== before ============================');
file_get_contents($filename);
println(time()); // 1667296106
println(fileatime($filename), 'fileatime > file_get_contents');
println(filemtime($filename), 'filemtime > file_get_contents');
println(stat($filename));
println(realpath_cache_size(), 'before clearstatcache');
println(realpath_cache_get(), 'before clearstatcache'); # expires => 1667296226
clearstatcache(true, $filename);
println('=========================== after ============================');
println(stat($filename));
println(realpath_cache_size(), 'after');
println(realpath_cache_get(), 'after');
println(fileatime($filename), 'fileatime > after file_get_contents');
println(filemtime($filename), 'filemtime > after file_get_contents');

exit;