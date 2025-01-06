<?php

namespace App\Tests\Mock;

use App\Service\TelegramApiServiceInterface;
use Longman\TelegramBot\Entities\ServerResponse;

class TelegramApiMock implements TelegramApiServiceInterface
{
    /** @var array<int|string, mixed> */
    private array $messages = [];

    public function initialize(string $apiKey, string $botUsername): void
    {
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
        $this->messages[] = $data;

        return new ServerResponseMock();
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

    /**
     * @return array<int|string, mixed>
     */
    public function getMessages(): array
    {
        return $this->messages;
    }
}
