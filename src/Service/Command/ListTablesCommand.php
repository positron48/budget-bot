<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Service\GoogleSheetsService;
use App\Service\TelegramApiServiceInterface;

class ListTablesCommand implements CommandInterface
{
    public function __construct(
        private readonly GoogleSheetsService $sheetsService,
        private readonly TelegramApiServiceInterface $telegramApiService,
    ) {
    }

    public function supports(string $command): bool
    {
        return '/list_tables' === $command;
    }

    public function getName(): string
    {
        return '/list_tables';
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

        $message = 'Список ваших таблиц:'.PHP_EOL;
        foreach ($spreadsheets as $spreadsheet) {
            $message .= sprintf(
                '%s %d: %s'.PHP_EOL,
                $spreadsheet['month'],
                $spreadsheet['year'],
                $spreadsheet['url']
            );
        }

        $this->telegramApiService->sendMessage([
            'chat_id' => $chatId,
            'text' => $message,
            'parse_mode' => 'HTML',
        ]);
    }
}
