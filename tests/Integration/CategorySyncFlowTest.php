<?php

namespace App\Tests\Integration;

use App\Repository\UserRepository;

class CategorySyncFlowTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private const TEST_SPREADSHEET_ID = '1234567890';

    private UserRepository $userRepository;

    protected function setUp(): void
    {
        parent::setUp();

        $this->userRepository = self::getContainer()->get(UserRepository::class);

        // Set up test data
        $this->setupTestSpreadsheet(self::TEST_SPREADSHEET_ID);
        $this->setupTestCategories(self::TEST_SPREADSHEET_ID);
    }

    public function testUserCreationOnStart(): void
    {
        // Execute /start command
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Verify welcome message
        $this->assertMessageCount(1);
        $this->assertLastMessageContains('Привет!');

        // Verify user creation
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);
    }

    public function testSpreadsheetSetup(): void
    {
        // Start with user creation
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add spreadsheet
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Отправьте ссылку на таблицу');

        $this->executeCommand('https://docs.google.com/spreadsheets/d/'.self::TEST_SPREADSHEET_ID.'/edit', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Выберите месяц');

        $this->executeCommand('Январь 2025', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Таблица за Январь 2025 успешно добавлена');
    }

    public function testCategorySync(): void
    {
        // Setup initial state
        $this->testSpreadsheetSetup();

        // Sync categories
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        $messages = $this->telegramApi->getMessages();
        $lastMessages = array_slice($messages, -2);

        // Verify categories cleared message
        $this->assertStringContainsString('Пользовательские категории очищены:', $lastMessages[0]['text']);
        $this->assertStringContainsString('- Расходы: 3', $lastMessages[0]['text']);
        $this->assertStringContainsString('- Доходы: 2', $lastMessages[0]['text']);

        // Verify sync results message
        $this->assertStringContainsString('Синхронизация категорий завершена:', $lastMessages[1]['text']);
        $this->assertStringContainsString('- Расходы: Питание, Транспорт, Развлечения', $lastMessages[1]['text']);
        $this->assertStringContainsString('- Доходы: Зарплата, Фриланс', $lastMessages[1]['text']);
    }

    public function testCategoryListing(): void
    {
        // Setup categories
        $this->testCategorySync();

        // Check categories list
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Выберите действие');

        // Check expense categories
        $this->executeCommand('Категории расходов', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Питание');
        $this->assertLastMessageContains('Транспорт');
        $this->assertLastMessageContains('Развлечения');

        // Check income categories
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->executeCommand('Категории доходов', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Зарплата');
        $this->assertLastMessageContains('Фриланс');
    }

    public function testCategoryMapping(): void
    {
        // Setup categories
        $this->testCategorySync();

        // Add mapping
        $this->executeCommand('/map еда = Питание', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Добавлено сопоставление: "еда" → "Питание"');

        // Verify mapping
        $this->executeCommand('/map еда', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Описание "еда" соответствует категории "Питание"');
    }

    public function testExpenseAdditionWithMapping(): void
    {
        // Setup mapping
        $this->testCategoryMapping();

        // Add expense with mapped category
        $this->executeCommand('1500 еда обед', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен в категорию "Питание"');

        // Add expense with unmapped category
        $this->executeCommand('1000 продукты', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Не удалось определить категорию для "продукты"');
        $this->assertLastMessageContains('Выберите категорию из списка');

        // Select category for unmapped keyword
        $this->executeCommand('Питание', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен в категорию "Питание"');

        // Verify automatic mapping creation
        $this->executeCommand('/map продукты', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Описание "продукты" соответствует категории "Питание"');
    }
}
