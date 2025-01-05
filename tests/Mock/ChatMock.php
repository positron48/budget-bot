<?php

namespace App\Tests\Mock;

use Longman\TelegramBot\Entities\Chat;

class ChatMock extends Chat
{
    public int $id;
    /** @var array<string, mixed> */
    public array $raw_data = [];
    public string $bot_username = '';
    public string $type = 'private';

    /**
     * @param array<string, mixed> $data
     */
    public function __construct(array $data)
    {
        $this->raw_data = $data;
        $this->id = $data['id'];
        $this->bot_username = $data['bot_username'] ?? '';
        $this->type = $data['type'] ?? 'private';
    }
}
