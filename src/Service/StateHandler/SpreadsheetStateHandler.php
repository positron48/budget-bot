<?php

namespace App\Service\StateHandler;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\GoogleSheetsService;
use Longman\TelegramBot\Request;
use Psr\Log\LoggerInterface;

class SpreadsheetStateHandler implements StateHandlerInterface
{
    private const SUPPORTED_STATES = [
        'WAITING_SPREADSHEET_ID',
        'WAITING_MONTH',
    ];

    private UserRepository $userRepository;
    private GoogleSheetsService $sheetsService;
    private LoggerInterface $logger;

    public function __construct(
        UserRepository $userRepository,
        GoogleSheetsService $sheetsService,
        LoggerInterface $logger
    ) {
        $this->userRepository = $userRepository;
        $this->sheetsService = $sheetsService;
        $this->logger = $logger;
    }

    public function supports(string $state): bool
    {
        return in_array($state, self::SUPPORTED_STATES, true);
    }

    public function handle(int $chatId, User $user, string $message): void
    {
        $state = $user->getState();

        if ($state === 'WAITING_SPREADSHEET_ID') {
            $this->handleSpreadsheetId($chatId, $user, $message);
            return;
        }

        if ($state === 'WAITING_MONTH') {
            $this->handleMonthSelection($chatId, $user, $message);
        }
    }

    private function handleSpreadsheetId(int $chatId, User $user, string $spreadsheetId): void
    {
        try {
            $this->logger->info('Validating spreadsheet access', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
            ]);
            $spreadsheetId = $this->sheetsService->handleSpreadsheetId($spreadsheetId);

            $this->logger->info('Setting user state to WAITING_MONTH', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
            ]);

            $user->setTempData(['spreadsheet_id' => $spreadsheetId]);
            $this->userRepository->save($user, true);
            $this->userRepository->setUserState($user, 'WAITING_MONTH');

            $keyboard = $this->buildMonthsKeyboard();

            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Выберите месяц из списка или введите в формате "Месяц Год" (например: Январь 2025):',
                'reply_markup' => json_encode([
                    'keyboard' => $keyboard,
                    'one_time_keyboard' => true,
                    'resize_keyboard' => true,
                ]),
            ]);
        } catch (\RuntimeException $e) {
            $this->logger->warning('Failed to handle spreadsheet', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
                'error' => $e->getMessage(),
            ]);
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => $e->getMessage(),
            ]);
            $this->userRepository->clearUserState($user);
        }
    }

    private function handleMonthSelection(int $chatId, User $user, string $text): void
    {
        $tempData = $user->getTempData();
        $spreadsheetId = $tempData['spreadsheet_id'] ?? null;

        if (!$spreadsheetId) {
            $this->logger->error('Missing spreadsheet_id in temp data', [
                'chat_id' => $chatId,
                'temp_data' => $tempData,
            ]);
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Произошла ошибка. Попробуйте начать сначала с команды /add',
            ]);
            $this->userRepository->clearUserState($user);
            return;
        }

        try {
            [$monthNumber, $year] = $this->parseMonthAndYear($text);

            $this->logger->info('Adding spreadsheet for month', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
                'month' => $monthNumber,
                'year' => $year,
            ]);

            $this->sheetsService->addSpreadsheet($user, $spreadsheetId, $monthNumber, $year);

            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf(
                    'Таблица успешно привязана к месяцу %s %d',
                    $this->getMonthName($monthNumber),
                    $year
                ),
            ]);

            $this->logger->info('Successfully added spreadsheet', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
                'month' => $monthNumber,
                'year' => $year,
            ]);
        } catch (\RuntimeException $e) {
            $this->logger->error('Failed to add spreadsheet', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
                'text' => $text,
                'error' => $e->getMessage(),
            ]);
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => $e->getMessage(),
            ]);
        }

        $this->userRepository->clearUserState($user);
    }

    /**
     * @return array<int, array<array<string, string>>>
     */
    private function buildMonthsKeyboard(): array
    {
        $months = $this->getMonthsList();
        $keyboard = [];
        $row = [];
        $i = 0;

        foreach ($months as $label => $month) {
            $row[] = ['text' => $label];
            ++$i;
            if (0 === $i % 3) {
                $keyboard[] = $row;
                $row = [];
            }
        }

        if (!empty($row)) {
            $keyboard[] = $row;
        }

        return $keyboard;
    }

    /**
     * @return array{0: int, 1: int}
     */
    private function parseMonthAndYear(string $text): array
    {
        // Try to parse month and year from text
        if (preg_match('/^(\p{L}+)\s+(\d{4})$/u', $text, $matches)) {
            $monthName = $matches[1];
            $year = (int) $matches[2];

            $months = $this->getMonthsMap();
            if (!isset($months[$monthName])) {
                throw new \RuntimeException('Пожалуйста, используйте корректное название месяца (например: Январь 2025)');
            }

            return [$months[$monthName], $year];
        }

        // Try to parse from the keyboard selection
        $months = $this->getMonthsList();
        if (!isset($months[$text])) {
            throw new \RuntimeException('Пожалуйста, выберите месяц из предложенных вариантов или введите в формате "Месяц Год" (например: Январь 2025)');
        }

        if (!preg_match('/^.*\s+(\d{4})$/', $text, $matches)) {
            throw new \RuntimeException('Произошла ошибка при обработке выбранного месяца. Попробуйте ввести месяц и год вручную (например: Январь 2025)');
        }

        return [$months[$text], (int) $matches[1]];
    }

    /**
     * @return array<string, int>
     */
    private function getMonthsList(): array
    {
        $months = $this->getMonthsMap();
        $currentMonth = (int) date('m');
        $currentYear = (int) date('Y');
        $result = [];

        // Get next month and 5 previous months
        $nextMonth = $currentMonth + 1;
        $nextMonthYear = $currentYear;
        if ($nextMonth > 12) {
            $nextMonth = 1;
            ++$nextMonthYear;
        }

        // Add next month first
        $monthName = array_search($nextMonth, $months, true);
        if ($monthName) {
            $result[$monthName.' '.$nextMonthYear] = $nextMonth;
        }

        // Then add 5 previous months
        for ($i = 0; $i < 5; ++$i) {
            $month = $nextMonth - 1 - $i;
            $year = $nextMonthYear;

            if ($month <= 0) {
                $month += 12;
                --$year;
            }

            $monthName = array_search($month, $months, true);
            if ($monthName) {
                $result[$monthName.' '.$year] = $month;
            }
        }

        return $result;
    }

    /**
     * @return array<string, int>
     */
    private function getMonthsMap(): array
    {
        return [
            'Январь' => 1,
            'Февраль' => 2,
            'Март' => 3,
            'Апрель' => 4,
            'Май' => 5,
            'Июнь' => 6,
            'Июль' => 7,
            'Август' => 8,
            'Сентябрь' => 9,
            'Октябрь' => 10,
            'Ноябрь' => 11,
            'Декабрь' => 12,
        ];
    }

    private function getMonthName(int $month): string
    {
        $months = array_flip($this->getMonthsMap());
        return $months[$month] ?? '';
    }
} 