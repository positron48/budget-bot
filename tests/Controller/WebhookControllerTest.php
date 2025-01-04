<?php

namespace Tests\Controller;

use App\Controller\WebhookController;
use App\Service\TelegramBotService;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\Request;

class WebhookControllerTest extends TestCase
{
    /** @var TelegramBotService&MockObject */
    private TelegramBotService $telegramBotService;

    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;

    private WebhookController $controller;

    protected function setUp(): void
    {
        $this->telegramBotService = $this->createMock(TelegramBotService::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->controller = new WebhookController($this->telegramBotService, $this->logger);
    }

    public function testHandleWebhook(): void
    {
        $content = json_encode([
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => 123, 'type' => 'private'],
                'date' => time(),
                'text' => '/start',
            ],
        ]);

        if (false === $content) {
            throw new \RuntimeException('Failed to encode JSON');
        }

        $request = new Request([], [], [], [], [], ['CONTENT_TYPE' => 'application/json'], $content);

        $response = $this->controller->webhook($request);

        $this->assertEquals(200, $response->getStatusCode());
        $this->assertEquals('OK', $response->getContent());
    }

    public function testHandleWebhookInvalidJson(): void
    {
        $request = new Request([], [], [], [], [], ['CONTENT_TYPE' => 'application/json'], 'invalid json');

        $response = $this->controller->webhook($request);

        $this->assertEquals(400, $response->getStatusCode());
        $this->assertEquals('Invalid request', $response->getContent());
    }
}
