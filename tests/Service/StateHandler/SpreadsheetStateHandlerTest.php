<?php

namespace App\Tests\Service\StateHandler;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\GoogleSheetsService;
use App\Service\StateHandler\SpreadsheetStateHandler;
use App\Service\TelegramApiServiceInterface;
use App\Utility\DateTimeUtility;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

class SpreadsheetStateHandlerTest extends TestCase
{
    private const TEST_CHAT_ID = 123456;

    private SpreadsheetStateHandler $handler;
    /** @var UserRepository&MockObject */
    private UserRepository $userRepository;
    /** @var GoogleSheetsService&MockObject */
    private GoogleSheetsService $sheetsService;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;
    /** @var TelegramApiServiceInterface&MockObject */
    private TelegramApiServiceInterface $telegramApi;
    /** @var DateTimeUtility */
    private DateTimeUtility $dateTimeUtility;

    protected function setUp(): void
    {
        $this->userRepository = $this->createMock(UserRepository::class);
        $this->sheetsService = $this->createMock(GoogleSheetsService::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->telegramApi = $this->createMock(TelegramApiServiceInterface::class);
        $this->dateTimeUtility = new DateTimeUtility();
        $this->dateTimeUtility->setCurrentDate(new \DateTime('2025-01-15'));

        $this->handler = new SpreadsheetStateHandler(
            $this->userRepository,
            $this->sheetsService,
            $this->logger,
            $this->telegramApi,
            $this->dateTimeUtility
        );
    }

    public function testSupportsMethod(): void
    {
        $this->assertTrue($this->handler->supports('WAITING_SPREADSHEET_ACTION'));
        $this->assertTrue($this->handler->supports('WAITING_SPREADSHEET_ID'));
        $this->assertTrue($this->handler->supports('WAITING_SPREADSHEET_MONTH'));
        $this->assertTrue($this->handler->supports('WAITING_SPREADSHEET_TO_DELETE'));
        $this->assertTrue($this->handler->supports('WAITING_REMOVE_SPREADSHEET'));
        $this->assertFalse($this->handler->supports('INVALID_STATE'));
    }

    public function testHandleWithUnsupportedState(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('INVALID_STATE');

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'test message');
        $this->assertFalse($result);
    }

