<?php

namespace App\Tests\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\Command\AddCommand;
use App\Service\TelegramApiServiceInterface;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;

class AddCommandTest extends TestCase
{
    private AddCommand $command;
    /** @var UserRepository&MockObject */
    private UserRepository $userRepository;
    /** @var TelegramApiServiceInterface&MockObject */
    private TelegramApiServiceInterface $telegramApi;

    protected function setUp(): void
    {
        $this->userRepository = $this->createMock(UserRepository::class);
        $this->telegramApi = $this->createMock(TelegramApiServiceInterface::class);

        $this->command = new AddCommand(
            $this->userRepository,
            $this->telegramApi
        );
    }

    public function testGetName(): void
    {
        $this->assertEquals('/add', $this->command->getName());
    }

    public function testSupports(): void
    {
        $this->assertTrue($this->command->supports('/add'));
        $this->assertFalse($this->command->supports('/start'));
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

        $this->command->execute($chatId, null, '/add');
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
                    return $savedUser === $user && 'WAITING_SPREADSHEET_ID' === $savedUser->getState();
                }),
                true
            );

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Отправьте ссылку на таблицу или её идентификатор. '.
                    'Таблица должна быть создана на основе шаблона: '.
                    'https://docs.google.com/spreadsheets/d/1-BxqnQqyBPjyuRxMSrwQ2FDDxR-sQGQs_EZbZEn_Xzc',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/add');
    }
}
