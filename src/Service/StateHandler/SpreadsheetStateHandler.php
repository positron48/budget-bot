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
        'WAITING_REMOVE_SPREADSHEET',
    ];

    private UserRepository $userRepository;
    private GoogleSheetsService $sheetsService;
    private LoggerInterface $logger;

    public function __construct(
        UserRepository $userRepository,
        GoogleSheetsService $sheetsService,
        LoggerInterface $logger,
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

        if ('WAITING_SPREADSHEET_ID' === $state) {
            $this->handleSpreadsheetId($chatId, $user, $message);

            return;
        }

        if ('WAITING_MONTH' === $state) {
            $this->handleMonthSelection($chatId, $user, $message);

            return;
        }

        if ('WAITING_REMOVE_SPREADSHEET' === $state) {
            $this->handleRemoveSpreadsheet($chatId, $user, $message);
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
            $user->setState('WAITING_MONTH');
            $this->userRepository->save($user, true);

            $keyboard = $this->buildMonthsKeyboard();
            $this->sendMessage(
                $chatId,
                'Выберите месяц из списка или введите в формате "Месяц Год" (например: Январь 2025):',
                $keyboard
            );
        } catch (\RuntimeException $e) {
            $this->logger->warning('Failed to handle spreadsheet', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
                'error' => $e->getMessage(),
            ]);
            $this->sendMessage($chatId, $e->getMessage());
            $user->setState('');
            $this->userRepository->save($user, true);
        }
    }

    private function handleMonthSelection(int $chatId, User $user, string $text): void
    {
        try {
            $monthYear = $this->parseMonthAndYear($text);
            if (!$monthYear) {
                throw new \RuntimeException('Неверный формат даты');
            }

            $month = $monthYear['month'];
            $year = $monthYear['year'];

            $tempData = $user->getTempData();
            $spreadsheetId = $tempData['spreadsheet_id'] ?? null;
            if (!$spreadsheetId) {
                throw new \RuntimeException('Не найден ID таблицы');
            }

            $this->sheetsService->addSpreadsheet($user, $spreadsheetId, $month, $year);

            $this->sendMessage($chatId, sprintf('Таблица за %s %d успешно добавлена', $this->getMonthName($month), $year));
        } catch (\RuntimeException $e) {
            $this->sendMessage($chatId, $e->getMessage());
            $keyboard = array_map(
                static function (array $button) {
                    return ['text' => $button['text']];
                },
                $this->buildMonthsKeyboard()
            );
            $this->sendMessage($chatId, 'Выберите месяц:', $keyboard);

            return;
        }

        $user->setState('');
        $this->userRepository->save($user, true);
    }

    private function handleRemoveSpreadsheet(int $chatId, User $user, string $text): void
    {
        $spreadsheets = $this->sheetsService->getSpreadsheetsList($user);
        $found = false;

        foreach ($spreadsheets as $spreadsheet) {
            $title = sprintf('%s %d', $spreadsheet['month'], $spreadsheet['year']);
            if ($title === $text) {
                $found = true;
                $monthYear = $this->parseMonthAndYear($text);
                if ($monthYear) {
                    try {
                        $this->sheetsService->removeSpreadsheet($user, $monthYear['month'], $monthYear['year']);
                        $this->sendMessage($chatId, 'Таблица успешно удалена');
                    } catch (\RuntimeException $e) {
                        $this->sendMessage($chatId, $e->getMessage());
                    }
                } else {
                    $this->sendMessage($chatId, 'Неверный формат даты');
                }
                break;
            }
        }

        if (!$found) {
            $this->sendMessage($chatId, 'Таблица не найдена');
        }

        $user->setState('');
        $this->userRepository->save($user, true);
    }

    /**
     * @return array<int, array<string, string>>
     */
    private function buildMonthsKeyboard(): array
    {
        $keyboard = [];
        $months = [
            'Январь',
            'Февраль',
            'Март',
            'Апрель',
            'Май',
            'Июнь',
            'Июль',
            'Август',
            'Сентябрь',
            'Октябрь',
            'Ноябрь',
            'Декабрь',
        ];

        foreach ($months as $text) {
            $keyboard[] = ['text' => $text];
        }

        return $keyboard;
    }

    /**
     * @return array{month: int, year: int}|null
     */
    private function parseMonthAndYear(string $text): ?array
    {
        $parts = explode(' ', trim($text));
        if (2 !== count($parts)) {
            return null;
        }

        $monthName = $parts[0];
        $year = (int) $parts[1];

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

        if (!isset($months[$monthName])) {
            return null;
        }

        return [
            'month' => $months[$monthName],
            'year' => $year,
        ];
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

    /**
     * @param array<int, array<string, string>>|null $keyboard
     */
    protected function sendMessage(int $chatId, string $text, ?array $keyboard = null): void
    {
        $data = [
            'chat_id' => $chatId,
            'text' => $text,
            'parse_mode' => 'HTML',
        ];

        if ($keyboard) {
            $data['reply_markup'] = json_encode([
                'keyboard' => array_map(
                    static function (array $button) {
                        return [$button];
                    },
                    $keyboard
                ),
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]);
        }

        Request::sendMessage($data);
    }
}
