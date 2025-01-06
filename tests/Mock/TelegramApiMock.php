<?php

namespace App\Tests\Mock;

use App\Service\TelegramApiServiceInterface;
use Longman\TelegramBot\Entities\ServerResponse;

class TelegramApiMock implements TelegramApiServiceInterface
{
    public function initialize(string $apiKey, string $botUsername): void
    {
        // Do nothing in tests
    }

    /**
     * @param array{
     *     chat_id: int,
     *     text: string,
     *     parse_mode: string,
     *     reply_markup?: string|false
     * } $data
     */
    public function sendMessage(array $data): ServerResponse
    {
        return new ServerResponse(['ok' => true]);
    }
}
