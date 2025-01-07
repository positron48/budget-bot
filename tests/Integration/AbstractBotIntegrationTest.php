<?php

namespace App\Tests\Integration;

use App\Service\Google\GoogleApiClientInterface;
use App\Service\TelegramApiServiceInterface;
use App\Service\TelegramBotService;
use App\Tests\Mock\TelegramApiMock;
use App\Tests\Mock\TestGoogleApiClient;
use Longman\TelegramBot\Entities\Update;

abstract class AbstractBotIntegrationTest extends IntegrationTestCase
{
    protected TelegramBotService $botService;
    protected TelegramApiServiceInterface&TelegramApiMock $telegramApi;
    protected GoogleApiClientInterface&TestGoogleApiClient $googleApiClient;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();

        /** @var TelegramApiServiceInterface&TelegramApiMock $telegramApi */
        $telegramApi = $container->get(TelegramApiServiceInterface::class);
        $this->telegramApi = $telegramApi;

        /** @var GoogleApiClientInterface&TestGoogleApiClient $googleApiClient */
        $googleApiClient = $container->get(GoogleApiClientInterface::class);
        $this->googleApiClient = $googleApiClient;

        $this->botService = $container->get(TelegramBotService::class);
    }

    protected function executeCommand(string $text, int $chatId): void
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

    protected function assertLastMessageContains(string $expectedText): void
    {
        $messages = $this->telegramApi->getMessages();
        $lastMessage = end($messages);
        $this->assertStringContainsString($expectedText, $lastMessage['text']);
    }

    protected function assertMessageCount(int $expectedCount): void
    {
        $messages = $this->telegramApi->getMessages();
        $this->assertCount($expectedCount, $messages);
    }

    protected function setupTestSpreadsheet(string $spreadsheetId, string $title = 'Test Budget'): void
    {
        $this->googleApiClient->addAccessibleSpreadsheet($spreadsheetId);
        $this->googleApiClient->setSpreadsheetTitle($spreadsheetId, $title);
    }

    protected function setupTestCategories(string $spreadsheetId): void
    {
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
}
