<?php

namespace App\Tests\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\CommandRegistry;
use App\Service\MessageParserService;
use App\Service\StateHandler\StateHandlerRegistry;
use App\Service\TelegramBotService;
use App\Service\TransactionHandler;
use Longman\TelegramBot\Entities\ServerResponse;
use Longman\TelegramBot\Request;
use Longman\TelegramBot\Telegram;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

/**
 * @runTestsInSeparateProcesses
 *
 * @preserveGlobalState disabled
 */
class TelegramBotServiceTest extends TestCase
{
    private const BOT_TOKEN = 'test_token';
    private const BOT_USERNAME = 'test_bot';
    private const CHAT_ID = 123456;

    private TelegramBotService $service;
    private MockObject&UserRepository $userRepository;
    private MockObject&CommandRegistry $commandRegistry;
    private MockObject&StateHandlerRegistry $stateHandlerRegistry;
    private MockObject&TransactionHandler $transactionHandler;
    private MockObject&MessageParserService $messageParser;
    private MockObject&LoggerInterface $logger;

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
        $serverResponse = $this->createMock(ServerResponse::class);
        $serverResponse->method('isOk')->willReturn(true);

        $requestMock = $this->getMockBuilder(Request::class)
            ->disableOriginalConstructor()
            ->getMock();
        $requestMock->method('sendMessage')->willReturn($serverResponse);

        // Mock static methods using runkit
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
            'return new \Longman\TelegramBot\Entities\ServerResponse(["ok" => true]);',
            RUNKIT7_ACC_STATIC | RUNKIT7_ACC_PUBLIC
        );
    }

    public function testHandleUpdateWithoutMessage(): void
    {
        $update = ['update_id' => 1];

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context = []) use ($update) {
                static $calls = 0;
                ++$calls;

                if (1 === $calls) {
                    $this->assertEquals('Processing update', $message);
                    $this->assertEquals(['update' => $update], $context);
                } elseif (2 === $calls) {
                    $this->assertEquals('Update does not contain a message', $message);
                }

                return null;
            });

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithoutText(): void
    {
        $update = [
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => self::CHAT_ID],
            ],
        ];

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context = []) use ($update) {
                static $calls = 0;
                ++$calls;

                if (1 === $calls) {
                    $this->assertEquals('Processing update', $message);
                    $this->assertEquals(['update' => $update], $context);
                } elseif (2 === $calls) {
                    $this->assertEquals('Message does not contain text', $message);
                    $this->assertEquals(['chat_id' => self::CHAT_ID], $context);
                }

                return null;
            });

        $this->service->handleUpdate($update);
    }

    public function testHandleUpdateWithCommand(): void
    {
        $update = [
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => self::CHAT_ID],
                'text' => '/start',
            ],
        ];

        $user = new User();
        $user->setTelegramId(self::CHAT_ID);

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with(self::CHAT_ID)
            ->willReturn($user);

        $this->commandRegistry->expects($this->once())
            ->method('findCommand')
            ->with('/start')
            ->willReturn($this->createMock(\App\Service\Command\CommandInterface::class));

        $this->commandRegistry->expects($this->once())
            ->method('executeCommand')
            ->with(
                $this->isInstanceOf(\App\Service\Command\CommandInterface::class),
                self::CHAT_ID,
                $user,
                '/start'
            );

        $this->service->handleUpdate($update);
    }
}
