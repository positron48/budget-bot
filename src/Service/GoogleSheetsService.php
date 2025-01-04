<?php

namespace App\Service;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use App\Repository\UserSpreadsheetRepository;
use Google\Service\Drive;
use Google\Service\Sheets;
use Google\Service\Sheets\ValueRange;
use Psr\Log\LoggerInterface;

class GoogleSheetsService
{
    private const FIRST_DATA_ROW = 2;

    private Sheets $sheetsService;
    private Drive $driveService;
    private string $serviceAccountEmail;
    private LoggerInterface $logger;
    private UserSpreadsheetRepository $spreadsheetRepository;

    public function __construct(
        string $credentialsPath,
        string $serviceAccountEmail,
        LoggerInterface $logger,
        UserSpreadsheetRepository $spreadsheetRepository,
    ) {
        $this->serviceAccountEmail = $serviceAccountEmail;
        $this->logger = $logger;
        $this->spreadsheetRepository = $spreadsheetRepository;

        try {
            $client = new \Google\Client();
            $client->setAuthConfig($credentialsPath);
            $client->setScopes([
                Sheets::SPREADSHEETS,
                Drive::DRIVE_READONLY,
            ]);

            $this->sheetsService = new Sheets($client);
            $this->driveService = new Drive($client);
        } catch (\Exception $e) {
            $this->logger->error('Failed to initialize Google services: {error}', [
                'error' => $e->getMessage(),
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    public function addExpense(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category,
    ): void {
        $this->logger->info('Adding expense', [
            'spreadsheet_id' => $spreadsheetId,
            'date' => $date,
            'amount' => $amount,
            'description' => $description,
            'category' => $category,
        ]);

        try {
            $range = 'Транзакции!B:E';
            $row = $this->findFirstEmptyRow($spreadsheetId, $range);

            $values = [
                [
                    $date,
                    $amount,
                    $description,
                    $category,
                ],
            ];

            $body = new ValueRange([
                'values' => $values,
            ]);

            $this->sheetsService->spreadsheets_values->update(
                $spreadsheetId,
                $range,
                $body,
                ['valueInputOption' => 'RAW']
            );

            $this->logger->info('Expense added successfully', [
                'spreadsheet_id' => $spreadsheetId,
                'row' => $row,
            ]);
        } catch (\Exception $e) {
            $this->logger->error('Failed to add expense: {error}', [
                'error' => $e->getMessage(),
                'spreadsheet_id' => $spreadsheetId,
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    public function addIncome(
        string $spreadsheetId,
        string $date,
        float $amount,
        string $description,
        string $category,
    ): void {
        $this->logger->info('Adding income', [
            'spreadsheet_id' => $spreadsheetId,
            'date' => $date,
            'amount' => $amount,
            'description' => $description,
            'category' => $category,
        ]);

        try {
            $range = 'Транзакции!G:J';
            $row = $this->findFirstEmptyRow($spreadsheetId, $range);

            $values = [
                [
                    $date,
                    $amount,
                    $description,
                    $category,
                ],
            ];

            $body = new ValueRange([
                'values' => $values,
            ]);

            $this->sheetsService->spreadsheets_values->update(
                $spreadsheetId,
                $range,
                $body,
                ['valueInputOption' => 'RAW']
            );

            $this->logger->info('Income added successfully', [
                'spreadsheet_id' => $spreadsheetId,
                'row' => $row,
            ]);
        } catch (\Exception $e) {
            $this->logger->error('Failed to add income: {error}', [
                'error' => $e->getMessage(),
                'spreadsheet_id' => $spreadsheetId,
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    private function findFirstEmptyRow(string $spreadsheetId, string $range): int
    {
        try {
            $response = $this->sheetsService->spreadsheets_values->get($spreadsheetId, $range);
            $values = $response->getValues();

            return count($values ?? []) + self::FIRST_DATA_ROW;
        } catch (\Exception $e) {
            $this->logger->error('Failed to find first empty row: {error}', [
                'error' => $e->getMessage(),
                'spreadsheet_id' => $spreadsheetId,
                'range' => $range,
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    public function validateSpreadsheetAccess(string $spreadsheetId): bool
    {
        try {
            $this->sheetsService->spreadsheets->get($spreadsheetId);
            $this->logger->info('Spreadsheet access validated', [
                'spreadsheet_id' => $spreadsheetId,
            ]);

            return true;
        } catch (\Exception $e) {
            $this->logger->warning('Failed to validate spreadsheet access: {error}', [
                'error' => $e->getMessage(),
                'spreadsheet_id' => $spreadsheetId,
            ]);

            return false;
        }
    }

    public function getSpreadsheetTitle(string $spreadsheetId): ?string
    {
        try {
            $spreadsheet = $this->sheetsService->spreadsheets->get($spreadsheetId);
            $title = $spreadsheet->getProperties()->getTitle();
            $this->logger->info('Retrieved spreadsheet title', [
                'spreadsheet_id' => $spreadsheetId,
                'title' => $title,
            ]);

            return $title;
        } catch (\Exception $e) {
            $this->logger->error('Failed to get spreadsheet title: {error}', [
                'error' => $e->getMessage(),
                'spreadsheet_id' => $spreadsheetId,
            ]);

            return null;
        }
    }

    public function getSharingInstructions(string $spreadsheetId): string
    {
        $this->logger->info('Generated sharing instructions', [
            'spreadsheet_id' => $spreadsheetId,
            'service_account' => $this->serviceAccountEmail,
        ]);

        return "Для работы с таблицей нужно предоставить доступ сервисному аккаунту:\n\n".
               "1. Откройте таблицу\n".
               "2. Нажмите кнопку \"Настройки доступа\" (или \"Share\")\n".
               "3. В поле \"Добавить пользователей или группы\" введите:\n".
               "{$this->serviceAccountEmail}\n".
               "4. Выберите роль \"Редактор\"\n".
               "5. Нажмите \"Готово\"\n\n".
               'После этого отправьте команду /add еще раз';
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

    public function addSpreadsheet(User $user, string $spreadsheetId, int $month, int $year): void
    {
        $title = $this->getSpreadsheetTitle($spreadsheetId);
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

    public function cloneSpreadsheet(string $sourceId, string $newTitle): string
    {
        try {
            $copy = $this->driveService->files->copy(
                $sourceId,
                new Drive\DriveFile(['name' => $newTitle])
            );

            $this->logger->info('Spreadsheet cloned', [
                'source_id' => $sourceId,
                'new_id' => $copy->getId(),
                'new_title' => $newTitle,
            ]);

            return $copy->getId();
        } catch (\Exception $e) {
            $this->logger->error('Failed to clone spreadsheet: {error}', [
                'error' => $e->getMessage(),
                'source_id' => $sourceId,
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    public function findSpreadsheetByDate(User $user, \DateTime $date): ?UserSpreadsheet
    {
        return $this->spreadsheetRepository->findByDate($user, $date);
    }

    public function findLatestSpreadsheet(User $user): ?UserSpreadsheet
    {
        return $this->spreadsheetRepository->findLatest($user);
    }

    public function getSpreadsheetUrl(string $spreadsheetId): string
    {
        return sprintf('https://docs.google.com/spreadsheets/d/%s', $spreadsheetId);
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
        if (!$this->validateSpreadsheetAccess($input)) {
            throw new \RuntimeException($this->getSharingInstructions($input));
        }

        return $input;
    }
}
