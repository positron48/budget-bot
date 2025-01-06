<?php

namespace App\Tests\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\Command\CategoriesCommand;
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
class CategoriesCommandTest extends TestCase
{
    private CategoriesCommand $command;
    /** @var UserRepository&MockObject */
    private UserRepository $userRepository;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;

    protected function setUp(): void
    {
        $this->userRepository = $this->createMock(UserRepository::class);
        $this->logger = $this->createMock(LoggerInterface::class);

        // Mock Telegram API
        $this->mockTelegramApi();

        $this->command = new CategoriesCommand(
            $this->userRepository,
            $this->logger,
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
        $this->assertEquals('/categories', $this->command->getName());
    }

    public function testSupports(): void
    {
        $this->assertTrue($this->command->supports('/categories'));
        $this->assertFalse($this->command->supports('/start'));
    }

    public function testExecuteWithoutUser(): void
    {
        $chatId = 123456;

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context) use ($chatId) {
                static $callNumber = 0;
                ++$callNumber;

                if (1 === $callNumber) {
                    $this->assertEquals('Sending message to Telegram API', $message);
                    $this->assertEquals([
                        'request' => [
                            'chat_id' => $chatId,
                            'text' => 'Пожалуйста, начните с команды /start',
                            'parse_mode' => 'HTML',
                        ],
                    ], $context);
                } elseif (2 === $callNumber) {
                    $this->assertEquals('Received response from Telegram API', $message);
                    $this->assertEquals([
                        'response' => [
                            'ok' => true,
                            'result' => null,
                            'description' => null,
                            'error_code' => null,
                        ],
                    ], $context);
                }
            });

        $this->command->execute($chatId, null, '/categories');
    }

    public function testExecuteWithUser(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->userRepository->expects($this->once())
            ->method('save')
            ->with(
                $this->callback(function (User $savedUser) use ($user) {
                    return $savedUser === $user && 'WAITING_CATEGORIES_ACTION' === $savedUser->getState();
                }),
                true
            );

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context) use ($chatId) {
                static $callNumber = 0;
                ++$callNumber;

                if (1 === $callNumber) {
                    $this->assertEquals('Sending message to Telegram API', $message);
                    $this->assertEquals([
                        'request' => [
                            'chat_id' => $chatId,
                            'text' => 'Выберите действие:',
                            'parse_mode' => 'HTML',
                            'reply_markup' => json_encode([
                                'keyboard' => [
                                    [['text' => 'Категории расходов']],
                                    [['text' => 'Категории доходов']],
                                ],
                                'resize_keyboard' => true,
                                'one_time_keyboard' => true,
                            ]),
                        ],
                    ], $context);
                } elseif (2 === $callNumber) {
                    $this->assertEquals('Received response from Telegram API', $message);
                    $this->assertEquals([
                        'response' => [
                            'ok' => true,
                            'result' => null,
                            'description' => null,
                            'error_code' => null,
                        ],
                    ], $context);
                }
            });

        $this->command->execute($chatId, $user, '/categories');
    }
}
