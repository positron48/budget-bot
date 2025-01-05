<?php

namespace App\Service;

use App\Entity\User;
use Longman\TelegramBot\Request;
use Psr\Log\LoggerInterface;

class TransactionHandler
{
    private GoogleSheetsService $sheetsService;
    private CategoryService $categoryService;
    private LoggerInterface $logger;

    public function __construct(
        GoogleSheetsService $sheetsService,
        CategoryService $categoryService,
        LoggerInterface $logger,
    ) {
        $this->sheetsService = $sheetsService;
        $this->categoryService = $categoryService;
        $this->logger = $logger;
    }

    /**
     * @param array{
     *     date: \DateTime,
     *     amount: float,
     *     description: string,
     *     isIncome: bool
     * } $data
     */
    public function handle(int $chatId, User $user, array $data): void
    {
        $spreadsheet = $this->sheetsService->findSpreadsheetByDate($user, $data['date']);
        if (!$spreadsheet) {
            $this->logger->warning('Spreadsheet not found for date', [
                'chat_id' => $chatId,
                'date' => $data['date']->format('Y-m-d'),
            ]);

            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf(
                    'У вас нет таблицы за %s %d. Пожалуйста, добавьте её с помощью команды /add',
                    $this->getMonthName((int) $data['date']->format('m')),
                    (int) $data['date']->format('Y')
                ),
            ]);

            return;
        }

        $category = $this->categoryService->detectCategory(
            $data['description'],
            $data['isIncome'] ? 'income' : 'expense',
            $user
        );

        if (!$category) {
            $this->logger->warning('Category not detected', [
                'chat_id' => $chatId,
                'description' => $data['description'],
                'type' => $data['isIncome'] ? 'income' : 'expense',
            ]);

            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Не удалось определить категорию',
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

        if ($data['isIncome']) {
            $this->logger->info('Adding income', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
                'date' => $data['date']->format('d.m.Y'),
                'amount' => $data['amount'],
                'description' => $data['description'],
                'category' => $category,
            ]);

            $this->sheetsService->addIncome(
                $spreadsheetId,
                $data['date']->format('d.m.Y'),
                $data['amount'],
                $data['description'],
                $category
            );

            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Доход успешно добавлен',
            ]);
        } else {
            $this->logger->info('Adding expense', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
                'date' => $data['date']->format('d.m.Y'),
                'amount' => $data['amount'],
                'description' => $data['description'],
                'category' => $category,
            ]);

            $this->sheetsService->addExpense(
                $spreadsheetId,
                $data['date']->format('d.m.Y'),
                $data['amount'],
                $data['description'],
                $category
            );

            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Расход успешно добавлен',
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
}
