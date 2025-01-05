<?php

namespace App\Service\Command;

use App\Entity\User;

class StartCommand extends AbstractCommand
{
    public function getName(): string
    {
        return '/start';
    }

    protected function handleCommand(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $user = new User();
            $user->setTelegramId($chatId);

            $this->userRepository->save($user, true);
            $this->logger->info('New user registered', [
                'telegram_id' => $chatId,
            ]);
        }

        $this->sendMessage(
            $chatId,
            'Привет! Я помогу вести учет доходов и расходов в Google Таблицах. '.
            'Отправляйте сообщения в формате: "[дата] [+]сумма описание"'.
            "\n\nДоступные команды:\n".
            "/list - список доступных таблиц\n".
            "/add - добавить таблицу\n".
            '/categories - управление категориями'
        );
    }
}
