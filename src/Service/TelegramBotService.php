<?php

namespace App\Service;

use App\Repository\UserRepository;
use App\Service\Command\CommandInterface;
use App\Service\StateHandler\StateHandlerRegistry;
use Longman\TelegramBot\Entities\Update;
use Psr\Log\LoggerInterface;

class TelegramBotService
{
    public function __construct(
        private readonly UserRepository $userRepository,
        private readonly CommandRegistry $commandRegistry,
        private readonly StateHandlerRegistry $stateHandlerRegistry,
        private readonly TransactionHandler $transactionHandler,
        private readonly MessageParserService $messageParser,
        private readonly LoggerInterface $logger,
        private readonly TelegramApiServiceInterface $telegramApi,
        string $token,
        string $username,
    ) {
        $this->telegramApi->initialize($token, $username);
    }

    public function handleUpdate(Update $update): void
    {
        $this->logger->info('Processing update', [
            'update' => $update,
        ]);

        $message = $update->getMessage();
        if (null === $message) {
            $this->logger->info('Update does not contain a message');

            return;
        }

        $text = $message->getText();
        if (null === $text) {
            $this->logger->info('Message does not contain text', [
                'chat_id' => $message->getChat()->getId(),
            ]);

            return;
        }

        $chatId = $message->getChat()->getId();
        $user = $this->userRepository->findByTelegramId($chatId);

        // Try to handle as a command
        $command = $this->commandRegistry->findCommand($text);
        if ($command instanceof CommandInterface) {
            $command->execute($chatId, $user, $text);

            return;
        }

        // Try to handle as a state
        if (null !== $user) {
            $handler = $this->stateHandlerRegistry->handleState($chatId, $user, $text);
            if ($handler) {
                return;
            }
        }

        // Try to handle as a transaction
        if (null !== $user) {
            try {
                $data = $this->messageParser->parseMessage($text);
                if (null !== $data) {
                    $this->transactionHandler->handle($chatId, $user, $data);

                    return;
                }
            } catch (\Exception $e) {
                $this->logger->warning('Failed to parse message: '.$e->getMessage(), [
                    'chat_id' => $chatId,
                    'text' => $text,
                ]);
            }
        }

        $this->logger->info('Message not handled', [
            'chat_id' => $chatId,
            'text' => $text,
        ]);
    }
}
