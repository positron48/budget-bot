<?php

namespace App\Tests\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\Command\RemoveCommand;
use App\Service\GoogleSheetsService;
use Longman\TelegramBot\Entities\ServerResponse;
use Longman\TelegramBot\Request;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

/**
 * @runTestsInSeparateProcesses
 *
 * @preserveGlobalState disabled
 */
class RemoveCommandTest extends TestCase
{
    private RemoveCommand $command;
    /** @var UserRepository&MockObject */
    private UserRepository $userRepository;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;
    /** @var GoogleSheetsService&MockObject */
    private GoogleSheetsService $sheetsService;

    protected function setUp(): void
    {
        $this->userRepository = $this->createMock(UserRepository::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->sheetsService = $this->createMock(GoogleSheetsService::class);

        // Mock Telegram API
        $this->mockTelegramApi();

        $this->command = new RemoveCommand(
            $this->userRepository,
            $this->logger,
            $this->sheetsService,
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

    public function testGetName(): void
    {
        $this->assertEquals('/remove', $this->command->getName());
    }

    public function testSupports(): void
    {
        $this->assertTrue($this->command->supports('/remove'));
        $this->assertFalse($this->command->supports('/start'));
    }

    public function testExecuteWithoutUser(): void
    {
        $chatId = 123456;

        $this->logger->expects($this->once())
            ->method('info')
            ->with(
                'Sending message to chat {chat_id}: {message}',
                [
                    'chat_id' => $chatId,
                    'message' => 'Пожалуйста, начните с команды /start',
                ]
            );

        $this->command->execute($chatId, null, '/remove');
    }

    public function testExecuteWithEmptySpreadsheets(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->sheetsService->method('getSpreadsheetsList')
            ->with($user)
            ->willReturn([]);

        $this->logger->expects($this->once())
            ->method('info')
            ->with(
                'Sending message to chat {chat_id}: {message}',
                [
                    'chat_id' => $chatId,
                    'message' => 'У вас пока нет добавленных таблиц. Используйте команду /add чтобы добавить таблицу',
                ]
            );

        $this->command->execute($chatId, $user, '/remove');
    }

    public function testExecuteWithSpreadsheets(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $spreadsheets = [
            [
                'month' => 'Январь',
                'year' => 2024,
                'url' => 'https://docs.google.com/spreadsheets/d/123',
            ],
            [
                'month' => 'Февраль',
                'year' => 2024,
                'url' => 'https://docs.google.com/spreadsheets/d/456',
            ],
        ];

        $this->sheetsService->method('getSpreadsheetsList')
            ->with($user)
            ->willReturn($spreadsheets);

        $this->userRepository->expects($this->once())
            ->method('save')
            ->with(
                $this->callback(function (User $savedUser) use ($user) {
                    return $savedUser === $user && 'WAITING_REMOVE_SPREADSHEET' === $savedUser->getState();
                }),
                true
            );

        $this->logger->expects($this->once())
            ->method('info')
            ->with(
                'Sending message to chat {chat_id}: {message}',
                [
                    'chat_id' => $chatId,
                    'message' => 'Выберите таблицу для удаления:',
                ]
            );

        $this->command->execute($chatId, $user, '/remove');
    }
}
