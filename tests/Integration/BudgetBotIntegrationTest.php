<?php

namespace App\Tests\Integration;

use App\Service\GoogleSheetsService;
use App\Service\TelegramBotService;
use App\Tests\Integration\DataFixtures\TestFixtures;
use App\Tests\Mock\ResponseCollector;

class BudgetBotIntegrationTest extends IntegrationTestCase
{
    private const TELEGRAM_ID = 123456;

    private TelegramBotService $botService;
    private TestFixtures $fixtures;
    private ResponseCollector $responseCollector;
    private GoogleSheetsService $sheetsService;

    protected function setUp(): void
    {
        parent::setUp();

        // Mock GoogleSheetsService
        $this->sheetsService = $this->createMock(GoogleSheetsService::class);
        $this->sheetsService->method('handleSpreadsheetId')
            ->willReturnArgument(0);

        // Replace real service with mock
        $container = self::getContainer();
        $container->set(GoogleSheetsService::class, $this->sheetsService);

        $this->botService = $container->get(TelegramBotService::class);
        $this->fixtures = new TestFixtures($this->entityManager);
        ResponseCollector::resetInstance();
        $this->responseCollector = ResponseCollector::getInstance();

        // Load test data
        $this->fixtures->load();

        // Reset responses
        $this->responseCollector->reset();
    }

    protected function tearDown(): void
    {
        parent::tearDown();
        ResponseCollector::resetInstance();
    }

    /** @return array<int, string> */
    private function getResponses(): array
    {
        return $this->responseCollector->getResponses();
    }

    private function clearTestData(): void
    {
        $this->entityManager->createQuery('DELETE FROM App\Entity\UserSpreadsheet')->execute();
        $this->entityManager->createQuery('DELETE FROM App\Entity\CategoryKeyword')->execute();
        $this->entityManager->createQuery('DELETE FROM App\Entity\UserCategory')->execute();
        $this->entityManager->createQuery('DELETE FROM App\Entity\User')->execute();
    }

    /**
     * @group skip
     */
    public function testFullUserJourney(): void
    {
        // Clear test data
        $this->clearTestData();

        // 1. Start command - welcome message
        $this->botService->handleUpdate([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => self::TELEGRAM_ID],
                'text' => '/start',
            ],
        ]);

        $responses = $this->getResponses();
        $this->assertCount(1, $responses);
        $this->assertStringContainsString('Привет! Я помогу вести учет доходов и расходов в Google Таблицах', $responses[0]);
        $this->assertStringContainsString('/list - список доступных таблиц', $responses[0]);
        $this->assertStringContainsString('/add - добавить таблицу', $responses[0]);

        // Reset responses for the next command
        $this->responseCollector->reset();

        // 2. List command - empty list
        $this->botService->handleUpdate([
            'update_id' => 2,
            'message' => [
                'message_id' => 2,
                'chat' => ['id' => self::TELEGRAM_ID],
                'text' => '/list',
            ],
        ]);

        $responses = $this->getResponses();
        $this->assertCount(1, $responses);
        $this->assertStringContainsString('У вас пока нет добавленных таблиц', $responses[0]);
        $this->assertStringContainsString('Используйте команду /add чтобы добавить таблицу', $responses[0]);

        // Reset responses for the next command
        $this->responseCollector->reset();

        // 3. Add command - request spreadsheet ID
        $this->botService->handleUpdate([
            'update_id' => 3,
            'message' => [
                'message_id' => 3,
                'chat' => ['id' => self::TELEGRAM_ID],
                'text' => '/add',
            ],
        ]);

        $responses = $this->getResponses();
        $this->assertCount(1, $responses);
        $this->assertStringContainsString('Отправьте ссылку на таблицу или её идентификатор', $responses[0]);

        // Reset responses for the next command
        $this->responseCollector->reset();

        // 4. Send spreadsheet ID
        $spreadsheetId = '1-BxqnQqyBPjyuRxMSrwQ2FDDxR-sQGQs_EZbZEn_Xzc';
        $this->botService->handleUpdate([
            'update_id' => 4,
            'message' => [
                'message_id' => 4,
                'chat' => ['id' => self::TELEGRAM_ID],
                'text' => $spreadsheetId,
            ],
        ]);

        $responses = $this->getResponses();
        $this->assertCount(1, $responses);
        $this->assertStringContainsString('Выберите месяц и год', $responses[0]);
        $this->assertStringContainsString('введите их в формате "Месяц Год"', $responses[0]);

        // Reset responses for the next command
        $this->responseCollector->reset();

        // 5. Send month and year
        $monthName = 'Январь';
        $year = 2025;
        $this->botService->handleUpdate([
            'update_id' => 5,
            'message' => [
                'message_id' => 5,
                'chat' => ['id' => self::TELEGRAM_ID],
                'text' => sprintf('%s %d', $monthName, $year),
            ],
        ]);

        $responses = $this->getResponses();
        $this->assertCount(1, $responses);
        $this->assertStringContainsString('успешно добавлена', $responses[0]);
    }

    /**
     * @group skip
     */
    public function testCategoryManagement(): void
    {
        $this->markTestSkipped('Temporarily disabled');
    }

    public function testErrorHandling(): void
    {
        error_log('Starting testErrorHandling');

        // 1. Invalid expense format (no amount)
        $this->botService->handleUpdate([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => self::TELEGRAM_ID],
                'text' => 'invalid',
            ],
        ]);

        $responses = $this->getResponses();
        error_log('After first command, responses count: '.count($responses));
        error_log('Responses: '.json_encode($responses, JSON_UNESCAPED_UNICODE));
        $this->assertNotEmpty($responses);
        $this->assertArrayHasKey(0, $responses);
        $this->assertStringContainsString('Неверный формат сообщения', $responses[0]);
        $this->assertStringContainsString('Используйте формат: "[дата] [+]сумма описание"', $responses[0]);

        // 2. Invalid amount format
        $this->botService->handleUpdate([
            'update_id' => 2,
            'message' => [
                'message_id' => 2,
                'chat' => ['id' => self::TELEGRAM_ID],
                'text' => 'taxi abc',
            ],
        ]);

        $responses = $this->getResponses();
        error_log('After second command, responses count: '.count($responses));
        error_log('Responses: '.json_encode($responses, JSON_UNESCAPED_UNICODE));
        $this->assertArrayHasKey(1, $responses);
        $this->assertStringContainsString('Неверный формат сообщения', $responses[1]);
        $this->assertStringContainsString('Используйте формат: "[дата] [+]сумма описание"', $responses[1]);
    }
}
