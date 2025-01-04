<?php

namespace App\Controller;

use App\Service\TelegramBotService;
use Psr\Log\LoggerInterface;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class WebhookController extends AbstractController
{
    private TelegramBotService $botService;
    private LoggerInterface $logger;

    public function __construct(TelegramBotService $botService, LoggerInterface $logger)
    {
        $this->botService = $botService;
        $this->logger = $logger;
    }

    #[Route('/webhook', name: 'telegram_webhook', methods: ['POST'])]
    public function webhook(Request $request): Response
    {
        $this->logger->info('Received webhook request', [
            'content' => $request->getContent(),
            'headers' => $request->headers->all(),
        ]);

        $update = json_decode($request->getContent(), true);

        if (!$update) {
            $this->logger->error('Invalid request: unable to decode JSON');

            return new Response('Invalid request', Response::HTTP_BAD_REQUEST);
        }

        try {
            $this->botService->handleUpdate($update);
            $this->logger->info('Update processed successfully');

            return new Response('OK');
        } catch (\Exception $e) {
            $this->logger->error('Error processing update: '.$e->getMessage(), [
                'exception' => $e,
                'update' => $update,
            ]);

            return new Response('Error processing update', Response::HTTP_INTERNAL_SERVER_ERROR);
        }
    }
}
