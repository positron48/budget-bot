<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\TelegramApiServiceInterface;

class CategoriesCommand implements CommandInterface
{
    public function __construct(
        private readonly UserRepository $userRepository,
        private readonly TelegramApiServiceInterface $telegramApiService,
    ) {
    }

    public function supports(string $command): bool
    {
        return '/categories' === $command;
    }

    public function getName(): string
    {
        return '/categories';
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

        $keyboard = [
            ['text' => 'Категории расходов'],
            ['text' => 'Категории доходов'],
        ];

        $user->setState('WAITING_CATEGORIES_ACTION');
        $this->userRepository->save($user, true);

        $this->telegramApiService->sendMessage([
            'chat_id' => $chatId,
            'text' => 'Выберите действие:',
            'parse_mode' => 'HTML',
            'reply_markup' => json_encode([
                'keyboard' => array_map(fn ($button) => [$button], $keyboard),
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]),
        ]);
    }
}
