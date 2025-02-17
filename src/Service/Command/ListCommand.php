<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Repository\UserSpreadsheetRepository;
use App\Service\StateHandler\StateHandlerRegistry;
use App\Service\TelegramApiServiceInterface;
use App\Utility\DateTimeUtility;
use App\Utility\MonthUtility;
use Psr\Log\LoggerInterface;

class ListCommand implements CommandInterface
{
    public function __construct(
        protected StateHandlerRegistry $stateHandlerRegistry,
        protected UserRepository $userRepository,
        protected UserSpreadsheetRepository $spreadsheetRepository,
        protected DateTimeUtility $dateTimeUtility,
        protected TelegramApiServiceInterface $telegramApi,
        protected LoggerInterface $logger,
    ) {
    }

    public function supports(string $command): bool
    {
        return '/list' === trim($command) || str_starts_with(trim($command), '/list ');
    }

    public function getName(): string
    {
        return '/list';
    }

    public function execute(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, начните с команды /start',
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        $text = trim($message);
        if (str_starts_with($text, '/list ')) {
            $this->handleMonthSpecified($text, $user);

            return;
        }

        $now = clone $this->dateTimeUtility->getCurrentDate();
        $month = (int) $now->format('n');
        $year = (int) $now->format('Y');

        // Find spreadsheet for the specified month and year
        $spreadsheet = $this->spreadsheetRepository->findByMonthAndYear($user, $month, $year);

        if (!$spreadsheet) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf('У вас нет таблицы за %s %d', MonthUtility::getMonthName($month), $year),
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        $spreadsheetId = $spreadsheet->getSpreadsheetId();
        if (!$spreadsheetId) {
            $this->logger->error('Spreadsheet ID is null', [
                'chat_id' => $chatId,
                'spreadsheet' => $spreadsheet,
            ]);
            throw new \RuntimeException('Spreadsheet ID is null');
        }

        // Ask user to choose transaction type
        $keyboard = [
            ['text' => 'Расходы'],
            ['text' => 'Доходы'],
        ];

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => $month,
            'list_year' => $year,
            'spreadsheet_id' => $spreadsheetId,
        ]);
        $this->userRepository->save($user, true);

        $this->telegramApi->sendMessage([
            'chat_id' => $chatId,
            'text' => sprintf('Выберите тип транзакций за %s %d:', MonthUtility::getMonthName($month), $year),
            'parse_mode' => 'HTML',
            'reply_markup' => json_encode([
                'keyboard' => array_map(fn ($button) => [$button], $keyboard),
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]),
        ]);
    }

    private function parseMonth(string $month): ?int
    {
        // Try to parse as number first
        if (is_numeric($month)) {
            $monthNum = (int) $month;
            if ($monthNum < 1 || $monthNum > 12) {
                return null;
            }

            return $monthNum;
        }

        return MonthUtility::getMonthNumber($month);
    }

    private function handleMonthSpecified(string $text, User $user): void
    {
        $chatId = $user->getTelegramId();
        if (null === $chatId) {
            throw new \RuntimeException('User telegram ID is null');
        }

        // Parse month and year from command
        $parts = explode(' ', trim($text));
        $now = clone $this->dateTimeUtility->getCurrentDate();
        $month = (int) $now->format('n');
        $year = (int) $now->format('Y');

        if (count($parts) >= 2) {
            $parsedMonth = $this->parseMonth($parts[1]);
            if (null === $parsedMonth) {
                $this->telegramApi->sendMessage([
                    'chat_id' => $chatId,
                    'text' => 'Неверный формат месяца. Пожалуйста, укажите месяц числом (1-12) или словом (Январь-Декабрь).',
                    'parse_mode' => 'HTML',
                ]);

                return;
            }
            $month = $parsedMonth;
        }
        if (count($parts) >= 3) {
            if (!is_numeric($parts[2])) {
                $this->telegramApi->sendMessage([
                    'chat_id' => $chatId,
                    'text' => 'Неверный формат года. Пожалуйста, укажите год в числовом формате.',
                    'parse_mode' => 'HTML',
                ]);

                return;
            }
            $year = (int) $parts[2];
            if ($year < 2020) {
                $this->telegramApi->sendMessage([
                    'chat_id' => $chatId,
                    'text' => 'Год не может быть меньше 2020.',
                    'parse_mode' => 'HTML',
                ]);

                return;
            }
        }

        // Find spreadsheet for the specified month and year
        $spreadsheet = $this->spreadsheetRepository->findByMonthAndYear($user, $month, $year);

        if (!$spreadsheet) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf('У вас нет таблицы за %s %d', MonthUtility::getMonthName($month), $year),
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        $spreadsheetId = $spreadsheet->getSpreadsheetId();
        if (!$spreadsheetId) {
            $this->logger->error('Spreadsheet ID is null', [
                'chat_id' => $chatId,
                'spreadsheet' => $spreadsheet,
            ]);
            throw new \RuntimeException('Spreadsheet ID is null');
        }

        // Ask user to choose transaction type
        $keyboard = [
            ['text' => 'Расходы'],
            ['text' => 'Доходы'],
        ];

        $user->setState('WAITING_LIST_ACTION');
        $user->setTempData([
            'list_month' => $month,
            'list_year' => $year,
            'spreadsheet_id' => $spreadsheetId,
        ]);
        $this->userRepository->save($user, true);

        $this->telegramApi->sendMessage([
            'chat_id' => $chatId,
            'text' => sprintf('Выберите тип транзакций за %s %d:', MonthUtility::getMonthName($month), $year),
            'parse_mode' => 'HTML',
            'reply_markup' => json_encode([
                'keyboard' => array_map(fn ($button) => [$button], $keyboard),
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]),
        ]);
    }
}
