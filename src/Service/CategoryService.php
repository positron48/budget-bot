<?php

namespace App\Service;

use App\Entity\Category;
use App\Entity\CategoryKeyword;
use App\Entity\User;
use App\Entity\UserCategory;
use App\Repository\CategoryKeywordRepository;
use App\Repository\CategoryRepository;
use App\Repository\UserCategoryRepository;

class CategoryService
{
    public function __construct(
        private readonly CategoryRepository $categoryRepository,
        private readonly UserCategoryRepository $userCategoryRepository,
        private readonly CategoryKeywordRepository $categoryKeywordRepository,
    ) {
    }

    /**
     * @return array<string>
     */
    public function getCategories(bool $isIncome, User $user): array
    {
        $categories = [];
        $type = $isIncome ? 'income' : 'expense';

        // Get user-specific categories
        $userCategories = $this->userCategoryRepository->findByUserAndType($user, $type);
        foreach ($userCategories as $category) {
            $categories[] = $category->getName();
        }

        // Get default categories
        $defaultCategories = $this->categoryRepository->findByType($type);
        foreach ($defaultCategories as $category) {
            $categories[] = $category->getName();
        }

        return array_unique($categories);
    }

    public function detectCategory(string $description, string $type, User $user): ?string
    {
        $isIncome = 'income' === $type;
        $type = $isIncome ? 'income' : 'expense';

        // Check user-specific categories first
        $userCategories = $this->userCategoryRepository->findByUserAndType($user, $type);
        foreach ($userCategories as $category) {
            foreach ($category->getKeywords() as $keyword) {
                if (str_contains(mb_strtolower($description), mb_strtolower($keyword->getKeyword()))) {
                    return $category->getName();
                }
            }
        }

        // Check default categories
        $defaultCategories = $this->categoryRepository->findByType($type);
        foreach ($defaultCategories as $category) {
            foreach ($category->getKeywords() as $keyword) {
                if (str_contains(mb_strtolower($description), mb_strtolower($keyword->getKeyword()))) {
                    return $category->getName();
                }
            }
        }

        return null;
    }

    /**
     * @param array<string> $keywords
     */
    public function addUserCategory(User $user, string $name, bool $isIncome, array $keywords = []): void
    {
        $category = new UserCategory();
        $category->setUser($user)
            ->setName($name)
            ->setType($isIncome ? 'income' : 'expense');

        $this->userCategoryRepository->save($category, true);

        if (!empty($keywords)) {
            $this->addKeywordsToCategory($keywords, $category);
        }
    }

    /**
     * @param array<string> $keywords
     */
    private function addKeywordsToCategory(array $keywords, UserCategory|Category $category): void
    {
        foreach ($keywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword($keyword);

            if ($category instanceof UserCategory) {
                $categoryKeyword->setUserCategory($category);
            } else {
                $categoryKeyword->setCategory($category);
            }

            $this->categoryKeywordRepository->save($categoryKeyword, true);
        }
    }
}
