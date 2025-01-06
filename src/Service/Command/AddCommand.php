<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\TelegramApiServiceInterface;

class AddCommand implements CommandInterface
{
    public function __construct(
        private readonly UserRepository $userRepository,
        private readonly TelegramApiServiceInterface $telegramApiService,
    ) {
    }

    public function supports(string $command): bool
    {
        return '/add' === $command;
    }

    public function getName(): string
    {
        return '/add';
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

        $user->setState('WAITING_SPREADSHEET_ID');
        $this->userRepository->save($user, true);

        $this->telegramApiService->sendMessage([
            'chat_id' => $chatId,
            'text' => 'Отправьте ссылку на таблицу или её идентификатор. '.
                'Таблица должна быть создана на основе шаблона: '.
                'https://docs.google.com/spreadsheets/d/1-BxqnQqyBPjyuRxMSrwQ2FDDxR-sQGQs_EZbZEn_Xzc',
            'parse_mode' => 'HTML',
        ]);
    }
}
