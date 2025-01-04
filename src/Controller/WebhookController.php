<?php

namespace App\Controller;

use App\Service\TelegramBotService;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class WebhookController extends AbstractController
{
    private TelegramBotService $botService;

    public function __construct(TelegramBotService $botService)
    {
        $this->botService = $botService;
    }

    #[Route('/webhook', name: 'telegram_webhook', methods: ['POST'])]
    public function webhook(Request $request): Response
    {
        $update = json_decode($request->getContent(), true);

        if (!$update) {
            return new Response('Invalid request', Response::HTTP_BAD_REQUEST);
        }

        try {
            $this->botService->handleUpdate($update);

            return new Response('OK');
        } catch (\Exception $e) {
            return new Response('Error processing update', Response::HTTP_INTERNAL_SERVER_ERROR);
        }
    }
}
