<?php

namespace App\Tests\Mock;

use Longman\TelegramBot\Entities\Message;

class MessageMock extends Message
{
    public int $message_id;
    /** @var array<string, mixed> */
    public array $raw_data = [];
    public string $bot_username = '';
    public ?ChatMock $chat = null;
    public ?string $text = null;

    /**
     * @param array<string, mixed> $data
     */
    public function __construct(array $data)
    {
        $this->raw_data = $data;
        $this->message_id = $data['message_id'];
        $this->bot_username = $data['bot_username'] ?? '';
        $this->text = $data['text'] ?? null;

        if (isset($data['chat'])) {
            $this->chat = new ChatMock($data['chat']);
        }
    }
}
