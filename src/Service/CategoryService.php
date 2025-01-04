<?php

namespace App\Service;

use Symfony\Component\Yaml\Yaml;

class CategoryService
{
    private array $categories;

    public function __construct(string $projectDir)
    {
        $this->categories = Yaml::parseFile($projectDir . '/config/categories.yaml')['categories'];
    }

    public function detectCategory(string $description, bool $isIncome = false): ?string
    {
        $type = $isIncome ? 'income' : 'expenses';
        $description = mb_strtolower($description);

        foreach ($this->categories[$type] as $category => $keywords) {
            foreach ($keywords as $keyword) {
                if (mb_strpos($description, mb_strtolower($keyword)) !== false) {
                    return $category;
                }
            }
        }

        return null;
    }

    public function getCategories(bool $isIncome = false): array
    {
        $type = $isIncome ? 'income' : 'expenses';
        return array_keys($this->categories[$type]);
    }

    public function isValidCategory(string $category, bool $isIncome = false): bool
    {
        $type = $isIncome ? 'income' : 'expenses';
        return isset($this->categories[$type][$category]);
    }
} 