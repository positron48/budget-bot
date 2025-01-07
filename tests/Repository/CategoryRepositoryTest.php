<?php

namespace App\Tests\Repository;

use App\Entity\Category;
use App\Repository\CategoryRepository;
use App\Tests\Integration\IntegrationTestCase;

class CategoryRepositoryTest extends IntegrationTestCase
{
    private CategoryRepository $repository;

    protected function setUp(): void
    {
        parent::setUp();
        $this->repository = $this->getContainer()->get(CategoryRepository::class);
    }

    public function testFindByType(): void
    {
        // Create test categories
        $defaultExpenseCategory = new Category();
        $defaultExpenseCategory->setName('Test Expense');
        $defaultExpenseCategory->setType('expense');
        $defaultExpenseCategory->setIsDefault(true);
        $this->repository->save($defaultExpenseCategory, true);

        $nonDefaultExpenseCategory = new Category();
        $nonDefaultExpenseCategory->setName('Non Default Expense');
        $nonDefaultExpenseCategory->setType('expense');
        $nonDefaultExpenseCategory->setIsDefault(false);
        $this->repository->save($nonDefaultExpenseCategory, true);

        $defaultIncomeCategory = new Category();
        $defaultIncomeCategory->setName('Test Income');
        $defaultIncomeCategory->setType('income');
        $defaultIncomeCategory->setIsDefault(true);
        $this->repository->save($defaultIncomeCategory, true);

        // Test finding expense categories
        $expenseCategories = $this->repository->findByType('expense');
        $this->assertCount(1, $expenseCategories);
        $this->assertEquals('Test Expense', $expenseCategories[0]->getName());

        // Test finding income categories
        $incomeCategories = $this->repository->findByType('income');
        $this->assertCount(1, $incomeCategories);
        $this->assertEquals('Test Income', $incomeCategories[0]->getName());

        // Test finding non-existent type
        $otherCategories = $this->repository->findByType('other');
        $this->assertEmpty($otherCategories);
    }

    public function testSave(): void
    {
        $category = new Category();
        $category->setName('Test Category');
        $category->setType('expense');
        $category->setIsDefault(true);

        // Test save without flush
        $this->repository->save($category, false);
        $this->entityManager->clear();

        // Entity should not be in database yet
        $found = $this->repository->findOneBy(['name' => 'Test Category']);
        $this->assertNull($found);

        // Test save with flush
        $this->repository->save($category, true);
        $this->entityManager->clear();

        // Entity should be in database now
        $found = $this->repository->findOneBy(['name' => 'Test Category']);
        $this->assertNotNull($found);
        $this->assertEquals('Test Category', $found->getName());
        $this->assertEquals('expense', $found->getType());
        $this->assertTrue($found->isDefault());
    }

    protected function tearDown(): void
    {
        $this->entityManager->createQuery('DELETE FROM App\Entity\Category')->execute();
        parent::tearDown();
    }
}
