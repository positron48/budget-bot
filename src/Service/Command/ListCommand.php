<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Repository\UserSpreadsheetRepository;
use App\Service\TelegramApiServiceInterface;
use Psr\Log\LoggerInterface;

class ListCommand implements CommandInterface
{
    public function __construct(
        private readonly UserRepository $userRepository,
        private readonly TelegramApiServiceInterface $telegramApi,
        private readonly UserSpreadsheetRepository $spreadsheetRepository,
        private readonly LoggerInterface $logger,
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

        // Parse month and year from command
        $parts = explode(' ', trim($message));
        $now = new \DateTime();
        $month = (int) $now->format('m');
        $year = (int) $now->format('Y');

        if (count($parts) >= 2) {
            $month = $this->parseMonth($parts[1]);
        }
        if (count($parts) >= 3) {
            $year = (int) $parts[2];
        }

        // Find spreadsheet for the specified month and year
        $spreadsheet = $this->spreadsheetRepository->findByMonthAndYear($user, $month, $year);

        if (!$spreadsheet) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf('У вас нет таблицы за %s %d', $this->getMonthName($month), $year),
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
            'text' => sprintf('Выберите тип транзакций за %s %d:', $this->getMonthName($month), $year),
            'parse_mode' => 'HTML',
            'reply_markup' => json_encode([
                'keyboard' => array_map(fn ($button) => [$button], $keyboard),
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]),
        ]);
    }

    private function parseMonth(string $month): int
    {
        // Try to parse as number first
        if (is_numeric($month)) {
            return (int) $month;
        }

        // Try to parse Russian month name
        $months = [
            'январь' => 1,
            'февраль' => 2,
            'март' => 3,
            'апрель' => 4,
            'май' => 5,
            'июнь' => 6,
            'июль' => 7,
            'август' => 8,
            'сентябрь' => 9,
            'октябрь' => 10,
            'ноябрь' => 11,
            'декабрь' => 12,
        ];

        $monthLower = mb_strtolower($month);

        return $months[$monthLower] ?? (int) (new \DateTime())->format('m');
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
}
