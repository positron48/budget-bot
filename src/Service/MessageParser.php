<?php

declare(strict_types=1);

namespace App\Service;

use Longman\TelegramBot\Entities\Message;

class MessageParser
{
    /**
     * @return array{command: string, arguments: string[]}
     */
    public function parse(Message $message): array
    {
        $text = $message->getText();
        if (empty($text)) {
            return [
                'command' => '',
                'arguments' => [],
            ];
        }

        // Remove multiple spaces and trim
        $text = preg_replace('/\s+/', ' ', trim($text));

        // Split text into parts
        $parts = explode(' ', $text);

        // First part is always a command
        $command = strtolower(array_shift($parts));

        return [
            'command' => $command,
            'arguments' => $parts,
        ];
    }
}
