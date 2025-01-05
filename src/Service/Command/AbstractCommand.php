<?php

namespace App\Service\Command;

use App\Entity\User;
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
     *
     * @throws \RuntimeException
     */
    protected function sendMessage(int $chatId, string $text, ?array $keyboard = null): void
    {
        $data = [
            'chat_id' => $chatId,
            'text' => $text,
            'parse_mode' => 'HTML',
        ];

        if ($keyboard) {
            // Convert flat keyboard array to array of arrays (each inner array is a row)
            $keyboardRows = array_map(
                static function (array $button) {
                    return [$button];
                },
                $keyboard
            );

            $data['reply_markup'] = json_encode([
                'keyboard' => $keyboardRows,
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]);

            $this->logger->debug('Prepared keyboard for Telegram API', [
                'original' => $keyboard,
                'converted' => $keyboardRows,
            ]);
        }

        $this->logger->info('Sending message to Telegram API', [
            'request' => $data,
        ]);

        $response = Request::sendMessage($data);

        $this->logger->info('Received response from Telegram API', [
            'response' => [
                'ok' => $response->isOk(),
                'result' => $response->getResult(),
                'description' => $response->getDescription(),
                'error_code' => $response->getErrorCode(),
            ],
        ]);

        if (!$response->isOk()) {
            $this->logger->error('Failed to send message to Telegram API', [
                'error_code' => $response->getErrorCode(),
                'description' => $response->getDescription(),
                'request' => $data,
            ]);

            throw new \RuntimeException(sprintf('Failed to send message to Telegram API: %s (Error code: %d)', $response->getDescription() ?: 'Unknown error', $response->getErrorCode() ?: 0));
        }
    }

    public function supports(string $command): bool
    {
        return $command === $this->getName();
    }

    public function execute(int $chatId, ?User $user, string $message): void
    {
        if ($user) {
            $user->setState('');
            $user->setTempData([]);
        }
        $this->handleCommand($chatId, $user, $message);
        if ($user && '' === $user->getState()) {
            $this->userRepository->save($user, true);
        }
    }

    abstract protected function handleCommand(int $chatId, ?User $user, string $message): void;

    protected function setState(?User $user, string $state): void
    {
        if ($user) {
            $user->setState($state);
            $this->userRepository->save($user, true);
        }
    }
}
