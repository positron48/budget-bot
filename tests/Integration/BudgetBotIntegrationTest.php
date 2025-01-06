<?php

namespace App\Tests\Integration;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\TelegramBotService;
use Longman\TelegramBot\Entities\Update;
use Symfony\Bundle\FrameworkBundle\Test\KernelTestCase;

class BudgetBotIntegrationTest extends KernelTestCase
{
    private TelegramBotService $telegramBotService;
    private UserRepository $userRepository;

    protected function setUp(): void
    {
        self::bootKernel();

        $container = static::getContainer();
        $this->telegramBotService = $container->get(TelegramBotService::class);
        $this->userRepository = $container->get(UserRepository::class);
    }

    public function testStartCommand(): void
    {
        $chatId = 123456;
        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => '/start',
            ],
        ]);

        $this->telegramBotService->handleUpdate($update);

        $user = $this->userRepository->findByTelegramId($chatId);
        $this->assertNotNull($user);
        $this->assertEquals($chatId, $user->getTelegramId());
    }

    public function testAddCommand(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);
        $this->userRepository->save($user, true);

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => '/add',
            ],
        ]);

        $this->telegramBotService->handleUpdate($update);

        $user = $this->userRepository->findByTelegramId($chatId);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_SPREADSHEET_ID', $user->getState());
    }

    public function testCategoriesCommand(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);
        $this->userRepository->save($user, true);

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => '/categories',
            ],
        ]);

        $this->telegramBotService->handleUpdate($update);

        $user = $this->userRepository->findByTelegramId($chatId);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_CATEGORIES_ACTION', $user->getState());
    }

    public function testRemoveCommand(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);
        $this->userRepository->save($user, true);

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => '/remove',
            ],
        ]);

        $this->telegramBotService->handleUpdate($update);

        $user = $this->userRepository->findByTelegramId($chatId);
        $this->assertNotNull($user);
        $this->assertEquals('WAITING_REMOVE_SPREADSHEET', $user->getState());
    }

    public function testMapCommand(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);
        $this->userRepository->save($user, true);

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => '/map',
            ],
        ]);

        $this->telegramBotService->handleUpdate($update);

        $user = $this->userRepository->findByTelegramId($chatId);
        $this->assertNotNull($user);
    }

    public function testSyncCategoriesCommand(): void
    {
        $chatId = 123456;
        $user = new User();
        $user->setTelegramId($chatId);
        $this->userRepository->save($user, true);

        $update = new Update([
            'update_id' => 1,
            'message' => [
                'message_id' => 1,
                'chat' => ['id' => $chatId],
                'text' => '/sync_categories',
            ],
        ]);

        $this->telegramBotService->handleUpdate($update);

        $user = $this->userRepository->findByTelegramId($chatId);
        $this->assertNotNull($user);
    }
}
