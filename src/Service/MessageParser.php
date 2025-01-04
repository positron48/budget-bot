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
        if (!is_string($text)) {
            return [
                'command' => '',
                'arguments' => [],
            ];
        }

        // Split text into parts
        $parts = explode(' ', $text);

        // First part is always a command
        $command = strtolower(array_shift($parts));

        return [
            'command' => $command,
            'arguments' => $parts,
        ];
    }

    /**
     * @return array{date: \DateTime, amount: float, description: string, isIncome: bool}|null
     */
    public function parseMessage(?string $text): ?array
    {
        if (null === $text) {
            return null;
        }

        $text = trim($text);
        if ('' === $text) {
            return null;
        }

        $parts = explode(' ', $text);
        if (count($parts) < 2) {
            return null;
        }

        $date = new \DateTime();
        $amount = 0.0;
        $description = '';
        $isIncome = false;

        // Parse first part
        $firstPart = $parts[0];
        if (str_starts_with($firstPart, '+')) {
            $isIncome = true;
            $amount = (float) substr($firstPart, 1);
            $description = implode(' ', array_slice($parts, 1));
        } elseif (is_numeric($firstPart)) {
            $amount = (float) $firstPart;
            $description = implode(' ', array_slice($parts, 1));
        } else {
            try {
                $date = new \DateTime($firstPart);
                $secondPart = $parts[1];
                if (str_starts_with($secondPart, '+')) {
                    $isIncome = true;
                    $amount = (float) substr($secondPart, 1);
                    $description = implode(' ', array_slice($parts, 2));
                } else {
                    $amount = (float) $secondPart;
                    $description = implode(' ', array_slice($parts, 2));
                }
            } catch (\Exception $e) {
                return null;
            }
        }

        if ($amount <= 0) {
            return null;
        }

        return [
            'date' => $date,
            'amount' => $amount,
            'description' => $description,
            'isIncome' => $isIncome,
        ];
    }
}
