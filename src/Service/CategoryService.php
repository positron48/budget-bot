<?php

namespace App\Service;

use App\Entity\Category;
use App\Entity\CategoryKeyword;
use App\Entity\User;
use App\Entity\UserCategory;
use App\Repository\CategoryKeywordRepository;
use App\Repository\UserCategoryRepository;
use Doctrine\ORM\EntityManagerInterface;

class CategoryService
{
    public function __construct(
        private readonly UserCategoryRepository $userCategoryRepository,
        private readonly CategoryKeywordRepository $categoryKeywordRepository,
        private readonly EntityManagerInterface $entityManager,
    ) {
    }

    /**
     * @return array<string>
     */
    public function getCategories(bool $isIncome, User $user): array
    {
        $type = $isIncome ? 'income' : 'expense';

        // Get user-specific categories
        $userCategories = $this->userCategoryRepository->findByUserAndType($user, $type);
        $userCategoryNames = array_map(
            static fn (UserCategory $category): string => $category->getName() ?? '',
            $userCategories
        );

        sort($userCategoryNames);

        return array_values(array_unique($userCategoryNames));
    }

    public function detectCategory(string $description, string $type, User $user): ?string
    {
        $isIncome = 'income' === $type;
        $type = $isIncome ? 'income' : 'expense';
        $description = mb_strtolower($description);

        // First try exact match with full description
        $category = $this->findCategoryByKeyword($description, $type, $user);
        if ($category) {
            return $category;
        }

        // Then try matching individual words
        $words = preg_split('/\s+/', $description);
        if (!is_array($words)) {
            return null;
        }

        foreach ($words as $word) {
            $category = $this->findCategoryByKeyword($word, $type, $user);
            if ($category) {
                return $category;
            }
        }

        return null;
    }

    private function findCategoryByKeyword(string $keyword, string $type, User $user): ?string
    {
        // Check user-specific categories first
        $userCategories = $this->userCategoryRepository->findByUserAndType($user, $type);
        foreach ($userCategories as $category) {
            foreach ($category->getKeywords() as $categoryKeyword) {
                if (mb_strtolower($keyword) === mb_strtolower($categoryKeyword->getKeyword() ?? '')) {
                    return $category->getName();
                }
            }
        }

        return null;
    }

    public function addKeywordToCategory(string $keyword, string $categoryName, string $type, User $user): void
    {
        // Try to find user category first
        $userCategory = $this->userCategoryRepository->findOneBy([
            'user' => $user,
            'name' => $categoryName,
            'type' => $type,
        ]);

        if ($userCategory) {
            $this->addKeywordsToCategory([$keyword], $userCategory);

            return;
        }

        // Create a new user category
        $userCategory = new UserCategory();
        $userCategory->setUser($user)
            ->setName($categoryName)
            ->setType($type);

        $this->userCategoryRepository->save($userCategory, true);
        $this->addKeywordsToCategory([$keyword], $userCategory);
    }

    /**
     * @param array<string> $keywords
     */
    public function addUserCategory(User $user, string $name, bool $isIncome, array $keywords = []): void
    {
        // Check if category already exists
        $existingCategory = $this->userCategoryRepository->findOneBy([
            'user' => $user,
            'name' => $name,
            'type' => $isIncome ? 'income' : 'expense',
        ]);

        if ($existingCategory) {
            return;
        }

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

    public function removeUserCategory(User $user, string $name, bool $isIncome): void
    {
        $category = $this->userCategoryRepository->findOneBy([
            'user' => $user,
            'name' => $name,
            'type' => $isIncome ? 'income' : 'expense',
        ]);

        if ($category) {
            $this->userCategoryRepository->remove($category, true);
        }
    }

    public function clearUserCategories(User $user): void
    {
        $expenseCategories = $this->userCategoryRepository->findByUserAndType($user, 'expense');
        $incomeCategories = $this->userCategoryRepository->findByUserAndType($user, 'income');

        foreach ([...$expenseCategories, ...$incomeCategories] as $category) {
            $this->userCategoryRepository->remove($category, false);
        }

        $this->entityManager->flush();
    }
}
