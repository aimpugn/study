<?php

use PhpCsFixer\Config;
use PhpCsFixer\Finder;

$finder = Finder::create()
    ->in(__DIR__)
    ->name('*.php')
    // ->notName('*.?')
    ->ignoreDotFiles(true)
    ->ignoreVCS(true)
    ->exclude([
        // 의존성
        'vendor',
    ]);

// 옵션참고 URL
// https://github.com/PHP-CS-Fixer/PHP-CS-Fixer
// https://mlocati.github.io/php-cs-fixer-configurator
$config = (new Config())
    ->setRules([
        //
        // @Symfony 룰셋 적용. @Symfony에 @PSR12도 포함되어 있습니다.
        //
        '@Symfony' => true,
        // @PSR12에서 visibility_required 기본값은 ['property', 'method', 'const']이지만
        // class constant의 visibility는 php 7.1 이전에서는 설정 불가하여 덮어씁니다.
        'visibility_required' => ['elements' => ['method', 'property']],

        //
        // @Symfony에 정의된 룰 중 일부를 적용하지 않거나 덮어씁니다.
        //
        'blank_line_before_statement' => false, // 자율에 맡김
        'cast_spaces' => ['space' => 'none'], // 기존 코드에 none인 경우가 더 많았음
        'concat_space' => ['spacing' => 'one'], // 코어 스크럼 투표로 결정
        'empty_loop_body' => false, // 자율에 맡김
        'increment_style' => false, // 자율에 맡김
        'method_argument_space' => ['on_multiline' => 'ensure_fully_multiline'], // @PSR12 적용
        'method_chaining_indentation' => true,
        'no_alias_language_construct_call' => false, // 코드 변화들 원치 않아 비활성화
        'no_binary_string' => false, // 코드 변화를 원치 않아 비활성화
        'no_superfluous_phpdoc_tags' => false, // 설명을 덧붙이기 용이하도록 남김
        'nullable_type_declaration_for_default_null_value' => ['use_nullable_type_declaration' => false], // php 7.1 이상에서만 가능
        'phpdoc_align' => false, // 일부 주석에서 심각하게 여백이 많아지는 부작용이 있어 비활성화
        'phpdoc_annotation_without_dot' => false, // 엄격할 필요가 없어 비활성화
        'phpdoc_no_package' => false, // 기존 주석 손실을 줄이기 위해 비활성화
        'phpdoc_separation' => ['groups' => [], 'skip_unlisted_annotations' => true], // 기존 코드 스타일 유지
        'phpdoc_summary' => false, // 엄격할 필요가 없어 비활성화
        'phpdoc_to_comment' => false, // Psalm과 상충 https://github.com/PHP-CS-Fixer/PHP-CS-Fixer/issues/3611
        'single_line_throw' => false, // 기존 코드 스타일 보존을 위해 비활성화
        'standardize_increment' => false, // 코드 변화들 원치 않아 비활성화
        'yoda_style' => false, // 일반적이지 않고 기존 코드도 요다 스타일 사용하고 있지 않아 불허

        'align_multiline_comment' => true, // 주석을 phpdocs 규격에 맞게 정렬
        'array_indentation' => true, // 배열을 가독성 좋게 배열
        'phpdoc_add_missing_param_annotation' => true, // 누락된 phpdoc이 있으면 자동 생성
        'phpdoc_tag_casing' => true, // phpdoc 태그 대소문자 자동 교정
        'single_line_empty_body' => true, // 함수 본문이 비었으면 한 줄로 작성
    ])
    ->setFinder($finder)
    ->setUsingCache(true);

return $config;
