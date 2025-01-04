<?php

namespace App\Tests\Service;

use App\Entity\Category;
use App\Entity\CategoryKeyword;
use App\Entity\User;
use App\Entity\UserCategory;
use App\Repository\CategoryKeywordRepository;
use App\Repository\CategoryRepository;
use App\Repository\UserCategoryRepository;
use App\Service\CategoryService;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;

class CategoryServiceTest extends TestCase
{
    private CategoryService $service;
    private MockObject&CategoryRepository $repository;
    private MockObject&UserCategoryRepository $userCategoryRepository;
    private MockObject&CategoryKeywordRepository $categoryKeywordRepository;

    protected function setUp(): void
    {
        $this->repository = $this->createMock(CategoryRepository::class);
        $this->userCategoryRepository = $this->createMock(UserCategoryRepository::class);
        $this->categoryKeywordRepository = $this->createMock(CategoryKeywordRepository::class);
        $this->service = new CategoryService(
            $this->repository,
            $this->userCategoryRepository,
            $this->categoryKeywordRepository
        );
    }

    public function testGetCategories(): void
    {
        $user = new User();
        $userCategories = [
            (new UserCategory())
                ->setUser($user)
                ->setName('Продукты')
                ->setType('expense'),
            (new UserCategory())
                ->setUser($user)
                ->setName('Транспорт')
                ->setType('expense'),
        ];

        $defaultCategories = [
            (new Category())
                ->setName('Развлечения')
                ->setType('expense')
                ->setIsDefault(true),
            (new Category())
                ->setName('Здоровье')
                ->setType('expense')
                ->setIsDefault(true),
        ];

        $this->userCategoryRepository->expects($this->once())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $this->repository->expects($this->once())
            ->method('findByType')
            ->with('expense')
            ->willReturn($defaultCategories);

        $result = $this->service->getCategories(false, $user);

        $this->assertEquals(['Продукты', 'Транспорт', 'Развлечения', 'Здоровье'], $result);
    }

    public function testDetectCategory(): void
    {
        $user = new User();
        $userCategories = [];

        $defaultCategories = [
            $this->createCategoryWithKeywords('Продукты', ['продукты', 'еда'], 'expense'),
            $this->createCategoryWithKeywords('Транспорт', ['такси', 'метро'], 'expense'),
        ];

        $this->userCategoryRepository->expects($this->once())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $this->repository->expects($this->once())
            ->method('findByType')
            ->with('expense')
            ->willReturn($defaultCategories);

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

        $this->userCategoryRepository->expects($this->once())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $this->repository->expects($this->never())
            ->method('findByType');

        $result = $this->service->detectCategory('поездка на такси', 'expense', $user);

        $this->assertEquals('Транспорт', $result);
    }

    public function testDetectCategoryNoMatch(): void
    {
        $user = new User();
        $userCategories = [
            $this->createUserCategoryWithKeywords('Продукты', ['продукты', 'еда'], 'expense', $user),
            $this->createUserCategoryWithKeywords('Транспорт', ['такси', 'метро'], 'expense', $user),
        ];

        $defaultCategories = [
            $this->createCategoryWithKeywords('Развлечения', ['кино', 'театр'], 'expense'),
            $this->createCategoryWithKeywords('Здоровье', ['аптека', 'врач'], 'expense'),
        ];

        $this->userCategoryRepository->expects($this->once())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $this->repository->expects($this->once())
            ->method('findByType')
            ->with('expense')
            ->willReturn($defaultCategories);

        $result = $this->service->detectCategory('какое-то описание', 'expense', $user);

        $this->assertNull($result);
    }

    public function testDetectCategoryWithEmptyKeywords(): void
    {
        $user = new User();
        $userCategories = [
            $this->createUserCategoryWithKeywords('Продукты', [], 'expense', $user),
            $this->createUserCategoryWithKeywords('Транспорт', [], 'expense', $user),
        ];

        $defaultCategories = [
            $this->createCategoryWithKeywords('Развлечения', [], 'expense'),
            $this->createCategoryWithKeywords('Здоровье', [], 'expense'),
        ];

        $this->userCategoryRepository->expects($this->once())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn($userCategories);

        $this->repository->expects($this->once())
            ->method('findByType')
            ->with('expense')
            ->willReturn($defaultCategories);

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

    /**
     * @param array<string> $keywords
     */
    private function createCategoryWithKeywords(string $name, array $keywords, string $type): Category
    {
        $category = new Category();
        $category->setName($name)
            ->setType($type)
            ->setIsDefault(true);

        foreach ($keywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword($keyword)
                ->setCategory($category);
            $category->addKeyword($categoryKeyword);
        }

        return $category;
    }
}
