<?php

namespace App\Controller;

use App\Service\TelegramBotService;
use Psr\Log\LoggerInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class WebhookController
{
    private TelegramBotService $telegramBotService;
    private LoggerInterface $logger;

    public function __construct(TelegramBotService $telegramBotService, LoggerInterface $logger)
    {
        $this->telegramBotService = $telegramBotService;
        $this->logger = $logger;
    }

    public function webhook(Request $request): Response
    {
        $content = $request->getContent();

        $this->logger->info('Received webhook request', [
            'content' => $content,
        ]);

        $update = json_decode($content, true);

        if (null === $update) {
            $this->logger->error('Failed to decode webhook request content', [
                'content' => $content,
                'error' => json_last_error_msg(),
            ]);

            return new Response('Invalid request', Response::HTTP_BAD_REQUEST);
        }

        $this->telegramBotService->handleUpdate($update);

        $this->logger->info('Webhook request processed successfully');

        return new Response('OK');
    }
}
