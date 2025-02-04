<?php

namespace App\Tests\Integration;

use App\Service\Google\GoogleApiClientInterface;
use App\Service\TelegramApiServiceInterface;
use App\Service\TelegramBotService;
use App\Tests\Mock\TelegramApiMock;
use App\Tests\Mock\TestGoogleApiClient;
use Longman\TelegramBot\Entities\Update;

/**
 * Base class for bot integration tests.
 *
 * @internal
 */
abstract class AbstractBotIntegrationTestCase extends IntegrationTestCase
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

        // Set fixed date for tests
        $this->setFixedTestDate();
    }

    protected function tearDown(): void
    {
        parent::tearDown();

        // Reset the fixed date after each test
        /** @var \App\Utility\DateTimeUtility $dateTimeUtility */
        $dateTimeUtility = self::getContainer()->get(\App\Utility\DateTimeUtility::class);
        $dateTimeUtility->resetCurrentDate();
    }

    protected function setFixedTestDate(): void
    {
        // Set fixed date to January 2025 for all tests
        $fixedDate = new \DateTime('2025-01-15');
        /** @var \App\Utility\DateTimeUtility $dateTimeUtility */
        $dateTimeUtility = self::getContainer()->get(\App\Utility\DateTimeUtility::class);
        $dateTimeUtility->setCurrentDate($fixedDate);
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

    protected function setupTestSpreadsheet(string $spreadsheetId, bool $emptyCategories = false): void
    {
        /** @var TestGoogleApiClient $client */
        $client = self::getContainer()->get(GoogleApiClientInterface::class);
        $client->setSpreadsheetAccessible($spreadsheetId, true);
        $client->setSpreadsheetTitle($spreadsheetId, 'Test Budget');

        if (!$emptyCategories) {
            $this->setupTestCategories($spreadsheetId);
        } else {
            // Set empty categories
            $client->setValues($spreadsheetId, 'Сводка!B28:B', []);
            $client->setValues($spreadsheetId, 'Сводка!H28:H', []);
        }
    }

    protected function setupTestCategories(string $spreadsheetId): void
    {
        /** @var TestGoogleApiClient $client */
        $client = self::getContainer()->get(GoogleApiClientInterface::class);

        // Set expense categories
        $client->setValues($spreadsheetId, 'Сводка!B28:B', [
            ['Питание'],
            ['Подарки'],
            ['Здоровье/медицина'],
            ['Дом'],
            ['Транспорт'],
            ['Личные расходы'],
            ['Домашние животные'],
            ['Коммунальные услуги'],
            ['Путешествия'],
            ['Одежда'],
            ['Развлечения'],
            ['Кафе/Ресторан'],
            ['Алко'],
            ['Образование'],
            ['Услуги'],
            ['Авто'],
        ]);

        // Set income categories
        $client->setValues($spreadsheetId, 'Сводка!H28:H', [
            ['Зарплата'],
            ['Премия'],
            ['Кешбек, др. бонусы'],
            ['Процентный доход'],
            ['Инвестиции'],
            ['Другое'],
        ]);
    }
}
