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

    private function sendMessage(int $chatId, string $text): void
    {
        try {
            $data = [
                'chat_id' => $chatId,
                'text' => $text,
                'parse_mode' => 'HTML',
            ];

            $this->logger->info('Sending message to Telegram API', [
                'request' => $data,
            ]);

            $response = Request::sendMessage($data);

            $this->logger->info('Received response from Telegram API', [
                'response' => [
                    'ok' => $response->isOk(),
                    'result' => $response->getResult(),
                    'description' => $response->getDescription(),
                    'error_code' => $response->getErrorCode(),
                ],
            ]);

            if (!$response->isOk()) {
                throw new \RuntimeException(sprintf('Failed to send message to Telegram API: %s (Error code: %d)', $response->getDescription() ?: 'Unknown error', $response->getErrorCode() ?: 0));
            }
        } catch (\Throwable $e) {
            $this->logger->error('Error sending message to Telegram API: '.$e->getMessage(), [
                'exception' => $e,
                'request' => $data,
            ]);
        }
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

            $this->sendMessage(
                $chatId,
                sprintf(
                    'У вас нет таблицы за %s %d. Пожалуйста, добавьте её с помощью команды /add',
                    $this->getMonthName((int) $data['date']->format('m')),
                    (int) $data['date']->format('Y')
                )
            );

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

            $this->sendMessage($chatId, 'Не удалось определить категорию');

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

            $this->sendMessage($chatId, sprintf('Доход успешно добавлен в категорию "%s"', $category));
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

            $this->sendMessage($chatId, sprintf('Расход успешно добавлен в категорию "%s"', $category));
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
