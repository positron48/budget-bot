<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\TelegramApiServiceInterface;
use Psr\Log\LoggerInterface;

class StartCommand implements CommandInterface
{
    private UserRepository $userRepository;
    private LoggerInterface $logger;
    private TelegramApiServiceInterface $telegramApi;

    public function __construct(
        UserRepository $userRepository,
        LoggerInterface $logger,
        TelegramApiServiceInterface $telegramApi,
    ) {
        $this->userRepository = $userRepository;
        $this->logger = $logger;
        $this->telegramApi = $telegramApi;
    }

    public function supports(string $message): bool
    {
        return '/start' === $message;
    }

    public function getName(): string
    {
        return '/start';
    }

    public function execute(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $user = new User();
            $user->setTelegramId($chatId);

            $this->userRepository->save($user, true);
            $this->logger->info('New user registered', [
                'telegram_id' => $chatId,
            ]);
        }

        $this->telegramApi->sendMessage([
            'chat_id' => $chatId,
            'text' => 'Привет! Я помогу вести учет доходов и расходов в Google Таблицах. '.
                'Отправляйте сообщения в формате: "[дата] [+]сумма описание"'.
                "\n\nДоступные команды:\n".
                "/list - список доступных таблиц\n".
                "/add - добавить таблицу\n".
                '/categories - управление категориями',
            'parse_mode' => 'HTML',
        ]);
    }
}
