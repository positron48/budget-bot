<?php

namespace App\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Utility\MonthUtility;
use Psr\Log\LoggerInterface;

class TransactionHandler
{
    public function __construct(
        private GoogleSheetsService $sheetsService,
        private CategoryService $categoryService,
        private LoggerInterface $logger,
        private UserRepository $userRepository,
        private TelegramApiServiceInterface $telegramApi,
    ) {
    }

    /**
     * @param array<int, string>|null $keyboard
     */
    private function sendMessage(int $chatId, string $text, ?array $keyboard = null): void
    {
        $replyMarkup = false;
        if (null !== $keyboard) {
            $replyMarkup = json_encode([
                'keyboard' => array_map(fn ($button) => [$button], $keyboard),
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]);
            if (!$replyMarkup) {
                $replyMarkup = false;
            }
        }

        $this->telegramApi->sendMessage([
            'chat_id' => $chatId,
            'text' => $text,
            'parse_mode' => 'HTML',
            'reply_markup' => $replyMarkup,
        ]);
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
                    MonthUtility::getMonthName((int) $data['date']->format('m')),
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

            // Get available categories
            $categories = $this->categoryService->getCategories($data['isIncome'], $user);

            // Store transaction data in user's temp data and set state
            $user->setTempData([
                'pending_transaction' => $data,
            ]);
            $user->setState('WAITING_CATEGORY_SELECTION');
            $this->userRepository->save($user, true);

            // Show categories list
            $this->sendMessage(
                $chatId,
                sprintf(
                    'Не удалось определить категорию для "%s". Выберите категорию из списка или добавьте сопоставление:',
                    $data['description']
                ),
                array_merge($categories, ['Добавить сопоставление'])
            );

            return;
        }

        $this->addTransaction($chatId, $user, $data, $category);
    }

    /**
     * @param array{
     *     date: \DateTime,
     *     amount: float,
     *     description: string,
     *     isIncome: bool
     * } $data
     */
    public function addTransaction(int $chatId, User $user, array $data, string $category): void
    {
        $spreadsheet = $this->sheetsService->findSpreadsheetByDate($user, $data['date']);
        if (!$spreadsheet) {
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

            $this->sendMessage(
                $chatId,
                sprintf(
                    "Доход успешно добавлен\nДата: %s\nСумма: %.2f\nТип: доход\nОписание: %s\nКатегория: %s",
                    $data['date']->format('d.m.Y'),
                    $data['amount'],
                    $data['description'],
                    $category
                )
            );
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

            $this->sendMessage(
                $chatId,
                sprintf(
                    "Расход успешно добавлен\nДата: %s\nСумма: %.2f\nТип: расход\nОписание: %s\nКатегория: %s",
                    $data['date']->format('d.m.Y'),
                    $data['amount'],
                    $data['description'],
                    $category
                )
            );
        }
    }
}
