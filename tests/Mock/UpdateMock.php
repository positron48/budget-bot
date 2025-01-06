<?php

namespace App\Tests\Mock;

use Longman\TelegramBot\Entities\Update;

class UpdateMock extends Update
{
    public function __construct()
    {
        $message = new MessageMock();
        parent::__construct([
            'update_id' => 1,
            'message' => $message->getRawData(),
        ]);
    }
}
