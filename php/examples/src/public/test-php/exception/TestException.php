<?php
require_once dirname(__DIR__) . '/autoload.php';

function boolToString($bool) {
	if (filter_var($bool, FILTER_VALIDATE_BOOLEAN, ['options' => ['default' => false]])) {
		return 'true';
	}
	return 'false';
}

function func1() {
	println(__FUNCTION__ . ' is called');
	throw new Exception('이 다음 호출 되나?');
	println(__FUNCTION__ . ' 익셉션 다음' . PHP_EOL);
}

function func2() {
	println(__FUNCTION__ . ' 1. do something');
	
	try {
		func1();
	} catch (Exception $exception) {
		println(__FUNCTION__ . ' when catch, exception message: ' . $exception->getMessage());
	}
	
	print_r(__FUNCTION__ . 'after exception' . PHP_EOL);
}

function func3($error = false) {
	if ($error) {
		throw new Exception('에러 던지기');
	}
	
	return true;
}

function func4($innerTry, $success, $error, $throwAgain = false) {
	if ($innerTry) {
		println(__FUNCTION__ . ' do try/catch');
		try {
			if (func3($error)) {
				if ($success) {
					return '성공!';
				}
			}
		} catch (Exception $e) {
			$message = __FUNCTION__ . ' caught message is "' . $e->getMessage() . '"';
			if ($throwAgain) {
				throw new Exception('throw new Exception From ' . $message);
			}
			return 'just return ' . $message;
		}
	} else {
		println(__FUNCTION__ . ' do NOT try/catch');
		if (func3($error)) {
			if ($success) {
				return '성공!';
			}
		}
	}
	
	return '실패!';
}

function func5($innerTry, $success, $error, $throwAgain) {
	return func4($innerTry, $success, $error, $throwAgain);
}

function func6($isTry, $innerTry, $success, $error, $throwAgain) {
	println('## CASE> ' .
		'$isTry: ' . boolToString($isTry) .
		', $innerTry: ' . boolToString($innerTry) .
		', $success: ' . boolToString($success) .
		', $error: ' . boolToString($error) .
		', $throwAgain: ' . boolToString($throwAgain)
	);
	if ($isTry) {
		try {
			println(__FUNCTION__ . ' in try: ' . func5($innerTry, $success, $error, $throwAgain));
		} catch (Exception $e) {
			println(__FUNCTION__ . ' in catch: ' . $e->getMessage()); // 하위 함수의 throw가 도달하는 곳
		}
	} else {
		// try catch 안한 exception 발생 시
		// Fatal error: Uncaught exception 'Exception' with message '2nd> func4: 에러 던지기'
		println(__FUNCTION__ . ' when NOT try: ' . func5($innerTry, $success, $error, $throwAgain));
	}
	println();
}
//func6(true, true, true, true, true);
//func6(true, true, true, true, false);
//func6(true, false, true, true, false);



function A() {
    try {
        B();
    } catch (Exception $e) {
        print("Error");
    }
}

function B() {
    C();
}

function C() {
    throw new Exception();
}

A();