    public function testHandleSpreadsheetActionWithAddAction(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_ACTION');

        $user->expects($this->once())
            ->method('setState')
            ->with('WAITING_SPREADSHEET_ID');

        $this->userRepository->expects($this->once())
            ->method('save')
            ->with($user, true);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'Введите ID таблицы:',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'Добавить таблицу');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetActionWithDeleteActionNoSpreadsheets(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_ACTION');

        $user->expects($this->once())
            ->method('setState')
            ->with('WAITING_SPREADSHEET_TO_DELETE');

        $this->userRepository->expects($this->once())
            ->method('save')
            ->with($user, true);

        $this->sheetsService->expects($this->once())
            ->method('getSpreadsheetsList')
            ->with($user)
            ->willReturn([]);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'У вас нет добавленных таблиц',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'Удалить таблицу');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetActionWithDeleteActionWithSpreadsheets(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_ACTION');

        $user->expects($this->once())
            ->method('setState')
            ->with('WAITING_SPREADSHEET_TO_DELETE');

        $this->userRepository->expects($this->once())
            ->method('save')
            ->with($user, true);

        $this->sheetsService->expects($this->once())
            ->method('getSpreadsheetsList')
            ->with($user)
            ->willReturn([
                ['month' => 'Январь', 'year' => 2024],
                ['month' => 'Февраль', 'year' => 2024],
            ]);

        $this->telegramApi->expects($this->once())
            ->method('sendMessageWithKeyboard')
            ->with(
                self::TEST_CHAT_ID,
                'Выберите таблицу для удаления:',
                ['Январь 2024', 'Февраль 2024']
            );

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'Удалить таблицу');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetActionWithUnknownAction(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_ACTION');

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'Неизвестное действие',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'Unknown Action');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetIdWithValidId(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_ID');

        $this->sheetsService->expects($this->once())
            ->method('handleSpreadsheetId')
            ->with('test-id')
            ->willReturn('test-id');

        $user->expects($this->once())
            ->method('setTempData')
            ->with(['spreadsheet_id' => 'test-id']);

        $user->expects($this->once())
            ->method('setState')
            ->with('WAITING_SPREADSHEET_MONTH');

        $this->userRepository->expects($this->once())
            ->method('save')
            ->with($user, true);

        $this->telegramApi->expects($this->once())
            ->method('sendMessageWithKeyboard');

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'test-id');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetIdWithInvalidId(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_ID');

        $this->sheetsService->expects($this->once())
            ->method('handleSpreadsheetId')
            ->with('invalid-id')
            ->willThrowException(new \Exception('Invalid ID'));

        $this->logger->expects($this->once())
            ->method('warning')
            ->with('Invalid spreadsheet ID: Invalid ID', [
                'chat_id' => self::TEST_CHAT_ID,
                'spreadsheet_id' => 'invalid-id',
            ]);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'Неверный ID таблицы. Попробуйте еще раз:',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'invalid-id');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetMonthWithValidFormat(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_MONTH');

        $user->expects($this->exactly(2))
            ->method('getTempData')
            ->willReturn(['spreadsheet_id' => 'test-id']);

        $this->sheetsService->expects($this->once())
            ->method('addSpreadsheet')
            ->with($user, 'test-id', 1, 2024);

        $user->expects($this->once())
            ->method('setState')
            ->with('');

        $user->expects($this->once())
            ->method('setTempData')
            ->with([]);

        $this->userRepository->expects($this->once())
            ->method('save')
            ->with($user, true);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'Таблица за Январь 2024 успешно добавлена',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'Январь 2024');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetMonthWithInvalidFormat(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_MONTH');

        $user->expects($this->once())
            ->method('getTempData')
            ->willReturn(['spreadsheet_id' => 'test-id']);

        $this->logger->expects($this->once())
            ->method('warning')
            ->with('Message does not match pattern', [
                'message' => 'Invalid Format',
                'pattern' => '/^\s*([а-яА-Я]+)\s+(\d{4})\s*$/u',
            ]);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'Неверный формат. Используйте формат "Месяц Год" (например "Январь 2024")',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'Invalid Format');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetMonthWithInvalidMonth(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_MONTH');

        $user->expects($this->once())
            ->method('getTempData')
            ->willReturn(['spreadsheet_id' => 'test-id']);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'Неверный формат. Используйте формат "Месяц Год" (например "Январь 2024")',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'НеверныйМесяц 2024');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetMonthWithAddError(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_MONTH');

        $user->expects($this->exactly(2))
            ->method('getTempData')
            ->willReturn(['spreadsheet_id' => 'test-id']);

        $this->sheetsService->expects($this->once())
            ->method('addSpreadsheet')
            ->with($user, 'test-id', 1, 2024)
            ->willThrowException(new \Exception('Add Error'));

        $this->logger->expects($this->once())
            ->method('error')
            ->with('Failed to add spreadsheet: Add Error', [
                'chat_id' => self::TEST_CHAT_ID,
                'spreadsheet_id' => 'test-id',
                'month' => 1,
                'year' => 2024,
            ]);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'Не удалось добавить таблицу. Попробуйте еще раз.',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'Январь 2024');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetToDeleteWithValidSpreadsheet(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_TO_DELETE');

        $this->sheetsService->expects($this->once())
            ->method('getSpreadsheetsList')
            ->with($user)
            ->willReturn([
                ['month' => 'Январь', 'year' => 2024],
            ]);

        $this->sheetsService->expects($this->once())
            ->method('removeSpreadsheet')
            ->with($user, 1, 2024);

        $user->expects($this->once())
            ->method('setState')
            ->with('');

        $this->userRepository->expects($this->once())
            ->method('save')
            ->with($user, true);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'Таблица за Январь 2024 успешно удалена',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'Январь 2024');
        $this->assertTrue($result);
    }

    public function testHandleSpreadsheetToDeleteWithInvalidSpreadsheet(): void
    {
        $user = $this->createMock(User::class);
        $user->expects($this->once())
            ->method('getState')
            ->willReturn('WAITING_SPREADSHEET_TO_DELETE');

        $this->sheetsService->expects($this->once())
            ->method('getSpreadsheetsList')
            ->with($user)
            ->willReturn([]);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => self::TEST_CHAT_ID,
                'text' => 'Таблица не найдена',
                'parse_mode' => 'HTML',
            ]);

        $result = $this->handler->handle(self::TEST_CHAT_ID, $user, 'Invalid Spreadsheet');
        $this->assertTrue($result);
    }
}
