<?php

namespace App\Tests\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\Command\CommandInterface;
use App\Service\CommandRegistry;
use App\Service\MessageParserService;
use App\Service\StateHandler\StateHandlerRegistry;
use App\Service\TelegramBotService;
use App\Service\TransactionHandler;
use App\Tests\Mock\UpdateMock;
use Longman\TelegramBot\Request;
use Longman\TelegramBot\Telegram;
use PHPUnit\Framework\Attributes\PreserveGlobalState;
use PHPUnit\Framework\Attributes\RunTestsInSeparateProcesses;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

#[RunTestsInSeparateProcesses]
#[PreserveGlobalState(false)]
class TelegramBotServiceTest extends TestCase
{
    private const BOT_TOKEN = 'test_token';
    private const BOT_USERNAME = 'test_bot';
    private const CHAT_ID = 123456;

    private TelegramBotService $service;
    private UserRepository&MockObject $userRepository;
    private CommandRegistry&MockObject $commandRegistry;
    private StateHandlerRegistry&MockObject $stateHandlerRegistry;
    private TransactionHandler&MockObject $transactionHandler;
    private MessageParserService&MockObject $messageParser;
    private LoggerInterface&MockObject $logger;

    protected function setUp(): void
    {
        parent::setUp();

        $this->userRepository = $this->createMock(UserRepository::class);
        $this->commandRegistry = $this->createMock(CommandRegistry::class);
        $this->stateHandlerRegistry = $this->createMock(StateHandlerRegistry::class);
        $this->transactionHandler = $this->createMock(TransactionHandler::class);
        $this->messageParser = $this->createMock(MessageParserService::class);
        $this->logger = $this->createMock(LoggerInterface::class);

        // Mock Telegram API
        $this->mockTelegramApi();

        $this->service = new TelegramBotService(
            self::BOT_TOKEN,
            self::BOT_USERNAME,
            $this->userRepository,
            $this->commandRegistry,
            $this->stateHandlerRegistry,
            $this->transactionHandler,
            $this->messageParser,
            $this->logger
        );
    }

    private function mockTelegramApi(): void
    {
        if (!function_exists('runkit7_method_redefine')) {
            $this->markTestSkipped('runkit extension is required for this test');
        }

        if (!defined('RUNKIT7_ACC_PUBLIC')) {
            define('RUNKIT7_ACC_PUBLIC', 1);
        }
        if (!defined('RUNKIT7_ACC_STATIC')) {
            define('RUNKIT7_ACC_STATIC', 4);
        }

        runkit7_method_redefine(
            Telegram::class,
            '__construct',
            '$api_key, $bot_username = ""',
            'return;',
            RUNKIT7_ACC_PUBLIC
        );

        runkit7_method_redefine(
            Request::class,
            'initialize',
            '',
            'return;',
            RUNKIT7_ACC_STATIC | RUNKIT7_ACC_PUBLIC
        );

        runkit7_method_redefine(
            Request::class,
            'sendMessage',
            '',
            'return new \App\Tests\Mock\ServerResponseMock(["ok" => true]);',
            RUNKIT7_ACC_STATIC | RUNKIT7_ACC_PUBLIC
        );
    }

    public function testHandleUpdateWithoutMessage(): void
    {
        $update = new UpdateMock(['update_id' => 1]);

        $this->logger
            ->expects(self::exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context) use ($update) {
                static $callNumber = 0;
                ++$callNumber;

                match ($callNumber) {
                    1 => $this->assertLogMessage('Processing update', ['update' => $update->raw_data], $message, $context),
                    2 => $this->assertLogMessage('Update does not contain a message', [], $message, $context),
                    default => self::fail('Unexpected call number'),
                };
            });

        $this->service->handleUpdate($update->raw_data);
    }

    public function testHandleUpdateWithoutText(): void
    {
        $update = new UpdateMock([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => self::CHAT_ID],
            ],
        ]);

        $this->logger
            ->expects(self::exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context) use ($update) {
                static $callNumber = 0;
                ++$callNumber;

                match ($callNumber) {
                    1 => $this->assertLogMessage('Processing update', ['update' => $update->raw_data], $message, $context),
                    2 => $this->assertLogMessage('Message does not contain text', ['chat_id' => self::CHAT_ID], $message, $context),
                    default => self::fail('Unexpected call number'),
                };
            });

        $this->service->handleUpdate($update->raw_data);
    }

    public function testHandleUpdateWithCommand(): void
    {
        $update = new UpdateMock([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => self::CHAT_ID],
                'text' => '/start',
            ],
        ]);

        $user = new User();
        $user->setTelegramId(self::CHAT_ID);

        $this->userRepository
            ->expects(self::once())
            ->method('findByTelegramId')
            ->with(self::CHAT_ID)
            ->willReturn($user);

        $command = $this->createMock(CommandInterface::class);

        $this->commandRegistry
            ->expects(self::once())
            ->method('findCommand')
            ->with('/start')
            ->willReturn($command);

        $this->commandRegistry
            ->expects(self::once())
            ->method('executeCommand')
            ->with($command, self::CHAT_ID, $user, '/start');

        $this->service->handleUpdate($update->raw_data);
    }

    /**
     * @param array<string, mixed> $expectedContext
     * @param array<string, mixed> $actualContext
     */
    private function assertLogMessage(string $expectedMessage, array $expectedContext, string $actualMessage, array $actualContext): void
    {
        self::assertSame($expectedMessage, $actualMessage);
        self::assertSame($expectedContext, $actualContext);
    }
}
