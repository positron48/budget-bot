<?php

namespace App\Tests\Repository;

use App\Entity\Category;
use App\Entity\CategoryKeyword;
use App\Entity\User;
use App\Entity\UserCategory;
use App\Repository\CategoryKeywordRepository;
use App\Tests\Integration\IntegrationTestCase;

class CategoryKeywordRepositoryTest extends IntegrationTestCase
{
    private CategoryKeywordRepository $repository;

    protected function setUp(): void
    {
        parent::setUp();
        $this->repository = $this->getContainer()->get(CategoryKeywordRepository::class);
    }

    public function testFindMatchingKeywords(): void
    {
        $timestamp = time();

        // Create test data
        $defaultCategory = new Category();
        $defaultCategory->setName('Food Category '.$timestamp);
        $defaultCategory->setType('expense');
        $defaultCategory->setIsDefault(true);
        $this->entityManager->persist($defaultCategory);
        $this->entityManager->flush();

        $defaultKeyword = new CategoryKeyword();
        $defaultKeyword->setKeyword('продукты');
        $defaultKeyword->setCategory($defaultCategory);
        $this->entityManager->persist($defaultKeyword);
        $this->entityManager->flush();

        $user = new User();
        $user->setTelegramId(123456);
        $this->entityManager->persist($user);
        $this->entityManager->flush();

        $userCategory = new UserCategory();
        $userCategory->setName('My Food Category '.$timestamp);
        $userCategory->setType('expense');
        $userCategory->setUser($user);
        $this->entityManager->persist($userCategory);
        $this->entityManager->flush();

        $userKeyword = new CategoryKeyword();
        $userKeyword->setKeyword('еда');
        $userKeyword->setUserCategory($userCategory);
        $this->entityManager->persist($userKeyword);
        $this->entityManager->flush();

        // Test finding keywords without user (only default categories)
        $defaultMatches = $this->repository->findMatchingKeywords('прод', 'expense');
        $this->assertCount(1, $defaultMatches);
        $this->assertEquals('продукты', $defaultMatches[0]->getKeyword());

        // Test finding keywords with user (both default and user categories)
        $userMatches = $this->repository->findMatchingKeywords('еда', 'expense', $user);
        $this->assertCount(1, $userMatches);
        $this->assertEquals('еда', $userMatches[0]->getKeyword());

        // Test finding keywords with wrong type
        $wrongTypeMatches = $this->repository->findMatchingKeywords('прод', 'income');
        $this->assertEmpty($wrongTypeMatches);

        // Test finding keywords with non-matching text
        $nonMatchingMatches = $this->repository->findMatchingKeywords('xyz', 'expense');
        $this->assertEmpty($nonMatchingMatches);

        // Test case insensitive matching
        $caseInsensitiveMatches = $this->repository->findMatchingKeywords('ПРОД', 'expense');
        $this->assertCount(1, $caseInsensitiveMatches);
        $this->assertEquals('продукты', $caseInsensitiveMatches[0]->getKeyword());
    }

    public function testSave(): void
    {
        $timestamp = time();

        // Create and persist category first
        $category = new Category();
        $category->setName('Test Save Category '.$timestamp);
        $category->setType('expense');
        $category->setIsDefault(true);
        $this->entityManager->persist($category);
        $this->entityManager->flush();
        $categoryId = $category->getId();

        // Create keyword and link to the persisted category
        $keyword = new CategoryKeyword();
        $keyword->setKeyword('test');
        $keyword->setCategory($category);

        // Test save without flush
        $this->entityManager->persist($keyword);
        $this->entityManager->clear();

        // Entity should not be in database yet
        $found = $this->repository->findOneBy(['keyword' => 'test']);
        $this->assertNull($found);

        // Fetch category again after clear
        $category = $this->entityManager->find(Category::class, $categoryId);
        $keyword->setCategory($category);

        // Test save with flush
        $this->entityManager->persist($keyword);
        $this->entityManager->flush();
        $this->entityManager->clear();

        // Entity should be in database now
        $found = $this->repository->findOneBy(['keyword' => 'test']);
        $this->assertNotNull($found);
        $this->assertEquals('test', $found->getKeyword());
        $foundCategory = $found->getCategory();
        $this->assertNotNull($foundCategory);
        $this->assertSame($categoryId, $foundCategory->getId());
    }

    protected function tearDown(): void
    {
        $this->entityManager->createQuery('DELETE FROM App\Entity\CategoryKeyword')->execute();
        $this->entityManager->createQuery('DELETE FROM App\Entity\UserCategory')->execute();
        $this->entityManager->createQuery('DELETE FROM App\Entity\Category')->execute();
        $this->entityManager->createQuery('DELETE FROM App\Entity\User')->execute();
        parent::tearDown();
    }
}
