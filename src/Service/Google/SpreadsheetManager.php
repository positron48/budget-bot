<?php

namespace App\Service\Google;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use App\Repository\UserSpreadsheetRepository;
use Psr\Log\LoggerInterface;

class SpreadsheetManager
{
    private const EXPENSE_CATEGORIES_COLUMN = 'Сводка!B28:B';
    private const INCOME_CATEGORIES_COLUMN = 'Сводка!H28:H';
    private const EXPENSE_CATEGORIES_ROW_TEMPLATE = 'Сводка!B%d:F%d';
    private const INCOME_CATEGORIES_ROW_TEMPLATE = 'Сводка!H%d:L%d';

    private GoogleSheetsClient $client;
    private UserSpreadsheetRepository $spreadsheetRepository;
    private LoggerInterface $logger;

    public function __construct(
        GoogleSheetsClient $client,
        UserSpreadsheetRepository $spreadsheetRepository,
        LoggerInterface $logger,
    ) {
        $this->client = $client;
        $this->spreadsheetRepository = $spreadsheetRepository;
        $this->logger = $logger;
    }

    public function addSpreadsheet(User $user, string $spreadsheetId, int $month, int $year): void
    {
        $title = $this->client->getSpreadsheetTitle($spreadsheetId);
        if (!$title) {
            throw new \RuntimeException('Failed to get spreadsheet title');
        }

        // Check if spreadsheet already exists for this month and year
        $existing = $this->spreadsheetRepository->findByMonthAndYear($user, $month, $year);
        if ($existing) {
            throw new \RuntimeException('Таблица для этого месяца и года уже существует');
        }

        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setUser($user)
            ->setSpreadsheetId($spreadsheetId)
            ->setTitle($title)
            ->setMonth($month)
            ->setYear($year);

        $this->spreadsheetRepository->save($spreadsheet, true);

        $this->logger->info('Spreadsheet added for user', [
            'user_id' => $user->getId(),
            'spreadsheet_id' => $spreadsheetId,
            'title' => $title,
            'month' => $month,
            'year' => $year,
        ]);
    }

    public function removeSpreadsheet(User $user, int $month, int $year): void
    {
        $spreadsheet = $this->spreadsheetRepository->findByMonthAndYear($user, $month, $year);
        if (!$spreadsheet) {
            throw new \RuntimeException(sprintf('Таблица за %s %d не найдена', $this->getMonthName($month), $year));
        }

        $this->logger->info('Removing spreadsheet {spreadsheet_id} for user {telegram_id}', [
            'spreadsheet_id' => $spreadsheet->getSpreadsheetId(),
            'telegram_id' => $user->getTelegramId(),
            'month' => $month,
            'year' => $year,
        ]);

        $this->spreadsheetRepository->remove($spreadsheet, true);
    }

    public function findSpreadsheetByDate(User $user, \DateTime $date): ?UserSpreadsheet
    {
        return $this->spreadsheetRepository->findByDate($user, $date);
    }

    public function findLatestSpreadsheet(User $user): ?UserSpreadsheet
    {
        return $this->spreadsheetRepository->findLatest($user);
    }

    /**
     * @return array<int, array{month: string, year: int, url: string}>
     */
    public function getSpreadsheetsList(User $user): array
    {
        $spreadsheets = $this->spreadsheetRepository->findBy(['user' => $user], ['year' => 'DESC', 'month' => 'DESC']);
        $result = [];

        foreach ($spreadsheets as $spreadsheet) {
            $month = $spreadsheet->getMonth();
            $year = $spreadsheet->getYear();
            $id = $spreadsheet->getSpreadsheetId();

            if (null === $month || null === $year || null === $id) {
                continue;
            }

            $result[] = [
                'month' => $this->getMonthName($month),
                'year' => $year,
                'url' => $this->getSpreadsheetUrl($id),
            ];
        }

        return $result;
    }

