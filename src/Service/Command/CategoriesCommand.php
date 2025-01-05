<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use Psr\Log\LoggerInterface;

class CategoriesCommand extends AbstractCommand
{
    public function __construct(
        UserRepository $userRepository,
        LoggerInterface $logger,
    ) {
        parent::__construct($userRepository, $logger);
    }

    public function getName(): string
    {
        return '/categories';
    }

    public function execute(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $this->sendMessage($chatId, 'Пожалуйста, начните с команды /start');

            return;
        }

        $keyboard = [
            ['text' => 'Категории расходов'],
            ['text' => 'Категории доходов'],
            ['text' => 'Добавить категорию'],
            ['text' => 'Удалить категорию'],
        ];

        $user->setState('WAITING_CATEGORIES_ACTION');
        $this->userRepository->save($user, true);

        $this->sendMessage(
            $chatId,
            'Выберите действие:',
            $keyboard
        );
    }
}
