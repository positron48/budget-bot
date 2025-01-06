<?php

namespace App\Service;

use Longman\TelegramBot\Entities\ServerResponse;
use Longman\TelegramBot\Request;
use Longman\TelegramBot\Telegram;

class TelegramApiService implements TelegramApiServiceInterface
{
    public function initialize(string $apiKey, string $botUsername): void
    {
        $telegram = new Telegram($apiKey, $botUsername);
        Request::initialize($telegram);
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
        return Request::sendMessage($data);
    }
}
