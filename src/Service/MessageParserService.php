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
        $remainingPart = $parts[1] ?? '';

        // Try to parse the first part as a date
        $date = $this->parseDate($firstPart);
        if (null === $date) {
            // If first part is not a date, assume it's today
            $date = new \DateTime();
            $remainingPart = $text;
        } else {
            $remainingPart = trim(substr($text, strlen($firstPart)));
        }

        // Parse amount and description
        $remainingPart = str_replace(',', '.', trim($remainingPart));
        if (!preg_match('/^([+]?\d+(?:\.\d+)?)\s+(.+)$/', $remainingPart, $matches)) {
            return null;
        }

        $amount = (float) $matches[1];
        if ($amount <= 0) {
            return null;
        }

        $description = trim($matches[2]);
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
