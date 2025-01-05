<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\GoogleSheetsService;
use Psr\Log\LoggerInterface;

class ListCommand extends AbstractCommand
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
        return '/list';
    }

    protected function handleCommand(int $chatId, ?User $user, string $message): void
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

        $message = "Ваши таблицы:\n\n";
        foreach ($spreadsheets as $spreadsheet) {
            $message .= sprintf(
                "%s %d: %s\n",
                $spreadsheet['month'],
                $spreadsheet['year'],
                $spreadsheet['url']
            );
        }

        $this->sendMessage($chatId, $message);
    }
}
