<?php

namespace App\Tests\Service\Command;

use App\Entity\User;
use App\Repository\CategoryRepository;
use App\Repository\UserCategoryRepository;
use App\Service\CategoryService;
use App\Service\Command\MapCommand;
use App\Service\TelegramApiServiceInterface;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;

class MapCommandTest extends TestCase
{
    private MapCommand $command;
    /** @var CategoryService&MockObject */
    private CategoryService $categoryService;
    /** @var CategoryRepository&MockObject */
    private CategoryRepository $categoryRepository;
    /** @var UserCategoryRepository&MockObject */
    private UserCategoryRepository $userCategoryRepository;
    /** @var TelegramApiServiceInterface&MockObject */
    private TelegramApiServiceInterface $telegramApi;

    protected function setUp(): void
    {
        $this->categoryService = $this->createMock(CategoryService::class);
        $this->categoryRepository = $this->createMock(CategoryRepository::class);
        $this->userCategoryRepository = $this->createMock(UserCategoryRepository::class);
        $this->telegramApi = $this->createMock(TelegramApiServiceInterface::class);

        $this->command = new MapCommand(
            $this->categoryService,
            $this->categoryRepository,
            $this->userCategoryRepository,
            $this->telegramApi
        );
    }

    public function testGetName(): void
    {
        $this->assertEquals('/map', $this->command->getName());
    }

    public function testSupports(): void
    {
        $this->assertTrue($this->command->supports('/map'));
        $this->assertTrue($this->command->supports('/map test'));
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

        $this->command->execute($chatId, null, '/map');
    }

    public function testExecuteWithoutDescription(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, укажите описание расхода после команды /map. Например: /map еда',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/map');
    }

    public function testExecuteWithDescription(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->categoryService->expects($this->once())
            ->method('detectCategory')
            ->with('еда', 'expense', $user)
            ->willReturn('Питание');

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Описание "еда" соответствует категории "Питание"',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/map еда');
    }

    public function testExecuteWithDescriptionNoCategory(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->categoryService->expects($this->once())
            ->method('detectCategory')
            ->with('test', 'expense', $user)
            ->willReturn(null);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Для описания "test" категория не найдена',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/map test');
    }

    public function testExecuteWithMapping(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->categoryService->expects($this->once())
            ->method('getCategories')
            ->with(false, $user)
            ->willReturn(['Питание', 'Транспорт']);

        $this->categoryService->expects($this->once())
            ->method('addKeywordToCategory')
            ->with('еда', 'Питание', 'expense', $user);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Добавлено сопоставление: "еда" → "Питание"',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/map еда = Питание');
    }

    public function testExecuteWithMappingInvalidCategory(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->categoryService->expects($this->once())
            ->method('getCategories')
            ->with(false, $user)
            ->willReturn(['Питание', 'Транспорт']);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => 'Категория "Тест" не найдена. Доступные категории:'.PHP_EOL.'Питание'.PHP_EOL.'Транспорт',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/map еда = Тест');
    }

    public function testExecuteShowAllMappings(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);

        $this->categoryRepository->expects($this->once())
            ->method('findByType')
            ->with('expense')
            ->willReturn([]);

        $this->userCategoryRepository->expects($this->once())
            ->method('findByUserAndType')
            ->with($user, 'expense')
            ->willReturn([]);

        $this->telegramApi->expects($this->once())
            ->method('sendMessage')
            ->with([
                'chat_id' => $chatId,
                'text' => "Справочник категорий расходов:\n\n".
                    "Сопоставлений пока нет. Чтобы добавить сопоставление, используйте команду:\n".
                    "/map слово = категория\n\n".
                    "Например:\n".
                    '/map еда = Питание',
                'parse_mode' => 'HTML',
            ]);

        $this->command->execute($chatId, $user, '/map --all');
    }
}
