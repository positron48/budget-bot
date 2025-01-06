<?php

namespace App\Tests\Controller;

use App\Controller\WebhookController;
use App\Service\TelegramBotService;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Symfony\Component\HttpFoundation\Request;

class WebhookControllerTest extends TestCase
{
    private WebhookController $controller;
    /** @var TelegramBotService&MockObject */
    private TelegramBotService $telegramBotService;

    protected function setUp(): void
    {
        $this->telegramBotService = $this->createMock(TelegramBotService::class);
        $this->controller = new WebhookController($this->telegramBotService);
    }

    public function testWebhook(): void
    {
        $content = json_encode([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => 123456],
                'text' => '/start',
            ],
        ], JSON_THROW_ON_ERROR);

        $request = new Request([], [], [], [], [], [], $content);

        $this->telegramBotService->expects($this->once())
            ->method('handleUpdate');

        $response = $this->controller->webhook($request);

        $this->assertEquals(200, $response->getStatusCode());
        $this->assertEquals('{"status":"ok"}', $response->getContent());
    }
}
