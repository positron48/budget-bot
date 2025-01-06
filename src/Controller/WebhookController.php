<?php

namespace App\Controller;

use App\Service\TelegramBotService;
use Longman\TelegramBot\Entities\Update;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\Routing\Annotation\Route;

class WebhookController extends AbstractController
{
    public function __construct(
        private readonly TelegramBotService $telegramBotService,
    ) {
    }

    #[Route('/webhook', name: 'webhook', methods: ['POST'])]
    public function webhook(Request $request): JsonResponse
    {
        $update = new Update(json_decode($request->getContent(), true, 512, JSON_THROW_ON_ERROR));
        $this->telegramBotService->handleUpdate($update);

        return new JsonResponse(['status' => 'ok']);
    }
}
