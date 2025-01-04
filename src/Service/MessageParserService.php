<?php

namespace App\Service;

use DateTime;
use Exception;

class MessageParserService
{
    private const DATE_FORMATS = [
        'd.m.Y',
        'd.m',
        'Y-m-d',
        'd F',
        'd F Y',
    ];

    public function parseMessage(string $message): ?array
    {
        if (empty($message)) {
            return null;
        }

        $parts = preg_split('/\s+/', trim($message), 2);
        if (count($parts) < 2) {
            return null;
        }

        $firstPart = $parts[0];
        $remainingPart = $parts[1];

        // Try to parse the first part as a date
        $date = $this->parseDate($firstPart);
        if ($date === null) {
            // If first part is not a date, assume it's today
            $date = new DateTime();
            $remainingPart = $message;
        }

        // Parse amount and description
        if (!preg_match('/^([+]?\d+(?:[.,]\d{1,2})?)\s+(.+)$/', $remainingPart, $matches)) {
            return null;
        }

        $amount = (float) str_replace(',', '.', $matches[1]);
        $description = trim($matches[2]);
        $isIncome = str_starts_with($matches[1], '+');

        return [
            'date' => $date,
            'amount' => abs($amount),
            'description' => $description,
            'isIncome' => $isIncome,
        ];
    }

    private function parseDate(string $dateStr): ?DateTime
    {
        $dateStr = mb_strtolower($dateStr);

        // Handle special cases
        if ($dateStr === 'сегодня') {
            return new DateTime();
        }
        if ($dateStr === 'вчера') {
            return new DateTime('-1 day');
        }

        // Try different date formats
        foreach (self::DATE_FORMATS as $format) {
            $date = DateTime::createFromFormat($format, $dateStr);
            if ($date !== false) {
                // If year is not specified, use current year
                if (strpos($format, 'Y') === false) {
                    $date->setDate((int)date('Y'), (int)$date->format('m'), (int)$date->format('d'));
                }
                return $date;
            }
        }

        return null;
    }
} 