<?php

class ProtectedClass
{
    protected $protectedProperty = 'protected';
}

class ClassNotExtendsProtectedClass
{
    public function publicFunction()
    {
        $protected = new ProtectedClass();
        print_r([
            __METHOD__,
            // Fatal error: Uncaught Error: Cannot access protected property ProtectedClass::$protectedProperty
            $protected,
        ]);
        // Output:
        // Array
        // (
        //     [0] => ClassNotExtendsProtectedClass::publicFunction
        //     [1] => ProtectedClass Object
        //         (
        //             [protectedProperty:protected] => protected
        //         )
        //
        // )
    }
}

(new ClassNotExtendsProtectedClass())->publicFunction();


// `extends ProtectedClass` 없는 class PublicClass 경우 `$protected->protectedProperty`에서 Fatal 에러 발생
class ClassExtendsProtectedClass extends ProtectedClass
{
    public function publicFunction()
    {
        $protected = new ProtectedClass();
        print_r([
            __METHOD__,
            $protected->protectedProperty,
            $this->protectedProperty,
        ]);
        // Output:
        // Array
        //(
        //    [0] => protected
        //    [1] => protected
        //)
    }
}

(new ClassExtendsProtectedClass())->publicFunction();
