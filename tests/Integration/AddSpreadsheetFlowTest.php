<?php

namespace App\Tests\Integration;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Repository\UserSpreadsheetRepository;
use App\Service\Google\GoogleApiClientInterface;
use App\Service\TelegramApiServiceInterface;
use App\Service\TelegramBotService;
use App\Tests\Mock\TelegramApiMock;
use App\Tests\Mock\TestGoogleApiClient;
use Longman\TelegramBot\Entities\Update;

class AddSpreadsheetFlowTest extends IntegrationTestCase
{
    private TelegramBotService $botService;
    private UserRepository $userRepository;
    private UserSpreadsheetRepository $spreadsheetRepository;
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
        $this->spreadsheetRepository = $container->get(UserSpreadsheetRepository::class);

        // Set up a test spreadsheet in Google API mock
        $spreadsheetId = '1234567890';
        $this->googleApiClient->addAccessibleSpreadsheet($spreadsheetId);
        $this->googleApiClient->setSpreadsheetTitle($spreadsheetId, 'Test Budget');

        // Set up mock data for categories in the spreadsheet
        $this->googleApiClient->setValues($spreadsheetId, 'Settings!A2:A50', [
            ['Питание'],
            ['Транспорт'],
            ['Развлечения'],
        ]);
        $this->googleApiClient->setValues($spreadsheetId, 'Settings!C2:C50', [
            ['Зарплата'],
            ['Фриланс'],
        ]);
    }

    public function testCompleteAddSpreadsheetFlow(): void
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

        // Step 2: Add command
        $this->executeCommand('/add', $chatId);

        $messages = $this->telegramApi->getMessages();
        $this->assertCount(2, $messages);
        $this->assertStringContainsString('Отправьте ссылку на таблицу', $messages[1]['text']);
        $this->assertEquals('WAITING_SPREADSHEET_ID', $user->getState());

        // Step 3: Send spreadsheet link
        $spreadsheetLink = 'https://docs.google.com/spreadsheets/d/1234567890/edit';
        $this->executeCommand($spreadsheetLink, $chatId);

        $messages = $this->telegramApi->getMessages();
        $this->assertCount(3, $messages);
        $this->assertStringContainsString('Выберите месяц', $messages[2]['text']);

        // Step 4: Select month
        $this->executeCommand('Январь 2024', $chatId);

        $messages = $this->telegramApi->getMessages();
        $this->assertCount(4, $messages);
        $this->assertStringContainsString('Таблица за Январь 2024 успешно добавлена', $messages[3]['text']);

        // Refresh user from database
        $user = $this->userRepository->findOneBy(['telegramId' => $chatId]);
        $this->assertNotNull($user);
        $this->assertTrue(null === $user->getState() || '' === $user->getState(), 'State should be null or empty string');

        // Verify spreadsheet was saved
        $spreadsheet = $this->spreadsheetRepository->findOneBy(['user' => $user]);
        $this->assertNotNull($spreadsheet);
        $this->assertEquals('1234567890', $spreadsheet->getSpreadsheetId());
        $this->assertEquals(1, $spreadsheet->getMonth());
        $this->assertEquals(2024, $spreadsheet->getYear());

        // Step 5: Check list command
        $this->executeCommand('/list', $chatId);

        $messages = $this->telegramApi->getMessages();
        $this->assertCount(5, $messages);
        $this->assertStringContainsString('Январь 2024', $messages[4]['text']);

        // Step 6: Remove command
        $this->executeCommand('/remove', $chatId);

        $messages = $this->telegramApi->getMessages();
        $this->assertCount(6, $messages);
        $this->assertStringContainsString('Выберите таблицу для удаления', $messages[5]['text']);
        $this->assertSame('WAITING_REMOVE_SPREADSHEET', $user->getState());

        // Step 7: Select spreadsheet to delete
        $this->executeCommand('Январь 2024', $chatId);

        $messages = $this->telegramApi->getMessages();
        $this->assertCount(7, $messages);
        $this->assertStringContainsString('Таблица за Январь 2024 успешно удалена', $messages[6]['text']);

        // Verify spreadsheet was deleted
        $spreadsheet = $this->spreadsheetRepository->findOneBy(['user' => $user]);
        $this->assertNull($spreadsheet);

        // Step 8: Check list command shows no spreadsheets
        $this->executeCommand('/list', $chatId);

        $messages = $this->telegramApi->getMessages();
        $this->assertCount(8, $messages);
        $this->assertStringContainsString('У вас пока нет добавленных таблиц', $messages[7]['text']);
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
