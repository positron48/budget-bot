<?php

namespace App\Tests\Integration;

use App\Service\TelegramBotService;
use App\Tests\Integration\DataFixtures\TestFixtures;
use App\Tests\Mock\ResponseCollector;

class BudgetBotIntegrationTest extends IntegrationTestCase
{
    private const TELEGRAM_ID = 123456;

    private TelegramBotService $botService;
    private TestFixtures $fixtures;
    private ResponseCollector $responseCollector;

    protected function setUp(): void
    {
        parent::setUp();

        $this->botService = self::getContainer()->get(TelegramBotService::class);
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

    /**
     * @group skip
     */
    public function testFullUserJourney(): void
    {
        $this->markTestSkipped('Temporarily disabled');
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
