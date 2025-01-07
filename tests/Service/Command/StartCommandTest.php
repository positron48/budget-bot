<?php

namespace App\Tests\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\Command\StartCommand;
use App\Service\TelegramApiServiceInterface;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

class StartCommandTest extends TestCase
{
    private StartCommand $command;
    /** @var UserRepository&MockObject */
    private UserRepository $userRepository;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;
    /** @var TelegramApiServiceInterface&MockObject */
    private TelegramApiServiceInterface $telegramApi;

    protected function setUp(): void
    {
        $this->userRepository = $this->createMock(UserRepository::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->telegramApi = $this->createMock(TelegramApiServiceInterface::class);

        $this->command = new StartCommand(
            $this->userRepository,
            $this->logger,
            $this->telegramApi
        );
    }

    public function testGetName(): void
    {
        $this->assertEquals('/start', $this->command->getName());
    }

    public function testSupports(): void
    {
        $this->assertTrue($this->command->supports('/start'));
        $this->assertFalse($this->command->supports('/list_tables'));
    }

    public function testExecuteWithNewUser(): void
    {
        $chatId = 123456;

        $this->userRepository->expects($this->once())
            ->method('save')
            ->with(
                $this->callback(function (User $user) use ($chatId) {
                    return $user->getTelegramId() === $chatId;
                }),
                true
            );

        $this->logger->expects($this->once())
            ->method('info')
            ->with('New user registered', [
                'telegram_id' => $chatId,
            ]);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Привет! Я помогу вести учет доходов и расходов в Google Таблицах. '.
                    'Отправляйте сообщения в формате: "[дата] [+]сумма описание"'.
                    "\n\nДоступные команды:\n".
                    "/list_tables - список доступных таблиц\n".
                    "/list [месяц] [год] - список транзакций\n".
                    "/add - добавить таблицу\n".
                    '/categories - управление категориями',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, null, '/start');
    }

    public function testExecuteWithExistingUser(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->userRepository->expects($this->never())
            ->method('save');

        $this->logger->expects($this->never())
            ->method('info');

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Привет! Я помогу вести учет доходов и расходов в Google Таблицах. '.
                    'Отправляйте сообщения в формате: "[дата] [+]сумма описание"'.
                    "\n\nДоступные команды:\n".
                    "/list_tables - список доступных таблиц\n".
                    "/list [месяц] [год] - список транзакций\n".
                    "/add - добавить таблицу\n".
                    '/categories - управление категориями',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/start');
    }
}
