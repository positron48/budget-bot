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
    private CategoryRepository $categoryRepository;
    private UserCategoryRepository $userCategoryRepository;
    private CategoryKeywordRepository $keywordRepository;

    public function __construct(
        CategoryRepository $categoryRepository,
        UserCategoryRepository $userCategoryRepository,
        CategoryKeywordRepository $keywordRepository
    ) {
        $this->categoryRepository = $categoryRepository;
        $this->userCategoryRepository = $userCategoryRepository;
        $this->keywordRepository = $keywordRepository;
    }

    public function detectCategory(string $description, bool $isIncome = false, ?User $user = null): ?string
    {
        $type = $isIncome ? 'income' : 'expense';
        $keywords = $this->keywordRepository->findMatchingKeywords($description, $type, $user);

        if (empty($keywords)) {
            return null;
        }

        // Return the first matching category
        $keyword = $keywords[0];
        if ($keyword->getUserCategory()) {
            return $keyword->getUserCategory()->getName();
        }
        return $keyword->getCategory()->getName();
    }

    public function getCategories(bool $isIncome = false, ?User $user = null): array
    {
        $type = $isIncome ? 'income' : 'expense';
        $categories = [];

        // Get default categories
        foreach ($this->categoryRepository->findByType($type) as $category) {
            $categories[] = $category->getName();
        }

        // Get user categories if user is provided
        if ($user) {
            foreach ($this->userCategoryRepository->findByUserAndType($user, $type) as $category) {
                $categories[] = $category->getName();
            }
        }

        return array_unique($categories);
    }

    public function isValidCategory(string $categoryName, bool $isIncome = false, ?User $user = null): bool
    {
        return in_array($categoryName, $this->getCategories($isIncome, $user));
    }

    public function addUserCategory(User $user, string $name, string $type, array $keywords = []): UserCategory
    {
        $category = new UserCategory();
        $category->setUser($user)
                ->setName($name)
                ->setType($type);

        foreach ($keywords as $keywordText) {
            $keyword = new CategoryKeyword();
            $keyword->setKeyword($keywordText)
                   ->setUserCategory($category);
            $this->keywordRepository->save($keyword);
        }

        $this->userCategoryRepository->save($category, true);
        return $category;
    }

    public function addKeywordToCategory(string $keyword, string $categoryName, bool $isIncome = false, ?User $user = null): void
    {
        $type = $isIncome ? 'income' : 'expense';
        $categoryKeyword = new CategoryKeyword();
        $categoryKeyword->setKeyword($keyword);

        if ($user) {
            $userCategories = $this->userCategoryRepository->findByUserAndType($user, $type);
            foreach ($userCategories as $category) {
                if ($category->getName() === $categoryName) {
                    $categoryKeyword->setUserCategory($category);
                    break;
                }
            }
        }

        if (!$categoryKeyword->getUserCategory()) {
            $defaultCategories = $this->categoryRepository->findByType($type);
            foreach ($defaultCategories as $category) {
                if ($category->getName() === $categoryName) {
                    $categoryKeyword->setCategory($category);
                    break;
                }
            }
        }

        if ($categoryKeyword->getCategory() || $categoryKeyword->getUserCategory()) {
            $this->keywordRepository->save($categoryKeyword, true);
        }
    }
} 