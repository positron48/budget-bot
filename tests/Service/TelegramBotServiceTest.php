<?php

namespace App\Tests\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\CommandRegistry;
use App\Service\MessageParserService;
use App\Service\StateHandler\StateHandlerRegistry;
use App\Service\TelegramApiServiceInterface;
use App\Service\TelegramBotService;
use App\Service\TransactionHandler;
use Longman\TelegramBot\Entities\Update;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

class TelegramBotServiceTest extends TestCase
{
    private TelegramBotService $service;
    /** @var UserRepository&MockObject */
    private UserRepository $userRepository;
    /** @var CommandRegistry&MockObject */
    private CommandRegistry $commandRegistry;
    /** @var StateHandlerRegistry&MockObject */
    private StateHandlerRegistry $stateHandlerRegistry;
    /** @var TransactionHandler&MockObject */
    private TransactionHandler $transactionHandler;
    /** @var MessageParserService&MockObject */
    private MessageParserService $messageParser;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;
    /** @var TelegramApiServiceInterface&MockObject */
    private TelegramApiServiceInterface $telegramApi;

    protected function setUp(): void
    {
        $this->userRepository = $this->createMock(UserRepository::class);
        $this->commandRegistry = $this->createMock(CommandRegistry::class);
        $this->stateHandlerRegistry = $this->createMock(StateHandlerRegistry::class);
        $this->transactionHandler = $this->createMock(TransactionHandler::class);
        $this->messageParser = $this->createMock(MessageParserService::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->telegramApi = $this->createMock(TelegramApiServiceInterface::class);

        $this->service = new TelegramBotService(
            $this->userRepository,
            $this->commandRegistry,
            $this->stateHandlerRegistry,
            $this->transactionHandler,
            $this->messageParser,
            $this->logger,
            $this->telegramApi,
            'test_token',
            'test_username'
        );
    }

    public function testHandleUpdateWithoutMessage(): void
    {
        $update = new Update(['update_id' => 1]);

        $this->logger->method('info')
            ->with('Update does not contain a message');

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithoutText(): void
    {
        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => 123456],
            ],
        ]);

        $this->logger->method('info')
            ->with('Message does not contain text', ['chat_id' => 123456]);

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithCommand(): void
    {
        $chatId = 123456;
        $text = '/start';
        $user = new User();

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => $text,
            ],
        ]);

        $this->userRepository->method('findByTelegramId')
            ->with($chatId)
            ->willReturn($user);

        $command = $this->createMock(\App\Service\Command\CommandInterface::class);
        $command->method('execute')
            ->with($chatId, $user, $text);

        $this->commandRegistry->method('findCommand')
            ->with($text)
            ->willReturn($command);

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithState(): void
    {
        $chatId = 123456;
        $text = 'some text';
        $user = new User();

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => $text,
            ],
        ]);

        $this->userRepository->method('findByTelegramId')
            ->with($chatId)
            ->willReturn($user);

        $this->commandRegistry->method('findCommand')
            ->with($text)
            ->willReturn(null);

        $this->stateHandlerRegistry->method('handleState')
            ->with($chatId, $user, $text)
            ->willReturn(true);

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithTransaction(): void
    {
        $chatId = 123456;
        $text = '100 test';
        $user = new User();
        $data = ['amount' => 100, 'description' => 'test'];

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => $text,
            ],
        ]);

        $this->userRepository->method('findByTelegramId')
            ->with($chatId)
            ->willReturn($user);

        $this->commandRegistry->method('findCommand')
            ->with($text)
            ->willReturn(null);

        $this->stateHandlerRegistry->method('handleState')
            ->with($chatId, $user, $text)
            ->willReturn(false);

        $this->messageParser->method('parseMessage')
            ->with($text)
            ->willReturn($data);

        $this->transactionHandler->method('handle')
            ->with($chatId, $user, $data);

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithParseError(): void
    {
        $chatId = 123456;
        $text = 'invalid text';
        $user = new User();
        $exception = new \Exception('Parse error');

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => $text,
            ],
        ]);

        $this->userRepository->method('findByTelegramId')
            ->with($chatId)
            ->willReturn($user);

        $this->commandRegistry->method('findCommand')
            ->with($text)
            ->willReturn(null);

        $this->stateHandlerRegistry->method('handleState')
            ->with($chatId, $user, $text)
            ->willReturn(false);

        $this->messageParser->method('parseMessage')
            ->with($text)
            ->willThrowException($exception);

        $this->logger->method('warning')
            ->with('Failed to parse message: Parse error', [
                'chat_id' => $chatId,
                'text' => $text,
            ]);

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateNotHandled(): void
    {
        $chatId = 123456;
        $text = 'unhandled text';
        $user = new User();

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => $text,
            ],
        ]);

        $this->userRepository->method('findByTelegramId')
            ->with($chatId)
            ->willReturn($user);

        $this->commandRegistry->method('findCommand')
            ->with($text)
            ->willReturn(null);

        $this->stateHandlerRegistry->method('handleState')
            ->with($chatId, $user, $text)
            ->willReturn(false);

        $this->messageParser->method('parseMessage')
            ->with($text)
            ->willReturn(null);

        $this->logger->method('info')
            ->with('Message not handled', [
                'chat_id' => $chatId,
                'text' => $text,
            ]);

        $this->service->handleUpdate($update);
    }
}
