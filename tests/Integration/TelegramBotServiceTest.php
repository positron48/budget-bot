<?php

namespace App\Tests\Integration;

use App\Entity\User;
use App\Service\TelegramBotService;
use Longman\TelegramBot\Entities\Chat;
use Longman\TelegramBot\Entities\Message;
use Longman\TelegramBot\Entities\Update;

class TelegramBotServiceTest extends AbstractBotIntegrationTestCase
{
    private TelegramBotService $telegramBotService;
    private const TEST_CHAT_ID = 123456789;

    protected function setUp(): void
    {
        parent::setUp();
        $this->telegramBotService = $this->getContainer()->get(TelegramBotService::class);
    }

    public function testHandleUpdateWithoutMessage(): void
    {
        $update = new Update(['update_id' => 1]);

        $this->telegramBotService->handleUpdate($update);
        // No exception should be thrown
        $this->assertTrue(true);
    }

    public function testHandleUpdateWithoutText(): void
    {
        $chat = new Chat([
            'id' => self::TEST_CHAT_ID,
            'type' => 'private',
        ]);

        $message = new Message([
            'message_id' => 1,
            'chat' => $chat,
            'date' => time(),
        ]);

        $update = new Update([
            'update_id' => 1,
            'message' => $message,
        ]);

        $this->telegramBotService->handleUpdate($update);
        // No exception should be thrown
        $this->assertTrue(true);
    }

    public function testHandleUpdateWithCommand(): void
    {
        $this->createUser();

        $chat = new Chat([
            'id' => self::TEST_CHAT_ID,
            'type' => 'private',
        ]);

        $message = new Message([
            'message_id' => 1,
            'chat' => $chat,
            'date' => time(),
            'text' => '/start',
        ]);

        $update = new Update([
            'update_id' => 1,
            'message' => $message,
        ]);

        $this->telegramBotService->handleUpdate($update);

        // Verify that the command was processed (we should see a response in the mock)
        $this->assertNotEmpty($this->telegramApi->getMessages());
    }

    public function testHandleUpdateWithTransaction(): void
    {
        $user = $this->createUser();
        $this->setupTestSpreadsheet('test-spreadsheet-id');
        $this->setupTestCategories('test-spreadsheet-id');

        $chat = new Chat([
            'id' => self::TEST_CHAT_ID,
            'type' => 'private',
        ]);

        $message = new Message([
            'message_id' => 1,
            'chat' => $chat,
            'date' => time(),
            'text' => '100 продукты',
        ]);

        $update = new Update([
            'update_id' => 1,
            'message' => $message,
        ]);

        $this->telegramBotService->handleUpdate($update);

        // Verify that the transaction was processed
        $this->assertNotEmpty($this->telegramApi->getMessages());
    }

    private function createUser(): User
    {
        $user = new User();
        $user->setTelegramId(self::TEST_CHAT_ID);
        $this->entityManager->persist($user);
        $this->entityManager->flush();

        return $user;
    }
}
