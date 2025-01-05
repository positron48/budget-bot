<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\GoogleSheetsService;
use Psr\Log\LoggerInterface;

class RemoveCommand extends AbstractCommand
{
    private GoogleSheetsService $sheetsService;

    public function __construct(
        UserRepository $userRepository,
        LoggerInterface $logger,
        GoogleSheetsService $sheetsService,
    ) {
        parent::__construct($userRepository, $logger);
        $this->sheetsService = $sheetsService;
    }

    public function getName(): string
    {
        return '/remove';
    }

    public function execute(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $this->sendMessage($chatId, 'Пожалуйста, начните с команды /start');

            return;
        }

        $spreadsheets = $this->sheetsService->getSpreadsheetsList($user);

        if (empty($spreadsheets)) {
            $this->sendMessage(
                $chatId,
                'У вас пока нет добавленных таблиц. Используйте команду /add чтобы добавить таблицу'
            );

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
                            (int) $spreadsheet['month'],
                            (int) $spreadsheet['year']
                        );
                        $this->sendMessage($chatId, 'Таблица успешно удалена');
                    } catch (\RuntimeException $e) {
                        $this->sendMessage($chatId, $e->getMessage());
                    }
                    break;
                }
            }

            if (!$found) {
                $this->sendMessage($chatId, 'Таблица не найдена');
            }

            return;
        }

        $keyboard = [];
        foreach ($spreadsheets as $spreadsheet) {
            $keyboard[] = ['text' => sprintf('%s %d', $spreadsheet['month'], $spreadsheet['year'])];
        }

        $user->setState('WAITING_REMOVE_SPREADSHEET');
        $this->userRepository->save($user, true);

        $this->sendMessage(
            $chatId,
            'Выберите таблицу для удаления:',
            $keyboard
        );
    }
}
