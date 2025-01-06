<?php

namespace App\Tests\Integration;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\Google\GoogleApiClientInterface;
use App\Service\TelegramApiServiceInterface;
use App\Service\TelegramBotService;
use App\Tests\Mock\TelegramApiMock;
use App\Tests\Mock\TestGoogleApiClient;
use Longman\TelegramBot\Entities\Update;

class CategorySyncFlowTest extends IntegrationTestCase
{
    private TelegramBotService $botService;
    private UserRepository $userRepository;
    private TelegramApiServiceInterface&TelegramApiMock $telegramApi;
    private GoogleApiClientInterface&TestGoogleApiClient $googleApiClient;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();

        // Get services from container
        /** @var TelegramApiServiceInterface&TelegramApiMock $telegramApi */
        $telegramApi = $container->get(TelegramApiServiceInterface::class);
        $this->telegramApi = $telegramApi;

        /** @var GoogleApiClientInterface&TestGoogleApiClient $googleApiClient */
        $googleApiClient = $container->get(GoogleApiClientInterface::class);
        $this->googleApiClient = $googleApiClient;

        $this->botService = $container->get(TelegramBotService::class);
        $this->userRepository = $container->get(UserRepository::class);

        // Set up a test spreadsheet in Google API mock
        $spreadsheetId = '1234567890';
        $this->googleApiClient->addAccessibleSpreadsheet($spreadsheetId);
        $this->googleApiClient->setSpreadsheetTitle($spreadsheetId, 'Test Budget');

        // Set up mock data for categories in the spreadsheet
        $this->googleApiClient->setValues($spreadsheetId, 'Сводка!B28:B', [
            ['Питание'],
            ['Транспорт'],
            ['Развлечения'],
        ]);
        $this->googleApiClient->setValues($spreadsheetId, 'Сводка!H28:H', [
            ['Зарплата'],
            ['Фриланс'],
        ]);
    }

    public function testCategorySyncFlow(): void
    {
        $chatId = 123456;

        // Step 1: Start command
        $this->executeCommand('/start', $chatId);

        $messages = $this->telegramApi->getMessages();
        $this->assertCount(1, $messages);
        $this->assertStringContainsString('Привет!', $messages[0]['text']);

        // Verify user was created
        $user = $this->userRepository->findOneBy(['telegramId' => $chatId]);
        $this->assertNotNull($user);

        // Step 2: Add spreadsheet
        $this->executeCommand('/add', $chatId);
        $spreadsheetLink = 'https://docs.google.com/spreadsheets/d/1234567890/edit';
        $this->executeCommand($spreadsheetLink, $chatId);
        $this->executeCommand('Январь 2025', $chatId);

        // Step 3: Sync categories
        $this->executeCommand('/sync_categories', $chatId);

        $messages = $this->telegramApi->getMessages();
        $lastMessages = array_slice($messages, -2); // Get last 2 messages

        // Check first message about clearing categories
        $this->assertStringContainsString('Пользовательские категории очищены:', $lastMessages[0]['text']);
        $this->assertStringContainsString('- Расходы: 3', $lastMessages[0]['text']);
        $this->assertStringContainsString('- Доходы: 2', $lastMessages[0]['text']);

        // Check second message about sync results
        $this->assertStringContainsString('Синхронизация категорий завершена:', $lastMessages[1]['text']);
        $this->assertStringContainsString('Добавлены в базу данных:', $lastMessages[1]['text']);
        $this->assertStringContainsString('- Расходы: Питание, Транспорт, Развлечения', $lastMessages[1]['text']);
        $this->assertStringContainsString('- Доходы: Зарплата, Фриланс', $lastMessages[1]['text']);

        // Step 4: Check categories list
        $this->executeCommand('/categories', $chatId);
        $messages = $this->telegramApi->getMessages();
        $categoriesMessage = end($messages);
        $this->assertStringContainsString('Выберите действие', $categoriesMessage['text']);

        // Select expense categories
        $this->executeCommand('Категории расходов', $chatId);
        $messages = $this->telegramApi->getMessages();
        $expenseCategoriesMessage = end($messages);
        $this->assertStringContainsString('Питание', $expenseCategoriesMessage['text']);
        $this->assertStringContainsString('Транспорт', $expenseCategoriesMessage['text']);
        $this->assertStringContainsString('Развлечения', $expenseCategoriesMessage['text']);

        // Select income categories
        $this->executeCommand('/categories', $chatId);
        $this->executeCommand('Категории доходов', $chatId);
        $messages = $this->telegramApi->getMessages();
        $incomeCategoriesMessage = end($messages);
        $this->assertStringContainsString('Зарплата', $incomeCategoriesMessage['text']);
        $this->assertStringContainsString('Фриланс', $incomeCategoriesMessage['text']);

        // Step 5: Add category mapping
        $this->executeCommand('/map еда = Питание', $chatId);
        $messages = $this->telegramApi->getMessages();
        $mapMessage = end($messages);
        $this->assertStringContainsString('Добавлено сопоставление: "еда" → "Питание"', $mapMessage['text']);

        // Step 6: Check that mapping works
        $this->executeCommand('/map еда', $chatId);
        $messages = $this->telegramApi->getMessages();
        $checkMessage = end($messages);
        $this->assertStringContainsString('Описание "еда" соответствует категории "Питание"', $checkMessage['text']);

        // Step 7: Add expense using mapped category
        $this->executeCommand('1500 еда обед', $chatId);
        $messages = $this->telegramApi->getMessages();
        $expenseMessage = end($messages);
        $this->assertStringContainsString('Расход успешно добавлен в категорию "Питание"', $expenseMessage['text']);

        // Step 8: Add expense with unmapped category
        $this->executeCommand('1000 продукты', $chatId);
        $messages = $this->telegramApi->getMessages();
        $categoryPromptMessage = end($messages);
        $this->assertStringContainsString('Не удалось определить категорию для "продукты"', $categoryPromptMessage['text']);
        $this->assertStringContainsString('Выберите категорию из списка', $categoryPromptMessage['text']);

        // Select category for unmapped keyword
        $this->executeCommand('Питание', $chatId);
        $messages = $this->telegramApi->getMessages();
        $expenseMessage = end($messages);
        $this->assertStringContainsString('Расход успешно добавлен в категорию "Питание"', $expenseMessage['text']);

        // Step 9: Verify that mapping was automatically created
        $this->executeCommand('/map продукты', $chatId);
        $messages = $this->telegramApi->getMessages();
        $mapMessage = end($messages);
        $this->assertStringContainsString('Описание "продукты" соответствует категории "Питание"', $mapMessage['text']);
    }

    private function executeCommand(string $text, int $chatId): void
    {
        $update = new Update([
            'update_id' => random_int(1, 100000),
            'message' => [
                'message_id' => random_int(1, 100000),
                'from' => [
                    'id' => $chatId,
                    'first_name' => 'Test User',
                    'is_bot' => false,
                ],
                'chat' => [
                    'id' => $chatId,
                    'type' => 'private',
                ],
                'date' => time(),
                'text' => $text,
            ],
        ]);

        $this->botService->handleUpdate($update);
    }
}
