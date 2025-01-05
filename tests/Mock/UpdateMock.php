<?php

namespace App\Tests\Mock;

use Longman\TelegramBot\Entities\Update;

class UpdateMock extends Update
{
    public int $update_id;
    /** @var array<string, mixed> */
    public array $raw_data = [];
    public string $bot_username = '';
    public ?MessageMock $message = null;

    /**
     * @param array<string, mixed> $data
     */
    public function __construct(array $data)
    {
        $this->raw_data = $data;
        $this->update_id = $data['update_id'];
        $this->bot_username = $data['bot_username'] ?? '';

        if (isset($data['message'])) {
            $this->message = new MessageMock($data['message']);
        }
    }
}
