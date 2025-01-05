<?php

namespace App\Service\Command;

use App\Repository\UserRepository;
use Longman\TelegramBot\Request;
use Psr\Log\LoggerInterface;

abstract class AbstractCommand implements CommandInterface
{
    protected UserRepository $userRepository;
    protected LoggerInterface $logger;

    public function __construct(
        UserRepository $userRepository,
        LoggerInterface $logger,
    ) {
        $this->userRepository = $userRepository;
        $this->logger = $logger;
    }

    /**
     * @param array<int, array<string, string>>|null $keyboard
     */
    protected function sendMessage(int $chatId, string $text, ?array $keyboard = null): void
    {
        $data = [
            'chat_id' => $chatId,
            'text' => $text,
            'parse_mode' => 'HTML',
        ];

        if ($keyboard) {
            $data['reply_markup'] = json_encode([
                'keyboard' => $keyboard,
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]);
        }

        $this->logger->info('Sending message to chat {chat_id}: {message}', [
            'chat_id' => $chatId,
            'message' => $text,
        ]);

        Request::sendMessage($data);
    }

    public function supports(string $command): bool
    {
        return $command === $this->getName();
    }
}
