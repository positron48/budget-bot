<?php

declare(strict_types=1);

namespace App\Utility;

class MonthUtility
{
    private const MONTHS = [
        1 => 'Январь',
        2 => 'Февраль',
        3 => 'Март',
        4 => 'Апрель',
        5 => 'Май',
        6 => 'Июнь',
        7 => 'Июль',
        8 => 'Август',
        9 => 'Сентябрь',
        10 => 'Октябрь',
        11 => 'Ноябрь',
        12 => 'Декабрь',
    ];

    public static function getMonthName(int $month): string
    {
        return self::MONTHS[$month] ?? '';
    }

    public static function getMonthNumber(string $name): ?int
    {
        $monthLower = mb_strtolower($name);
        $months = array_map(fn (string $month): string => mb_strtolower($month), self::MONTHS);
        $months = array_flip($months);

        return $months[$monthLower] ?? null;
    }
}
