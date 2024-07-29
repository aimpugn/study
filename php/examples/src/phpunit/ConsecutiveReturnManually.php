<?php

class ComponentTest extends PHPUnit_Framework_TestCase
{
    public function setUp()
    {
        $this->Component = new Component();
    }

    public function testSomething()
    {
        $this->Component->MyModel = $this->getMockBuilder('MyModel')
            ->disableOriginalConstructor()
            ->getMock();

        $values = [
            [
                'MyModel' => [
                    'id' => 1000,
                    'status' => 'ready',
                ],
            ],
            [
                'MyModel' => [
                    'id' => 1000,
                    'status' => 'processed',
                ],
            ]
        ];

        $this->Component->MyModel->expects($this->any())
            ->method('myMethod')
            //                                            &$values 로 참조로 전달합니다.
            ->will($this->returnCallback(function () use (&$values) {
                return array_shift($values);
            }));
    }
}