    public function handleSpreadsheetId(string $input): string
    {
        // Extract ID from URL if needed
        if (str_contains($input, 'docs.google.com/spreadsheets/d/')) {
            if (preg_match('/spreadsheets\/d\/([a-zA-Z0-9-_]+)/', $input, $matches)) {
                return $matches[1];
            }
            throw new \RuntimeException('Неверный формат ссылки. Пожалуйста, убедитесь, что вы скопировали полную ссылку на таблицу.');
        }

        // Validate access to the spreadsheet
        if (!$this->client->validateSpreadsheetAccess($input)) {
            throw new \RuntimeException($this->client->getSharingInstructions($input));
        }

        return $input;
    }

    /**
     * @return array<string>
     */
    public function getExpenseCategories(string $spreadsheetId): array
    {
        $values = $this->client->getValues($spreadsheetId, self::EXPENSE_CATEGORIES_COLUMN);
        if (!$values) {
            return [];
        }

        // Get all non-empty values from the column
        $categories = array_filter(array_column($values, 0), static fn ($value): bool => !empty($value) && is_string($value));

        return array_values(array_unique($categories));
    }

    /**
     * @return array<string>
     */
    public function getIncomeCategories(string $spreadsheetId): array
    {
        $values = $this->client->getValues($spreadsheetId, self::INCOME_CATEGORIES_COLUMN);
        if (!$values) {
            return [];
        }

        // Get all non-empty values from the column
        $categories = array_filter(array_column($values, 0), static fn ($value): bool => !empty($value) && is_string($value));

        return array_values(array_unique($categories));
    }

    public function addExpenseCategory(string $spreadsheetId, string $category): void
    {
        $categories = $this->getExpenseCategories($spreadsheetId);
        if (in_array($category, $categories, true)) {
            return;
        }

        // Find the first empty cell in the column
        $values = $this->client->getValues($spreadsheetId, self::EXPENSE_CATEGORIES_COLUMN);
        if (!$values) {
            // If the range is empty, add to the first cell
            $this->client->updateValues($spreadsheetId, self::EXPENSE_CATEGORIES_COLUMN, [[$category]]);

            return;
        }

        // Find the first empty row
        $rowIndex = 28; // Start from row 28
        foreach ($values as $value) {
            if (empty($value[0])) {
                break;
            }
            ++$rowIndex;
        }

        // Add the category to the row and copy the formula
        $range = sprintf(self::EXPENSE_CATEGORIES_ROW_TEMPLATE, $rowIndex, $rowIndex);
        $this->client->updateValues($spreadsheetId, $range, [[$category, '', '', '', '']]);
    }

    public function addIncomeCategory(string $spreadsheetId, string $category): void
    {
        $categories = $this->getIncomeCategories($spreadsheetId);
        if (in_array($category, $categories, true)) {
            return;
        }

        // Find the first empty cell in the column
        $values = $this->client->getValues($spreadsheetId, self::INCOME_CATEGORIES_COLUMN);
        if (!$values) {
            // If the range is empty, add to the first cell
            $this->client->updateValues($spreadsheetId, self::INCOME_CATEGORIES_COLUMN, [[$category]]);

            return;
        }

        // Find the first empty row
        $rowIndex = 28; // Start from row 28
        foreach ($values as $value) {
            if (empty($value[0])) {
                break;
            }
            ++$rowIndex;
        }

        // Add the category to the row and copy the formula
        $range = sprintf(self::INCOME_CATEGORIES_ROW_TEMPLATE, $rowIndex, $rowIndex);
        $this->client->updateValues($spreadsheetId, $range, [[$category, '', '', '', '']]);
    }

    private function getMonthName(int $month): string
    {
        $months = [
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

        return $months[$month] ?? '';
    }

    private function getSpreadsheetUrl(string $spreadsheetId): string
    {
        return sprintf('https://docs.google.com/spreadsheets/d/%s', $spreadsheetId);
    }
}
