<?php

namespace App\Tests\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\Command\CommandInterface;
use App\Service\CommandRegistry;
use App\Service\MessageParserService;
use App\Service\StateHandler\StateHandlerRegistry;
use App\Service\TelegramApiServiceInterface;
use App\Service\TelegramBotService;
use App\Service\TransactionHandler;
use Longman\TelegramBot\Entities\Chat;
use Longman\TelegramBot\Entities\Message;
use Longman\TelegramBot\Entities\Update;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

class TelegramBotServiceTest extends TestCase
{
    private const TEST_CHAT_ID = 123456;

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
            'test_bot'
        );
    }

    public function testHandleUpdateWithoutMessage(): void
    {
        $update = new Update(['update_id' => 1]);

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context = []) use ($update) {
                static $callCount = 0;
                ++$callCount;

                if (1 === $callCount) {
                    $this->assertEquals('Processing update', $message);
                    $this->assertEquals(['update' => $update], $context);
                } elseif (2 === $callCount) {
                    $this->assertEquals('Update does not contain a message', $message);
                }

                return null;
            });

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithoutText(): void
    {
        $chat = new Chat([
            'id' => self::TEST_CHAT_ID,
            'type' => 'private',
        ]);

        $message = new Message([
            'message_id' => 1,
            'chat' => $chat,
            'date' => time(),
        ]);

        $update = new Update([
            'update_id' => 1,
            'message' => $message,
        ]);

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context = []) use ($update) {
                static $callCount = 0;
                ++$callCount;

                if (1 === $callCount) {
                    $this->assertEquals('Processing update', $message);
                    $this->assertEquals(['update' => $update], $context);
                } elseif (2 === $callCount) {
                    $this->assertEquals('Message does not contain text', $message);
                    $this->assertEquals(['chat_id' => self::TEST_CHAT_ID], $context);
                }

                return null;
            });

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithCommand(): void
    {
        $chat = new Chat([
            'id' => self::TEST_CHAT_ID,
            'type' => 'private',
        ]);

        $message = new Message([
            'message_id' => 1,
            'chat' => $chat,
            'date' => time(),
            'text' => '/start',
        ]);

        $update = new Update([
            'update_id' => 1,
            'message' => $message,
        ]);

        $user = $this->createMock(User::class);
        $command = $this->createMock(CommandInterface::class);

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with(self::TEST_CHAT_ID)
            ->willReturn($user);

        $this->commandRegistry->expects($this->once())
            ->method('findCommand')
            ->with('/start')
            ->willReturn($command);

        $command->expects($this->once())
            ->method('execute')
            ->with(self::TEST_CHAT_ID, $user, '/start');

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Processing update', ['update' => $update]);

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithState(): void
    {
        $chat = new Chat([
            'id' => self::TEST_CHAT_ID,
            'type' => 'private',
        ]);

        $message = new Message([
            'message_id' => 1,
            'chat' => $chat,
            'date' => time(),
            'text' => 'test message',
        ]);

        $update = new Update([
            'update_id' => 1,
            'message' => $message,
        ]);

        $user = $this->createMock(User::class);

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with(self::TEST_CHAT_ID)
            ->willReturn($user);

        $this->commandRegistry->expects($this->once())
            ->method('findCommand')
            ->with('test message')
            ->willReturn(null);

        $this->stateHandlerRegistry->expects($this->once())
            ->method('handleState')
            ->with(self::TEST_CHAT_ID, $user, 'test message')
            ->willReturn(true);

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Processing update', ['update' => $update]);

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithTransaction(): void
    {
        $chat = new Chat([
            'id' => self::TEST_CHAT_ID,
            'type' => 'private',
        ]);

        $message = new Message([
            'message_id' => 1,
            'chat' => $chat,
            'date' => time(),
            'text' => '100 продукты',
        ]);

        $update = new Update([
            'update_id' => 1,
            'message' => $message,
        ]);

        $user = $this->createMock(User::class);
        $transactionData = ['amount' => 100, 'description' => 'продукты'];

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with(self::TEST_CHAT_ID)
            ->willReturn($user);

        $this->commandRegistry->expects($this->once())
            ->method('findCommand')
            ->with('100 продукты')
            ->willReturn(null);

        $this->stateHandlerRegistry->expects($this->once())
            ->method('handleState')
            ->with(self::TEST_CHAT_ID, $user, '100 продукты')
            ->willReturn(false);

        $this->messageParser->expects($this->once())
            ->method('parseMessage')
            ->with('100 продукты')
            ->willReturn($transactionData);

        $this->transactionHandler->expects($this->once())
            ->method('handle')
            ->with(self::TEST_CHAT_ID, $user, $transactionData);

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Processing update', ['update' => $update]);

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithTransactionParseError(): void
    {
        $chat = new Chat([
            'id' => self::TEST_CHAT_ID,
            'type' => 'private',
        ]);

        $message = new Message([
            'message_id' => 1,
            'chat' => $chat,
            'date' => time(),
            'text' => 'invalid transaction',
        ]);

        $update = new Update([
            'update_id' => 1,
            'message' => $message,
        ]);

        $user = $this->createMock(User::class);

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with(self::TEST_CHAT_ID)
            ->willReturn($user);

        $this->commandRegistry->expects($this->once())
            ->method('findCommand')
            ->with('invalid transaction')
            ->willReturn(null);

        $this->stateHandlerRegistry->expects($this->once())
            ->method('handleState')
            ->with(self::TEST_CHAT_ID, $user, 'invalid transaction')
            ->willReturn(false);

        $this->messageParser->expects($this->once())
            ->method('parseMessage')
            ->with('invalid transaction')
            ->willThrowException(new \Exception('Parse error'));

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context = []) use ($update) {
                static $callCount = 0;
                ++$callCount;

                if (1 === $callCount) {
                    $this->assertEquals('Processing update', $message);
                    $this->assertEquals(['update' => $update], $context);
                } elseif (2 === $callCount) {
                    $this->assertEquals('Message not handled', $message);
                    $this->assertEquals(['chat_id' => self::TEST_CHAT_ID, 'text' => 'invalid transaction'], $context);
                }

                return null;
            });

        $this->logger->expects($this->once())
            ->method('warning')
            ->with('Failed to parse message: Parse error', [
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'invalid transaction',
            ]);

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithoutUser(): void
    {
        $chat = new Chat([
            'id' => self::TEST_CHAT_ID,
            'type' => 'private',
        ]);

        $message = new Message([
            'message_id' => 1,
            'chat' => $chat,
            'date' => time(),
            'text' => 'test message',
        ]);

        $update = new Update([
            'update_id' => 1,
            'message' => $message,
        ]);

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with(self::TEST_CHAT_ID)
            ->willReturn(null);

        $this->commandRegistry->expects($this->once())
            ->method('findCommand')
            ->with('test message')
            ->willReturn(null);

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context = []) use ($update) {
                static $callCount = 0;
                ++$callCount;

                if (1 === $callCount) {
                    $this->assertEquals('Processing update', $message);
                    $this->assertEquals(['update' => $update], $context);
                } elseif (2 === $callCount) {
                    $this->assertEquals('Message not handled', $message);
                    $this->assertEquals(['chat_id' => self::TEST_CHAT_ID, 'text' => 'test message'], $context);
                }

                return null;
            });

        $this->service->handleUpdate($update);
    }
}
