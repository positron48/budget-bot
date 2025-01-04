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
    public function parseMessage(string $message): ?array
    {
        $parts = preg_split('/\s+/', trim($message), 2);
        if (count($parts) < 2) {
            return null;
        }

        $firstPart = $parts[0];
        $remainingPart = $parts[1] ?? '';

        // Try to parse the first part as a date
        $date = $this->parseDate($firstPart);
        if (null === $date) {
            // If first part is not a date, assume it's today
            $date = new \DateTime();
            $remainingPart = $message;
        }

        // Parse amount and description
        if (!preg_match('/^([+]?\d+(?:\.\d+)?)\s*(.*)$/', $remainingPart, $matches)) {
            return null;
        }

        $amount = (float) $matches[1];
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
            $date = \DateTime::createFromFormat($format, $dateStr);
            if (false !== $date) {
                // If year is not specified, use current year
                if (false === strpos($format, 'Y')) {
                    $date->setDate((int) date('Y'), (int) $date->format('m'), (int) $date->format('d'));
                }

                return $date;
            }
        }

        return null;
    }
}
