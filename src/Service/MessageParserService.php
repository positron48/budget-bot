<?php

namespace App\Service;

class MessageParserService
{
    private const DATE_FORMATS = [
        'd.m.Y',
        'd.m',
        'd/m/Y',
        'd/m',
    ];

    /**
     * @return array{date: \DateTime, amount: float, description: string, isIncome: bool}|null
     */
    public function parseMessage(string $text): ?array
    {
        $parts = preg_split('/\s+/', trim($text));
        if (!is_array($parts) || count($parts) < 2) {
            return null;
        }

        $firstPart = $parts[0];
        $remainingPart = implode(' ', array_slice($parts, 1));

        // Try to parse the first part as a date
        $date = $this->parseDate($firstPart);
        if (null === $date) {
            // If first part is not a date, assume it's today and it's part of the amount
            $date = new \DateTime();
            $remainingPart = $text;
        }

        // Parse amount and description
        $remainingPart = str_replace(',', '.', trim($remainingPart));

        // Try to parse amount and description
        if (!preg_match('/^([+]?\d+(?:\.\d+)?)\s+(.+)$/', $remainingPart, $matches)) {
            // Try simpler format without decimal places
            if (!preg_match('/^([+]?\d+)\s+(.+)$/', $remainingPart, $matches)) {
                // Try parsing just the first part as amount
                if (!preg_match('/^([+]?\d+)$/', $firstPart, $matches)) {
                    return null;
                }
                $description = implode(' ', array_slice($parts, 1));
            } else {
                $description = $matches[2];
            }
        } else {
            $description = $matches[2];
        }

        $amount = (float) $matches[1];
        if ($amount <= 0) {
            return null;
        }

        $description = trim($description);
        $isIncome = str_starts_with($matches[1], '+');

        return [
            'date' => $date,
            'amount' => $amount,
            'description' => $description,
            'isIncome' => $isIncome,
        ];
    }

    private function parseDate(string $dateStr): ?\DateTime
    {
        $dateStr = mb_strtolower($dateStr);

        // Handle special cases
        if ('сегодня' === $dateStr) {
            return new \DateTime();
        }
        if ('вчера' === $dateStr) {
            return new \DateTime('-1 day');
        }

        // Try different date formats
        foreach (self::DATE_FORMATS as $format) {
            if (preg_match('/^(\d{1,2})\.(\d{1,2})(?:\.(\d{4}))?$/', $dateStr, $matches)) {
                $day = (int) $matches[1];
                $month = (int) $matches[2];
                $year = isset($matches[3]) ? (int) $matches[3] : (int) date('Y');

                // Validate year format
                if (isset($matches[3]) && ($year < 1000 || $year > 9999)) {
                    continue;
                }

                // Validate date
                if (!checkdate($month, $day, $year)) {
                    continue;
                }

                $date = new \DateTime();
                $date->setDate($year, $month, $day);

                return $date;
            }
        }

        return null;
    }
}
