<?php

namespace App\Tests\Mock;

use App\Repository\UserRepository;
use App\Service\CategoryService;
use App\Service\GoogleSheetsService;
use App\Service\TransactionHandler;
use Psr\Log\LoggerInterface;

class TestTransactionHandler extends TransactionHandler
{
    private ResponseCollector $responseCollector;

    public function __construct(
        GoogleSheetsService $sheetsService,
        CategoryService $categoryService,
        LoggerInterface $logger,
        UserRepository $userRepository,
    ) {
        parent::__construct($sheetsService, $categoryService, $logger, $userRepository);
        $this->responseCollector = ResponseCollector::getInstance();
    }

    protected function sendMessage(int $chatId, string $text): void
    {
        $data = [
            'chat_id' => $chatId,
            'text' => $text,
            'parse_mode' => 'HTML',
        ];

        error_log('Telegram API Request: '.json_encode($data, JSON_UNESCAPED_UNICODE));
        $response = ['ok' => true, 'result' => null, 'description' => null, 'error_code' => null];
        error_log('Telegram API Response: '.json_encode($response, JSON_UNESCAPED_UNICODE));

        $this->responseCollector->addResponse($text);
    }
}
