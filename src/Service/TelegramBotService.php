<?php

namespace App\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\StateHandler\StateHandlerRegistry;
use Longman\TelegramBot\Entities\Message;
use Longman\TelegramBot\Entities\Update;
use Longman\TelegramBot\Exception\TelegramException;
use Longman\TelegramBot\Request;
use Longman\TelegramBot\Telegram;
use Psr\Log\LoggerInterface;

class TelegramBotService
{
    private UserRepository $userRepository;
    private CommandRegistry $commandRegistry;
    private StateHandlerRegistry $stateHandlerRegistry;
    private TransactionHandler $transactionHandler;
    private MessageParserService $messageParser;
    private LoggerInterface $logger;

    public function __construct(
        string $botToken,
        string $botUsername,
        UserRepository $userRepository,
        CommandRegistry $commandRegistry,
        StateHandlerRegistry $stateHandlerRegistry,
        TransactionHandler $transactionHandler,
        MessageParserService $messageParser,
        LoggerInterface $logger,
    ) {
        $this->userRepository = $userRepository;
        $this->commandRegistry = $commandRegistry;
        $this->stateHandlerRegistry = $stateHandlerRegistry;
        $this->transactionHandler = $transactionHandler;
        $this->messageParser = $messageParser;
        $this->logger = $logger;

        try {
            $telegram = new Telegram($botToken, $botUsername);
            Request::initialize($telegram);
        } catch (TelegramException $e) {
            $this->logger->error('Failed to initialize Telegram bot: '.$e->getMessage(), [
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    /**
     * @param array<string, mixed> $updateData
     */
    public function handleUpdate(array $updateData): void
    {
        try {
            $this->logger->info('Processing update', ['update' => $updateData]);
            $update = new Update($updateData);
            $message = $update->getMessage();

            if (!$message instanceof Message) {
                $this->logger->info('Update does not contain a message');

                return;
            }

            $chatId = $message->getChat()->getId();
            $text = $message->getText();

            if (null === $text) {
                $this->logger->info('Message does not contain text', ['chat_id' => $chatId]);

                return;
            }

            $this->logger->info('Processing message', [
                'chat_id' => $chatId,
                'text' => $text,
            ]);

            $user = $this->userRepository->findByTelegramId($chatId);

            // Try to find and execute a command
            $command = $this->commandRegistry->findCommand(message: $text);
            if ($command) {
                $this->commandRegistry->executeCommand($command, $chatId, $user, $text);

                return;
            }

            // If no command found and user exists, try to handle as a regular message
            if ($user) {
                $this->handleRegularMessage($chatId, $user, $text);
            } else {
                $this->sendMessage($chatId, 'Пожалуйста, используйте /start для начала работы.');
            }
        } catch (TelegramException $e) {
            $this->logger->error('Error handling update: '.$e->getMessage(), [
                'exception' => $e,
                'update' => $updateData,
            ]);
        }
    }

    private function handleRegularMessage(int $chatId, User $user, string $text): void
    {
        $state = $user->getState();
        if ($state) {
            $handled = $this->stateHandlerRegistry->handleState($chatId, $user, $text);
            if ($handled) {
                return;
            }
        }

        // Try to parse as a transaction
        try {
            $data = $this->messageParser->parseMessage($text);
            if (null !== $data) {
                $this->transactionHandler->handle($chatId, $user, $data);
            } else {
                $this->sendMessage(
                    $chatId,
                    'Неверный формат сообщения. Используйте формат: "[дата] [+]сумма описание"'
                );
            }
        } catch (\Exception $e) {
            $this->logger->warning('Failed to parse message: '.$e->getMessage(), [
                'chat_id' => $chatId,
                'text' => $text,
            ]);
            $this->sendMessage(
                $chatId,
                'Неверный формат сообщения. Используйте формат: "[дата] [+]сумма описание"'
            );
        }
    }

    /**
     * @param array<int, array<string, string>>|null $keyboard
     */
    private function sendMessage(int $chatId, string $text, ?array $keyboard = null): void
    {
        try {
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
                throw new \RuntimeException(sprintf('Failed to send message to Telegram API: %s (Error code: %d)', $response->getDescription() ?: 'Unknown error', $response->getErrorCode() ?: 0));
            }
        } catch (\Throwable $e) {
            $this->logger->error('Error sending message to Telegram API: '.$e->getMessage(), [
                'exception' => $e,
                'request' => $data,
            ]);
        }
    }
}
