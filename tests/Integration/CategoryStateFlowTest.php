<?php

namespace App\Tests\Integration;

use App\Repository\UserRepository;
use App\Tests\Integration\DataFixtures\TestFixtures;
use Doctrine\Common\DataFixtures\Executor\ORMExecutor;
use Doctrine\Common\DataFixtures\Loader;
use Doctrine\Common\DataFixtures\Purger\ORMPurger;

class CategoryStateFlowTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private UserRepository $userRepository;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();
        $this->userRepository = $container->get(UserRepository::class);

        // Load test fixtures using DoctrineFixturesBundle
        $loader = new Loader();
        $loader->addFixture(new TestFixtures());

        $executor = new ORMExecutor($this->entityManager, new ORMPurger());
        $executor->execute($loader->getFixtures());
    }

    public function testCategorySelectionFlow(): void
    {
        // Start with user creation
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add expense with unknown category
        $this->executeCommand('1500 неизвестная категория', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Не удалось определить категорию для "неизвестная категория"');
        $this->assertLastMessageContains('Выберите категорию из списка');

        // Verify user state
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_CATEGORY_SELECTION', $user->getState());

        // Choose "Add mapping" option
        $this->executeCommand('Добавить сопоставление', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Введите сопоставление в формате "слово = категория"');
        $this->assertEquals('WAITING_CATEGORY_MAPPING', $user->getState());

        // Add invalid mapping format
        $this->executeCommand('неверный формат', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Неверный формат. Используйте: слово = категория');

        // Add mapping with non-existent category
        $this->executeCommand('неизвестная = НесуществующаяКатегория', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Категория "НесуществующаяКатегория" не найдена');

        // Add valid mapping
        $this->executeCommand('неизвестная = Питание', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Категория: Питание');

        // Verify state is cleared
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEmpty($user->getState());

        // Try adding another expense with the same keyword
        $this->executeCommand('2000 неизвестная категория', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Категория: Питание');
    }

    public function testCategoryListFlow(): void
    {
        // Start with user creation
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Check categories command
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Выберите действие');

        // Verify user state
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_CATEGORIES_ACTION', $user->getState());

        // Check expense categories
        $this->executeCommand('Категории расходов', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Категории расходов:');
        $this->assertLastMessageContains('Питание');
        $this->assertLastMessageContains('Кафе/Ресторан');
        $this->assertLastMessageContains('Транспорт');

        // Verify state is cleared
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEmpty($user->getState());

        // Check categories command again
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->assertEquals('WAITING_CATEGORIES_ACTION', $user->getState());

        // Check income categories
        $this->executeCommand('Категории доходов', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Категории доходов:');
        $this->assertLastMessageContains('Зарплата');
        $this->assertLastMessageContains('Премия');

        // Verify state is cleared
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEmpty($user->getState());

        // Test invalid action
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->executeCommand('Неизвестное действие', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Неизвестное действие');
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
        $this->assertEmpty($user->getState());
    }
}
