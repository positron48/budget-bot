<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\GoogleSheetsService;
use App\Service\TelegramApiServiceInterface;

class RemoveCommand implements CommandInterface
{
    public function __construct(
        private readonly UserRepository $userRepository,
        private readonly GoogleSheetsService $sheetsService,
        private readonly TelegramApiServiceInterface $telegramApiService,
    ) {
    }

    public function supports(string $command): bool
    {
        return '/remove' === $command;
    }

    public function getName(): string
    {
        return '/remove';
    }

    public function execute(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $this->telegramApiService->sendMessage([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, начните с команды /start',
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        $spreadsheets = $this->sheetsService->getSpreadsheetsList($user);

        if (empty($spreadsheets)) {
            $this->telegramApiService->sendMessage([
                'chat_id' => $chatId,
                'text' => 'У вас пока нет добавленных таблиц. Используйте команду /add чтобы добавить таблицу',
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        // Extract spreadsheet name from command if provided
        $parts = explode(' ', trim($message));
        array_shift($parts); // Remove command name
        $spreadsheetName = implode(' ', $parts);

        if ($spreadsheetName) {
            $found = false;
            foreach ($spreadsheets as $spreadsheet) {
                $title = sprintf('%s %d', $spreadsheet['month'], $spreadsheet['year']);
                if ($title === $spreadsheetName) {
                    $found = true;
                    try {
                        $this->sheetsService->removeSpreadsheet(
                            $user,
                            $this->getMonthNumber($spreadsheet['month']),
                            (int) $spreadsheet['year']
                        );
                        $this->telegramApiService->sendMessage([
                            'chat_id' => $chatId,
                            'text' => 'Таблица успешно удалена',
                            'parse_mode' => 'HTML',
                        ]);
                    } catch (\RuntimeException $e) {
                        $this->telegramApiService->sendMessage([
                            'chat_id' => $chatId,
                            'text' => $e->getMessage(),
                            'parse_mode' => 'HTML',
                        ]);
                    }
                    break;
                }
            }

            if (!$found) {
                $this->telegramApiService->sendMessage([
                    'chat_id' => $chatId,
                    'text' => 'Таблица не найдена',
                    'parse_mode' => 'HTML',
                ]);
            }

            return;
        }

        $keyboard = [];
        foreach ($spreadsheets as $spreadsheet) {
            $keyboard[] = ['text' => sprintf('%s %d', $spreadsheet['month'], $spreadsheet['year'])];
        }

        $user->setState('WAITING_REMOVE_SPREADSHEET');
        $this->userRepository->save($user, true);

        $this->telegramApiService->sendMessage([
            'chat_id' => $chatId,
            'text' => 'Выберите таблицу для удаления:',
            'parse_mode' => 'HTML',
            'reply_markup' => json_encode([
                'keyboard' => array_map(fn ($button) => [$button], $keyboard),
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]),
        ]);
    }

    private function getMonthNumber(string $monthName): int
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

        return $months[$monthName] ?? 0;
    }
}
