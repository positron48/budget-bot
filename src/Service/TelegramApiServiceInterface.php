<?php

namespace App\Service;

use Longman\TelegramBot\Entities\ServerResponse;

interface TelegramApiServiceInterface
{
    public function initialize(string $apiKey, string $botUsername): void;

    /**
     * @param array{
     *     chat_id: int,
     *     text: string,
     *     parse_mode: string,
     *     reply_markup?: string|false
     * } $data
     */
    public function sendMessage(array $data): ServerResponse;
}
