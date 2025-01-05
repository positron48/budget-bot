<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use Psr\Log\LoggerInterface;

class AddCommand extends AbstractCommand
{
    public function __construct(
        UserRepository $userRepository,
        LoggerInterface $logger,
    ) {
        parent::__construct($userRepository, $logger);
    }

    public function getName(): string
    {
        return '/add';
    }

    protected function handleCommand(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $this->sendMessage($chatId, 'Пожалуйста, начните с команды /start');

            return;
        }

        $this->setState($user, 'WAITING_SPREADSHEET_ID');

        $this->sendMessage(
            $chatId,
            'Отправьте ссылку на таблицу или её идентификатор. '.
            'Таблица должна быть создана на основе шаблона: '.
            'https://docs.google.com/spreadsheets/d/1-BxqnQqyBPjyuRxMSrwQ2FDDxR-sQGQs_EZbZEn_Xzc'
        );
    }
}
