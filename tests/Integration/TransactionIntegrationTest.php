<?php

namespace App\Tests\Integration;

use App\Repository\UserCategoryRepository;
use App\Repository\UserRepository;
use App\Tests\Integration\DataFixtures\TestFixtures;
use Doctrine\Common\DataFixtures\Executor\ORMExecutor;
use Doctrine\Common\DataFixtures\Loader;
use Doctrine\Common\DataFixtures\Purger\ORMPurger;

class TransactionIntegrationTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private UserRepository $userRepository;
    private UserCategoryRepository $categoryRepository;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();
        $this->userRepository = $container->get(UserRepository::class);
        $this->categoryRepository = $container->get(UserCategoryRepository::class);

        // Load test fixtures using DoctrineFixturesBundle
        $loader = new Loader();
        $loader->addFixture(new TestFixtures());

        $executor = new ORMExecutor($this->entityManager, new ORMPurger());
        $executor->execute($loader->getFixtures());
    }

    public function testCategoryListDisplay(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Check categories list
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Выберите действие');

        // Check expense categories
        $this->executeCommand('Категории расходов', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Питание');
        $this->assertLastMessageContains('Кафе/Ресторан');
        $this->assertLastMessageContains('Транспорт');

        // Check income categories
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->executeCommand('Категории доходов', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Зарплата');

        // Verify that all expected categories are present
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        $expenseCategories = $this->categoryRepository->findBy([
            'user' => $user,
            'type' => 'expense',
        ]);
        $this->assertCount(3, $expenseCategories);

        $incomeCategories = $this->categoryRepository->findBy([
            'user' => $user,
            'type' => 'income',
        ]);
        $this->assertCount(1, $incomeCategories);
    }
}
