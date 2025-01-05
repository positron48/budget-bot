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
        ], JSON_UNESCAPED_UNICODE);

        if (false === $content) {
            throw new \RuntimeException('Failed to encode JSON');
        }

        $request = new Request([], [], [], [], [], ['CONTENT_TYPE' => 'application/json'], $content);

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context) use ($content) {
                static $callNumber = 0;
                ++$callNumber;

                if (1 === $callNumber) {
                    $this->assertEquals('Received webhook request', $message);
                    $this->assertEquals([
                        'content' => $content,
                    ], $context);
                } elseif (2 === $callNumber) {
                    $this->assertEquals('Webhook request processed successfully', $message);
                    $this->assertEquals([], $context);
                }
            });

        $response = $this->controller->webhook($request);

        $this->assertEquals(200, $response->getStatusCode());
        $this->assertEquals('OK', $response->getContent());
    }

    public function testHandleWebhookInvalidJson(): void
    {
        $request = new Request([], [], [], [], [], ['CONTENT_TYPE' => 'application/json'], 'invalid json');

        $this->logger->expects($this->once())
            ->method('error')
            ->with(
                'Failed to decode webhook request content',
                [
                    'content' => 'invalid json',
                    'error' => 'Syntax error',
                ]
            );

        $response = $this->controller->webhook($request);

        $this->assertEquals(400, $response->getStatusCode());
        $this->assertEquals('Invalid request', $response->getContent());
    }
}
