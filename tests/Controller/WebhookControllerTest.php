<?php

namespace App\Tests\Controller;

use App\Controller\WebhookController;
use App\Service\TelegramBotService;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\Request;

class WebhookControllerTest extends TestCase
{
    private WebhookController $controller;
    private MockObject&TelegramBotService $botService;
    private MockObject&LoggerInterface $logger;

    protected function setUp(): void
    {
        $this->botService = $this->createMock(TelegramBotService::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->controller = new WebhookController($this->botService, $this->logger);
    }

    public function testWebhookWithValidUpdate(): void
    {
        $content = json_encode([
            'update_id' => 123456789,
            'message' => [
                'message_id' => 1,
                'from' => [
                    'id' => 123456789,
                    'first_name' => 'Test',
                    'username' => 'test',
                ],
                'chat' => [
                    'id' => 123456789,
                    'first_name' => 'Test',
                    'username' => 'test',
                    'type' => 'private',
                ],
                'date' => 1704371700,
                'text' => '/start',
            ],
        ]);

        $request = new Request([], [], [], [], [], [], $content);

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context = []) {
                static $calls = 0;
                ++$calls;

                if (1 === $calls) {
                    $this->assertEquals('Received webhook request', $message);
                } elseif (2 === $calls) {
                    $this->assertEquals('Update processed successfully', $message);
                }

                return null;
            });

        $this->botService->expects($this->once())
            ->method('handleUpdate')
            ->with($this->isType('array'));

        $response = $this->controller->webhook($request);

        $this->assertEquals(200, $response->getStatusCode());
        $this->assertEquals('OK', $response->getContent());
    }

    public function testWebhookWithInvalidJson(): void
    {
        $request = new Request([], [], [], [], [], [], 'invalid json');

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Received webhook request', $this->anything());

        $this->logger->expects($this->once())
            ->method('error')
            ->with('Invalid request: unable to decode JSON');

        $this->botService->expects($this->never())
            ->method('handleUpdate');

        $response = $this->controller->webhook($request);

        $this->assertEquals(400, $response->getStatusCode());
        $this->assertEquals('Invalid request', $response->getContent());
    }

    public function testWebhookWithError(): void
    {
        $content = json_encode([
            'update_id' => 123456789,
            'message' => [
                'message_id' => 1,
                'text' => '/start',
            ],
        ]);

        $request = new Request([], [], [], [], [], [], $content);

        $this->botService->expects($this->once())
            ->method('handleUpdate')
            ->willThrowException(new \Exception('Test error'));

        $this->logger->expects($this->once())
            ->method('error')
            ->with('Error processing update: Test error', $this->anything());

        $response = $this->controller->webhook($request);

        $this->assertEquals(500, $response->getStatusCode());
        $this->assertEquals('Error processing update', $response->getContent());
    }
}
