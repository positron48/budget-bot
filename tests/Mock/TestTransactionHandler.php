<?php

namespace App\Tests\Mock;

use App\Service\TransactionHandler;

class TestTransactionHandler extends TransactionHandler
{
    private ResponseCollector $responseCollector;

    public function __construct()
    {
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
