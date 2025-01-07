<?php

namespace App\Tests\Integration;

use App\Entity\User;
use App\Repository\UserRepository;

class UserRepositoryIntegrationTest extends AbstractBotIntegrationTestCase
{
    private UserRepository $userRepository;

    protected function setUp(): void
    {
        parent::setUp();
        $this->userRepository = $this->getContainer()->get(UserRepository::class);
    }

    public function testSaveAndFindByTelegramId(): void
    {
        // Create and save a new user
        $user = new User();
        $user->setTelegramId(123456789);
        $user->setState('initial_state');
        $user->setTempData(['key' => 'value']);

        $this->userRepository->save($user, true);

        // Clear entity manager to ensure we're reading from database
        $this->entityManager->clear();

        // Find the user and verify data
        $foundUser = $this->userRepository->findByTelegramId(123456789);
        $this->assertNotNull($foundUser);
        $this->assertEquals(123456789, $foundUser->getTelegramId());
        $this->assertEquals('initial_state', $foundUser->getState());
        $this->assertEquals(['key' => 'value'], $foundUser->getTempData());
    }

    public function testSetUserState(): void
    {
        // Create and save a new user
        $user = new User();
        $user->setTelegramId(987654321);
        $user->setState('initial_state');
        $this->userRepository->save($user, true);

        // Clear entity manager
        $this->entityManager->clear();

        // Get fresh user instance and set state
        $user = $this->userRepository->findByTelegramId(987654321);
        $this->assertNotNull($user);

        $this->userRepository->setUserState($user, 'new_state');

        // Clear entity manager again
        $this->entityManager->clear();

        // Verify state change
        $updatedUser = $this->userRepository->findByTelegramId(987654321);
        $this->assertNotNull($updatedUser);
        $this->assertEquals('new_state', $updatedUser->getState());
    }

    public function testClearUserState(): void
    {
        // Create and save a new user with state and temp data
        $user = new User();
        $user->setTelegramId(555555555);
        $user->setState('some_state');
        $user->setTempData(['some' => 'data']);
        $this->userRepository->save($user, true);

        // Clear entity manager
        $this->entityManager->clear();

        // Get fresh user instance and clear state
        $user = $this->userRepository->findByTelegramId(555555555);
        $this->assertNotNull($user);

        $this->userRepository->clearUserState($user);

        // Clear entity manager again
        $this->entityManager->clear();

        // Verify state and temp data are cleared
        $updatedUser = $this->userRepository->findByTelegramId(555555555);
        $this->assertNotNull($updatedUser);
        $this->assertEquals('', $updatedUser->getState());
        $this->assertEquals([], $updatedUser->getTempData());
    }

    public function testFindByTelegramIdNonExistent(): void
    {
        $nonExistentUser = $this->userRepository->findByTelegramId(999999999);
        $this->assertNull($nonExistentUser);
    }
}
