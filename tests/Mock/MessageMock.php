<?php

namespace App\Tests\Mock;

use Longman\TelegramBot\Entities\Message;

class MessageMock extends Message
{
    public function __construct()
    {
        $chat = new ChatMock();
        parent::__construct([
            'message_id' => 1,
            'from' => ['id' => 123456, 'first_name' => 'Test', 'is_bot' => false],
            'chat' => $chat->getRawData(),
            'date' => time(),
            'text' => '/start',
        ]);
    }
}
