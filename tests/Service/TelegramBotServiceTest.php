<?php

namespace Tests\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\CategoryService;
use App\Service\GoogleSheetsService;
use App\Service\MessageParserService;
use App\Service\TelegramBotService;
use Longman\TelegramBot\Request;
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
    /** @var UserRepository&MockObject */
    private UserRepository $userRepository;

    /** @var GoogleSheetsService&MockObject */
    private GoogleSheetsService $sheetsService;

    /** @var MessageParserService&MockObject */
    private MessageParserService $messageParser;

    /** @var CategoryService&MockObject */
    private CategoryService $categoryService;

    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;

    private TelegramBotService $telegramBotService;

    protected function setUp(): void
    {
        $this->userRepository = $this->createMock(UserRepository::class);
        $this->sheetsService = $this->createMock(GoogleSheetsService::class);
        $this->messageParser = $this->createMock(MessageParserService::class);
        $this->categoryService = $this->createMock(CategoryService::class);
        $this->logger = $this->createMock(LoggerInterface::class);

        // Mock Telegram API
        $this->createMockForTelegramRequest();

        $this->telegramBotService = new TelegramBotService(
            'test_token',
            'test_username',
            $this->sheetsService,
            $this->messageParser,
            $this->userRepository,
            $this->categoryService,
            $this->logger
        );
    }

    private function createMockForTelegramRequest(): void
    {
        $serverResponse = $this->createMock(\Longman\TelegramBot\Entities\ServerResponse::class);
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

    public function testHandleRemoveCommand(): void
    {
        $user = new User();
        $user->setTelegramId(123);

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with(123)
            ->willReturn($user);

        $this->sheetsService->expects($this->once())
            ->method('removeSpreadsheet')
            ->with($user, 'Январь', 2024);

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Sending message to chat {chat_id}: {message}', [
                'chat_id' => 123,
                'message' => 'Таблица за Январь 2024 успешно удалена',
            ]);

        $this->telegramBotService->handleUpdate([
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => 123, 'type' => 'private'],
                'date' => time(),
                'text' => '/remove Январь 2024',
            ],
        ]);
    }

    public function testHandleRemoveCommandInvalidFormat(): void
    {
        $user = new User();
        $user->setTelegramId(123);

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with(123)
            ->willReturn($user);

        $this->sheetsService->expects($this->never())
            ->method('removeSpreadsheet');

        $this->telegramBotService->handleUpdate([
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => 123, 'type' => 'private'],
                'date' => time(),
                'text' => '/remove InvalidFormat',
            ],
        ]);
    }

    public function testHandleRemoveCommandSpreadsheetNotFound(): void
    {
        $user = new User();
        $user->setTelegramId(123);

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with(123)
            ->willReturn($user);

        $this->sheetsService->expects($this->once())
            ->method('removeSpreadsheet')
            ->with($user, 'Январь', 2024)
            ->willThrowException(new \RuntimeException('Таблица не найдена'));

        $this->telegramBotService->handleUpdate([
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => 123, 'type' => 'private'],
                'date' => time(),
                'text' => '/remove Январь 2024',
            ],
        ]);
    }

    public function testMainUserFlow(): void
    {
        $chatId = 123;
        $spreadsheetId = 'test_spreadsheet_id';
        $month = 'Январь';
        $year = 2024;

        // Step 1: New user starts with /start command
        $user = new User();
        $user->setTelegramId($chatId);

        $this->userRepository->expects($this->once())
            ->method('findByTelegramId')
            ->with($chatId)
            ->willReturn(null);

        // Test consecutive saves
        $this->userRepository
            ->expects($this->exactly(4))
            ->method('save')
            ->willReturnCallback(function ($actualUser, $flush) use ($user) {
                $this->assertEquals($user->getTelegramId(), $actualUser->getTelegramId());
                $this->assertTrue($flush);

                return null;
            });

        // Test consecutive state changes
        $expectedStates = ['WAITING_SPREADSHEET_ID', 'WAITING_MONTH'];
        $stateIndex = 0;

        $this->userRepository
            ->expects($this->exactly(2))
            ->method('setUserState')
            ->willReturnCallback(function ($actualUser, $state) use ($user, &$stateIndex, $expectedStates) {
                $this->assertEquals($user->getTelegramId(), $actualUser->getTelegramId());
                $this->assertEquals($expectedStates[$stateIndex], $state);
                ++$stateIndex;

                return null;
            });

        $this->sheetsService->expects($this->once())
            ->method('handleSpreadsheetId')
            ->with($spreadsheetId)
            ->willReturn($spreadsheetId);

        // Step 3: User selects month
        $this->sheetsService->expects($this->once())
            ->method('addSpreadsheet')
            ->with($user, $spreadsheetId, $month, $year);

        // Step 4: User adds a record
        $this->messageParser->expects($this->once())
            ->method('parseMessage')
            ->with('1000 продукты')
            ->willReturn([
                'amount' => 1000,
                'description' => 'продукты',
                'isIncome' => false,
                'date' => new \DateTime(),
            ]);

        $this->categoryService->expects($this->once())
            ->method('getCategoryByDescription')
            ->with('продукты')
            ->willReturn('Питание');

        $this->sheetsService->expects($this->once())
            ->method('addRecord')
            ->with(
                $user,
                $this->callback(function ($date) {
                    return $date instanceof \DateTime;
                }),
                1000,
                'продукты',
                'Питание',
                false
            );

        // Execute the flow
        // Step 1: Start command
        $this->telegramBotService->handleUpdate([
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId, 'type' => 'private'],
                'date' => time(),
                'text' => '/start',
            ],
        ]);

        // Step 2: Add command
        $this->telegramBotService->handleUpdate([
            'message' => [
                'message_id' => 2,
                'chat' => ['id' => $chatId, 'type' => 'private'],
                'date' => time(),
                'text' => '/add',
            ],
        ]);

        // Step 2.1: Send spreadsheet ID
        $this->telegramBotService->handleUpdate([
            'message' => [
                'message_id' => 3,
                'chat' => ['id' => $chatId, 'type' => 'private'],
                'date' => time(),
                'text' => $spreadsheetId,
            ],
        ]);

        // Step 3: Select month
        $this->telegramBotService->handleUpdate([
            'message' => [
                'message_id' => 4,
                'chat' => ['id' => $chatId, 'type' => 'private'],
                'date' => time(),
                'text' => sprintf('%s %d', $month, $year),
            ],
        ]);

        // Step 4: Add record
        $this->telegramBotService->handleUpdate([
            'message' => [
                'message_id' => 5,
                'chat' => ['id' => $chatId, 'type' => 'private'],
                'date' => time(),
                'text' => '1000 продукты',
            ],
        ]);
    }
}
