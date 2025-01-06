<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\GoogleSheetsService;
use Psr\Log\LoggerInterface;

class SyncCategoriesCommand extends AbstractCommand
{
    public function __construct(
        UserRepository $userRepository,
        LoggerInterface $logger,
        private readonly GoogleSheetsService $sheetsService,
    ) {
        parent::__construct($userRepository, $logger);
    }

    public function getName(): string
    {
        return '/sync_categories';
    }

    protected function handleCommand(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $this->sendMessage($chatId, 'Пожалуйста, начните с команды /start');

            return;
        }

        $spreadsheet = $this->sheetsService->findLatestSpreadsheet($user);
        if (!$spreadsheet) {
            $this->sendMessage(
                $chatId,
                'У вас пока нет добавленных таблиц. Используйте команду /add чтобы добавить таблицу'
            );

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
            $changes = $this->sheetsService->syncCategories($user, $spreadsheetId);

            if (empty($changes['added_to_db']['expense'])
                && empty($changes['added_to_db']['income'])
                && empty($changes['added_to_sheet']['expense'])
                && empty($changes['added_to_sheet']['income'])
            ) {
                $this->sendMessage($chatId, 'Все категории уже синхронизированы');

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

            $this->sendMessage($chatId, $message);
        } catch (\Exception $e) {
            $this->logger->error('Failed to sync categories: '.$e->getMessage(), [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
                'exception' => $e,
            ]);
            $this->sendMessage($chatId, 'Не удалось синхронизировать категории. Попробуйте еще раз.');
        }
    }
}
