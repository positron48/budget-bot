<?php

namespace App\Tests\Service\Command;

use App\Entity\User;
use App\Service\Command\ListTablesCommand;
use App\Service\GoogleSheetsService;
use App\Service\TelegramApiServiceInterface;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;

class ListTablesCommandTest extends TestCase
{
    private ListTablesCommand $command;
    /** @var GoogleSheetsService&MockObject */
    private GoogleSheetsService $sheetsService;
    /** @var TelegramApiServiceInterface&MockObject */
    private TelegramApiServiceInterface $telegramApi;

    protected function setUp(): void
    {
        $this->sheetsService = $this->createMock(GoogleSheetsService::class);
        $this->telegramApi = $this->createMock(TelegramApiServiceInterface::class);

        $this->command = new ListTablesCommand(
            $this->sheetsService,
            $this->telegramApi
        );
    }

    public function testGetName(): void
    {
        $this->assertEquals('/list_tables', $this->command->getName());
    }

    public function testSupports(): void
    {
        $this->assertTrue($this->command->supports('/list_tables'));
        $this->assertFalse($this->command->supports('/list'));
        $this->assertFalse($this->command->supports('/list_tables 123'));
    }

    public function testExecuteWithoutUser(): void
    {
        $chatId = 123456;

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, начните с команды /start',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, null, '/list_tables');
    }

    public function testExecuteWithEmptySpreadsheets(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->sheetsService->method('getSpreadsheetsList')
            ->with($user)
            ->willReturn([]);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'У вас пока нет добавленных таблиц. Используйте команду /add чтобы добавить таблицу',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/list_tables');
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

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => "Список ваших таблиц:\n".
                    "Январь 2024: https://docs.google.com/spreadsheets/d/123\n".
                    "Февраль 2024: https://docs.google.com/spreadsheets/d/456\n",
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/list_tables');
    }
}
