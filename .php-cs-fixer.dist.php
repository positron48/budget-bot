<?php

$finder = (new PhpCsFixer\Finder())
    ->in(__DIR__)
    ->exclude('var')
    ->exclude('vendor')
;

return (new PhpCsFixer\Config())
    ->setRules([
        '@Symfony' => true,
        '@PSR12' => true,
        'array_syntax' => ['syntax' => 'short'],
        'ordered_imports' => ['sort_algorithm' => 'alpha'],
        'no_unused_imports' => true,
        'declare_strict_types' => true,
        'strict_param' => true,
        'no_superfluous_phpdoc_tags' => true,
        'phpdoc_align' => true,
        'phpdoc_order' => true,
        'phpdoc_separation' => true,
        'phpdoc_types_order' => ['null_adjustment' => 'always_last'],
    ])
    ->setFinder($finder)
;