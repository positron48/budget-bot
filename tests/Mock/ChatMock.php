<?php

namespace App\Tests\Mock;

use Longman\TelegramBot\Entities\Chat;

class ChatMock extends Chat
{
    public function __construct()
    {
        parent::__construct(['id' => 123456, 'type' => 'private']);
    }
}
