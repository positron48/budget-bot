<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Service\CategoryService;
use App\Service\GoogleSheetsService;
use App\Service\TelegramApiServiceInterface;
use Psr\Log\LoggerInterface;

class SyncCategoriesCommand implements CommandInterface
{
    public function __construct(
        private readonly LoggerInterface $logger,
        private readonly GoogleSheetsService $sheetsService,
        private readonly CategoryService $categoryService,
        private readonly TelegramApiServiceInterface $telegramApi,
    ) {
    }

    public function getName(): string
    {
        return '/sync_categories';
    }

    public function supports(string $command): bool
    {
        return $command === $this->getName();
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

        $spreadsheet = $this->sheetsService->findLatestSpreadsheet($user);
        if (!$spreadsheet) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => 'У вас пока нет добавленных таблиц. Используйте команду /add чтобы добавить таблицу',
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

        try {
            // Clear existing categories first
            $expenseCount = count($this->categoryService->getCategories(false, $user));
            $incomeCount = count($this->categoryService->getCategories(true, $user));

            $this->categoryService->clearUserCategories($user);

            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf(
                    'Пользовательские категории очищены:%s- Расходы: %d%s- Доходы: %d',
                    PHP_EOL,
                    $expenseCount,
                    PHP_EOL,
                    $incomeCount
                ),
                'parse_mode' => 'HTML',
            ]);

            // Now sync categories from spreadsheet
            $changes = $this->sheetsService->syncCategories($user, $spreadsheetId);

            if (empty($changes['added_to_db']['expense'])
                && empty($changes['added_to_db']['income'])
                && empty($changes['added_to_sheet']['expense'])
                && empty($changes['added_to_sheet']['income'])
            ) {
                $this->telegramApi->sendMessage([
                    'chat_id' => $chatId,
                    'text' => 'Все категории уже синхронизированы',
                    'parse_mode' => 'HTML',
                ]);

                return;
            }

            $message = 'Синхронизация категорий завершена:'.PHP_EOL;

            if (!empty($changes['added_to_db']['expense']) || !empty($changes['added_to_db']['income'])) {
                $message .= PHP_EOL.'Добавлены в базу данных:'.PHP_EOL;
                foreach ($changes['added_to_db'] as $type => $categories) {
                    if (empty($categories)) {
                        continue;
                    }
                    $message .= '- '.('expense' === $type ? 'Расходы' : 'Доходы').': '.implode(', ', $categories).PHP_EOL;
                }
            }

            if (!empty($changes['added_to_sheet']['expense']) || !empty($changes['added_to_sheet']['income'])) {
                $message .= PHP_EOL.'Добавлены в таблицу:'.PHP_EOL;
                foreach ($changes['added_to_sheet'] as $type => $categories) {
                    if (empty($categories)) {
                        continue;
                    }
                    $message .= '- '.('expense' === $type ? 'Расходы' : 'Доходы').': '.implode(', ', $categories).PHP_EOL;
                }
            }

            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => $message,
                'parse_mode' => 'HTML',
            ]);
        } catch (\Exception $e) {
            $this->logger->error('Failed to sync categories: '.$e->getMessage(), [
                'chat_id' => $chatId,
                'exception' => $e,
            ]);
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => 'Не удалось синхронизировать категории. Попробуйте еще раз.',
                'parse_mode' => 'HTML',
            ]);
        }
    }
}
