<?php

namespace App\Tests\Service;

use App\Entity\CategoryKeyword;
use App\Entity\User;
use App\Entity\UserCategory;
use App\Repository\CategoryKeywordRepository;
use App\Repository\UserCategoryRepository;
use App\Service\CategoryService;
use Doctrine\ORM\EntityManagerInterface;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;

class CategoryServiceTest extends TestCase
{
    private CategoryService $service;
    private MockObject&UserCategoryRepository $userCategoryRepository;
    private MockObject&CategoryKeywordRepository $categoryKeywordRepository;
    private MockObject&EntityManagerInterface $entityManager;

    protected function setUp(): void
    {
        $this->userCategoryRepository = $this->createMock(UserCategoryRepository::class);
        $this->categoryKeywordRepository = $this->createMock(CategoryKeywordRepository::class);
        $this->entityManager = $this->createMock(EntityManagerInterface::class);
        $this->service = new CategoryService(
            $this->userCategoryRepository,
            $this->categoryKeywordRepository,
            $this->entityManager,
        );
    }

    public function testGetCategories(): void
    {
        $user = new User();
        $userCategories = [
            $this->createUserCategoryWithKeywords('Продукты', [], 'expense', $user),
            $this->createUserCategoryWithKeywords('Транспорт', [], 'expense', $user),
            $this->createUserCategoryWithKeywords('Развлечения', [], 'expense', $user),
            $this->createUserCategoryWithKeywords('Здоровье', [], 'expense', $user),
        ];

        $this->userCategoryRepository->expects($this->once())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $result = $this->service->getCategories(false, $user);

        $this->assertEquals(['Здоровье', 'Продукты', 'Развлечения', 'Транспорт'], $result);
    }

    public function testDetectCategory(): void
    {
        $user = new User();
        $userCategories = [
            $this->createUserCategoryWithKeywords('Продукты', ['продукты', 'еда'], 'expense', $user),
            $this->createUserCategoryWithKeywords('Транспорт', ['такси', 'метро'], 'expense', $user),
        ];

        $this->userCategoryRepository->expects($this->atLeastOnce())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $result = $this->service->detectCategory('поездка на такси', 'expense', $user);

        $this->assertEquals('Транспорт', $result);
    }

    public function testDetectCategoryInUserCategories(): void
    {
        $user = new User();
        $userCategories = [
            $this->createUserCategoryWithKeywords('Продукты', ['продукты', 'еда'], 'expense', $user),
            $this->createUserCategoryWithKeywords('Транспорт', ['такси', 'метро'], 'expense', $user),
        ];

        $this->userCategoryRepository->expects($this->atLeastOnce())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $result = $this->service->detectCategory('поездка на такси', 'expense', $user);

        $this->assertEquals('Транспорт', $result);
    }

    public function testDetectCategoryNoMatch(): void
    {
        $user = new User();
        $userCategories = [
            $this->createUserCategoryWithKeywords('Продукты', ['продукты', 'еда'], 'expense', $user),
            $this->createUserCategoryWithKeywords('Транспорт', ['такси', 'метро'], 'expense', $user),
            $this->createUserCategoryWithKeywords('Развлечения', ['кино', 'театр'], 'expense', $user),
            $this->createUserCategoryWithKeywords('Здоровье', ['аптека', 'врач'], 'expense', $user),
        ];

        $this->userCategoryRepository->expects($this->atLeastOnce())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $result = $this->service->detectCategory('какое-то описание', 'expense', $user);

        $this->assertNull($result);
    }

    public function testDetectCategoryWithEmptyKeywords(): void
    {
        $user = new User();
        $userCategories = [
            $this->createUserCategoryWithKeywords('Продукты', [], 'expense', $user),
            $this->createUserCategoryWithKeywords('Транспорт', [], 'expense', $user),
            $this->createUserCategoryWithKeywords('Развлечения', [], 'expense', $user),
            $this->createUserCategoryWithKeywords('Здоровье', [], 'expense', $user),
        ];

        $this->userCategoryRepository->expects($this->atLeastOnce())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $result = $this->service->detectCategory('описание', 'expense', $user);

        $this->assertNull($result);
    }

    /**
     * @param array<string> $keywords
     */
    private function createUserCategoryWithKeywords(string $name, array $keywords, string $type, User $user): UserCategory
    {
        $category = new UserCategory();
        $category->setName($name)
            ->setType($type)
            ->setUser($user);

        foreach ($keywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword($keyword)
                ->setUserCategory($category);
            $category->addKeyword($categoryKeyword);
        }

        return $category;
    }

    public function testClearUserCategories(): void
    {
        $user = new User();
        $expenseCategories = [
            $this->createUserCategoryWithKeywords('Продукты', [], 'expense', $user),
            $this->createUserCategoryWithKeywords('Транспорт', [], 'expense', $user),
        ];
        $incomeCategories = [
            $this->createUserCategoryWithKeywords('Зарплата', [], 'income', $user),
            $this->createUserCategoryWithKeywords('Фриланс', [], 'income', $user),
        ];

        $this->userCategoryRepository->expects($this->exactly(2))
            ->method('findByUserAndType')
            ->willReturnCallback(function ($actualUser, $type) use ($user, $expenseCategories, $incomeCategories) {
                $this->assertSame($user, $actualUser);

                return 'expense' === $type ? $expenseCategories : $incomeCategories;
            });

        $removedCategories = [];
        $this->userCategoryRepository->expects($this->exactly(4))
            ->method('remove')
            ->willReturnCallback(function ($category, $flush) use (&$removedCategories) {
                $this->assertFalse($flush);
                $removedCategories[] = $category;
            });

        $this->entityManager->expects($this->once())
            ->method('flush');

        $this->service->clearUserCategories($user);

        // Verify that all categories were removed
        $this->assertCount(4, $removedCategories);
        $this->assertContainsEquals($expenseCategories[0], $removedCategories);
        $this->assertContainsEquals($expenseCategories[1], $removedCategories);
        $this->assertContainsEquals($incomeCategories[0], $removedCategories);
        $this->assertContainsEquals($incomeCategories[1], $removedCategories);
    }
}
