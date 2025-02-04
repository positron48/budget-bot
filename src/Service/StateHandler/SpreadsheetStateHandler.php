<?php

namespace App\Service\StateHandler;

use App\Entity\Message;
use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\GoogleSheetsService;
use App\Service\TelegramApiServiceInterface;
use App\Utility\DateTimeUtility;
use Psr\Log\LoggerInterface;

class SpreadsheetStateHandler implements StateHandlerInterface
{
    private const SUPPORTED_STATES = [
        'WAITING_SPREADSHEET_ACTION',
        'WAITING_SPREADSHEET_ID',
        'WAITING_SPREADSHEET_MONTH',
        'WAITING_SPREADSHEET_TO_DELETE',
        'WAITING_REMOVE_SPREADSHEET',
    ];

    public function __construct(
        protected UserRepository $userRepository,
        protected GoogleSheetsService $sheetsService,
        protected LoggerInterface $logger,
        protected TelegramApiServiceInterface $telegramApi,
        protected DateTimeUtility $dateTimeUtility,
    ) {
    }

    public function supports(string $state): bool
    {
        return in_array($state, self::SUPPORTED_STATES, true);
    }

    public function handle(int $chatId, User $user, string $message): bool
    {
        $state = $user->getState();
        $tempData = $user->getTempData();

        if ('WAITING_SPREADSHEET_ACTION' === $state) {
            $this->handleSpreadsheetAction($chatId, $user, $message);

            return true;
        }

        if ('WAITING_SPREADSHEET_ID' === $state) {
            $this->handleSpreadsheetId($chatId, $user, $message);

            return true;
        }

        if ('WAITING_SPREADSHEET_MONTH' === $state && isset($tempData['spreadsheet_id'])) {
            $this->handleSpreadsheetMonth($chatId, $user, $message);

            return true;
        }

        if ('WAITING_SPREADSHEET_TO_DELETE' === $state) {
            $this->handleSpreadsheetToDelete($chatId, $user, $message);

            return true;
        }

        if ('WAITING_REMOVE_SPREADSHEET' === $state) {
            $this->handleSpreadsheetToDelete($chatId, $user, $message);

            return true;
        }

        return false;
    }

    private function handleSpreadsheetAction(int $chatId, User $user, string $message): void
    {
        switch ($message) {
            case 'Добавить таблицу':
                $user->setState('WAITING_SPREADSHEET_ID');
                $this->userRepository->save($user, true);
                $this->sendMessage($chatId, 'Введите ID таблицы:');
                break;
            case 'Удалить таблицу':
                $user->setState('WAITING_SPREADSHEET_TO_DELETE');
                $this->userRepository->save($user, true);

                $spreadsheets = $this->sheetsService->getSpreadsheetsList($user);
                if (empty($spreadsheets)) {
                    $this->sendMessage($chatId, 'У вас нет добавленных таблиц');

                    return;
                }

                $keyboard = [];
                foreach ($spreadsheets as $spreadsheet) {
                    $keyboard[] = sprintf(
                        '%s %d',
                        $spreadsheet['month'],
                        $spreadsheet['year']
                    );
                }

                $this->sendMessage($chatId, 'Выберите таблицу для удаления:', $keyboard);
                break;
            default:
                $this->sendMessage($chatId, 'Неизвестное действие');
        }
    }

    private function handleSpreadsheetId(int $chatId, User $user, string $message): void
    {
        $spreadsheetId = $message;

        try {
            $spreadsheetId = $this->sheetsService->handleSpreadsheetId($spreadsheetId);
        } catch (\Exception $e) {
            $this->logger->warning('Invalid spreadsheet ID: '.$e->getMessage(), [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
            ]);
            $this->sendMessage($chatId, 'Неверный ID таблицы. Попробуйте еще раз:');

            return;
        }

        $user->setTempData(['spreadsheet_id' => $spreadsheetId]);
        $user->setState('WAITING_SPREADSHEET_MONTH');
        $this->userRepository->save($user, true);

        $keyboard = [];
        $now = $this->dateTimeUtility->getCurrentDate();
        // Get next month
        $nextMonth = (int) $now->modify('first day of next month')->format('n');
        $nextMonthYear = (int) $now->format('Y');

        // Add next month first
        $keyboard[] = sprintf('%s %d', $this->getMonthName($nextMonth), $nextMonthYear);

        // Reset to current month
        $now = $this->dateTimeUtility->getCurrentDate();

        // Add 5 previous months
        for ($i = 0; $i < 5; ++$i) {
            $month = (int) $now->format('n');
            $year = (int) $now->format('Y');
            $keyboard[] = sprintf('%s %d', $this->getMonthName($month), $year);
            $now->modify('-1 month');
        }

        $this->sendMessage(
            $chatId,
            'Выберите месяц и год или введите их в формате "Месяц Год" (например "Январь 2024"):',
            $keyboard
        );
    }

