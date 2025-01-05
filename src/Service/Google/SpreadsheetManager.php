<?php

namespace App\Service\Google;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use App\Repository\UserSpreadsheetRepository;
use Psr\Log\LoggerInterface;

class SpreadsheetManager
{
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
