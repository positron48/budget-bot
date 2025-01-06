<?php

namespace App\Service;

use Longman\TelegramBot\Entities\ServerResponse;
use Longman\TelegramBot\Request;
use Longman\TelegramBot\Telegram;

class TelegramApiService implements TelegramApiServiceInterface
{
    private Telegram $telegram;

    public function initialize(string $apiKey, string $botUsername): void
    {
        $this->telegram = new Telegram($apiKey, $botUsername);
        Request::initialize($this->telegram);
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

    /**
     * @param array<string> $keyboard
     */
    public function sendMessageWithKeyboard(int $chatId, string $text, array $keyboard): ServerResponse
    {
        $data = [
            'chat_id' => $chatId,
            'text' => $text,
            'parse_mode' => 'HTML',
            'reply_markup' => json_encode([
                'keyboard' => array_map(
                    static fn (string $button): array => [['text' => $button]],
                    $keyboard
                ),
                'one_time_keyboard' => true,
                'resize_keyboard' => true,
            ]),
        ];

        return $this->sendMessage($data);
    }
}