    private function handleSpreadsheetMonth(int $chatId, User $user, string $message): void
    {
        $this->logger->info('Handling spreadsheet month selection', [
            'message' => $message,
            'chat_id' => $chatId,
        ]);

        // Check if message contains both month and year
        if (preg_match('/^\s*([а-яА-Я]+)\s+(\d{4})\s*$/u', trim($message), $matches)) {
            $this->logger->info('Message matches pattern', [
                'matches' => $matches,
            ]);

            $monthName = $matches[1];
            $year = (int) $matches[2];

            $month = $this->getMonthNumber($monthName);
            $this->logger->info('Month number conversion', [
                'monthName' => $monthName,
                'month' => $month,
                'year' => $year,
            ]);

            if (!$month || $year < 2000 || $year > 2100) {
                $this->logger->warning('Invalid month or year', [
                    'month' => $month,
                    'year' => $year,
                ]);
                $this->sendMessage($chatId, 'Неверный формат. Используйте формат "Месяц Год" (например "Январь 2024")');

                return;
            }

            $tempData = $user->getTempData();
            $spreadsheetId = $tempData['spreadsheet_id'];

            try {
                $this->sheetsService->addSpreadsheet($user, $spreadsheetId, $month, $year);
            } catch (\Exception $e) {
                $this->logger->error('Failed to add spreadsheet: '.$e->getMessage(), [
                    'chat_id' => $chatId,
                    'spreadsheet_id' => $spreadsheetId,
                    'month' => $month,
                    'year' => $year,
                ]);
                $this->sendMessage($chatId, 'Не удалось добавить таблицу. Попробуйте еще раз.');

                return;
            }

            $user->setState('');
            $user->setTempData([]);
            $this->userRepository->save($user, true);

            $this->sendMessage($chatId, sprintf('Таблица за %s %d успешно добавлена', $this->getMonthName($month), $year));

            return;
        }

        $this->logger->warning('Message does not match pattern', [
            'message' => $message,
            'pattern' => '/^\s*([а-яА-Я]+)\s+(\d{4})\s*$/u',
        ]);
        $this->sendMessage($chatId, 'Неверный формат. Используйте формат "Месяц Год" (например "Январь 2024")');
    }

    private function handleSpreadsheetToDelete(int $chatId, User $user, string $message): void
    {
        $spreadsheets = $this->sheetsService->getSpreadsheetsList($user);
        $spreadsheetToDelete = null;

        foreach ($spreadsheets as $spreadsheet) {
            $spreadsheetName = sprintf(
                '%s %d',
                $spreadsheet['month'],
                $spreadsheet['year']
            );

            if ($spreadsheetName === $message) {
                $spreadsheetToDelete = $spreadsheet;
                break;
            }
        }

        if (!$spreadsheetToDelete) {
            $this->sendMessage($chatId, 'Таблица не найдена');

            return;
        }

        $month = $this->getMonthNumber($spreadsheetToDelete['month']);
        if (!$month) {
            $this->sendMessage($chatId, 'Неверный месяц');

            return;
        }

        $this->sheetsService->removeSpreadsheet($user, $month, $spreadsheetToDelete['year']);

        $user->setState('');
        $this->userRepository->save($user, true);

        $this->sendMessage($chatId, sprintf('Таблица за %s успешно удалена', $message));
    }

    /**
     * @param array<string>|null $keyboard
     */
    private function sendMessage(int $chatId, string $text, ?array $keyboard = null): void
    {
        try {
            if (null !== $keyboard) {
                $this->telegramApi->sendMessageWithKeyboard($chatId, $text, $keyboard);
            } else {
                $this->telegramApi->sendMessage([
                    'chat_id' => $chatId,
                    'text' => $text,
                    'parse_mode' => 'HTML',
                ]);
            }
        } catch (\Exception $e) {
            $this->logger->error('Failed to send message: {error}', [
                'error' => $e->getMessage(),
                'exception' => $e,
                'chat_id' => $chatId,
                'text' => $text,
            ]);
        }
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

    private function getMonthNumber(string $name): ?int
    {
        $months = [
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

        return $months[$name] ?? null;
    }
}
