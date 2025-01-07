<?php

namespace App\Tests\Integration;

class SyncCategoriesCommandTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private const TEST_SPREADSHEET_ID = 'test_spreadsheet';
    private const INVALID_SPREADSHEET_ID = 'invalid_id';

    protected function setUp(): void
    {
        parent::setUp();
    }

    private function setupInitialState(): void
    {
        // Setup test spreadsheet first
        $this->setupTestSpreadsheet(self::TEST_SPREADSHEET_ID);
        $this->googleApiClient->setSpreadsheetAccessible(self::TEST_SPREADSHEET_ID, true);
        $this->googleApiClient->setSpreadsheetTitle(self::TEST_SPREADSHEET_ID, 'Test Budget');

        // Execute /start command
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add spreadsheet
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->executeCommand('https://docs.google.com/spreadsheets/d/'.self::TEST_SPREADSHEET_ID.'/edit', self::TEST_CHAT_ID);
        $this->executeCommand('Январь 2025', self::TEST_CHAT_ID);
    }

    public function testSupportsMethod(): void
    {
        $command = self::getContainer()->get('App\Service\Command\SyncCategoriesCommand');
        $this->assertTrue($command->supports('/sync_categories'));
        $this->assertFalse($command->supports('/start'));
        $this->assertFalse($command->supports('/categories'));
        $this->assertFalse($command->supports('invalid command'));
    }

    public function testSyncWithoutSpreadsheet(): void
    {
        // Execute /start command
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Try to sync without adding a spreadsheet
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'У вас пока нет добавленных таблиц. Используйте команду /add чтобы добавить таблицу',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testSyncWithInvalidSpreadsheetId(): void
    {
        // Execute /start command
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Make the spreadsheet accessible first to allow it to be added
        $this->googleApiClient->setSpreadsheetAccessible(self::INVALID_SPREADSHEET_ID, true);
        $this->googleApiClient->setSpreadsheetTitle(self::INVALID_SPREADSHEET_ID, 'Invalid Spreadsheet');

        // Add invalid spreadsheet
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->executeCommand('https://docs.google.com/spreadsheets/d/'.self::INVALID_SPREADSHEET_ID.'/edit', self::TEST_CHAT_ID);
        $this->executeCommand('Январь 2025', self::TEST_CHAT_ID);

        // Then make it throw an exception when trying to get values
        $this->googleApiClient->throwOnGetValues(self::INVALID_SPREADSHEET_ID, new \Exception('Invalid spreadsheet ID'));

        // Sync categories
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Не удалось синхронизировать категории. Попробуйте еще раз.',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testSyncWithEmptyCategories(): void
    {
        // Setup initial state with empty categories
        $this->setupTestSpreadsheet(self::TEST_SPREADSHEET_ID, true);
        $this->googleApiClient->setSpreadsheetAccessible(self::TEST_SPREADSHEET_ID, true);
        $this->googleApiClient->setSpreadsheetTitle(self::TEST_SPREADSHEET_ID, 'Test Budget');

        // Execute /start command
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add spreadsheet
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->executeCommand('https://docs.google.com/spreadsheets/d/'.self::TEST_SPREADSHEET_ID.'/edit', self::TEST_CHAT_ID);
        $this->executeCommand('Январь 2025', self::TEST_CHAT_ID);

        // Sync categories
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Все категории уже синхронизированы',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testPartialSync(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Add some categories to the spreadsheet
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Сводка!B28:B', [['Новая категория расходов']]);
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Сводка!H28:H', [['Новая категория доходов']]);

        // Sync categories
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Синхронизация категорий завершена:',
            $lastMessages[count($lastMessages) - 1]['text']
        );
        $this->assertStringContainsString(
            'Новая категория расходов',
            $lastMessages[count($lastMessages) - 1]['text']
        );
        $this->assertStringContainsString(
            'Новая категория доходов',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testMultipleConsecutiveSyncs(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // First sync with some categories
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Сводка!B28:B', [['Тест']]);
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        // Clear the messages from the first sync
        $this->telegramApi->clearMessages();

        // Clear the values to simulate no changes
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Сводка!B28:B', []);
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Сводка!H28:H', []);

        // Second sync with no changes
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Все категории уже синхронизированы',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testSyncWithoutUser(): void
    {
        // Try to sync without starting the bot first
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Пожалуйста, начните с команды /start',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testSyncWithNullSpreadsheetId(): void
    {
        // Setup initial state
        $this->setupInitialState();

        // Make the spreadsheet throw an exception when trying to get values
        $this->googleApiClient->throwOnGetValues(self::TEST_SPREADSHEET_ID, new \RuntimeException('Spreadsheet ID is null'));

        // Sync categories
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Не удалось синхронизировать категории. Попробуйте еще раз.',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }

    public function testSyncWithBidirectionalChanges(): void
    {
        // Setup initial state with empty categories
        $this->setupTestSpreadsheet(self::TEST_SPREADSHEET_ID, true);
        $this->googleApiClient->setSpreadsheetAccessible(self::TEST_SPREADSHEET_ID, true);
        $this->googleApiClient->setSpreadsheetTitle(self::TEST_SPREADSHEET_ID, 'Test Budget');

        // Execute /start command
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add spreadsheet
        $this->executeCommand('/add', self::TEST_CHAT_ID);
        $this->executeCommand('https://docs.google.com/spreadsheets/d/'.self::TEST_SPREADSHEET_ID.'/edit', self::TEST_CHAT_ID);
        $this->executeCommand('Январь 2025', self::TEST_CHAT_ID);

        // Add some categories to the spreadsheet
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Сводка!B28:B', [['Новая категория расходов 1']]);
        $this->googleApiClient->setValues(self::TEST_SPREADSHEET_ID, 'Сводка!H28:H', [['Новая категория доходов 1']]);

        // Add some categories to the database (this will be added to the sheet during sync)
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->executeCommand('Категории расходов', self::TEST_CHAT_ID);

        // Clear messages before sync
        $this->telegramApi->clearMessages();

        // Sync categories
        $this->executeCommand('/sync_categories', self::TEST_CHAT_ID);

        $lastMessages = $this->telegramApi->getMessages();
        $this->assertStringContainsString(
            'Синхронизация категорий завершена:',
            $lastMessages[count($lastMessages) - 1]['text']
        );
        $this->assertStringContainsString(
            'Добавлены в базу данных:',
            $lastMessages[count($lastMessages) - 1]['text']
        );
        $this->assertStringContainsString(
            'Новая категория расходов 1',
            $lastMessages[count($lastMessages) - 1]['text']
        );
        $this->assertStringContainsString(
            'Новая категория доходов 1',
            $lastMessages[count($lastMessages) - 1]['text']
        );
    }
}
