<?php

namespace App\Tests\Integration;

use App\Repository\UserRepository;
use App\Service\CategoryService;
use App\Service\StateHandler\CategoryStateHandler;
use App\Service\TransactionHandler;
use App\Tests\Integration\DataFixtures\TestFixtures;
use Doctrine\Common\DataFixtures\Executor\ORMExecutor;
use Doctrine\Common\DataFixtures\Loader;
use Doctrine\Common\DataFixtures\Purger\ORMPurger;
use Psr\Log\LoggerInterface;

class CategoryStateHandlerTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private UserRepository $userRepository;
    private CategoryStateHandler $categoryStateHandler;
    private CategoryService $categoryService;
    private TransactionHandler $transactionHandler;
    private LoggerInterface $logger;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();
        $this->userRepository = $container->get(UserRepository::class);
        $this->categoryService = $container->get(CategoryService::class);
        $this->transactionHandler = $container->get(TransactionHandler::class);
        $this->logger = $container->get(LoggerInterface::class);

        $this->categoryStateHandler = new CategoryStateHandler(
            $this->userRepository,
            $this->categoryService,
            $this->transactionHandler,
            $this->logger,
            $this->telegramApi
        );

        // Load test fixtures
        $loader = new Loader();
        $loader->addFixture(new TestFixtures());
        $executor = new ORMExecutor($this->entityManager, new ORMPurger());
        $executor->execute($loader->getFixtures());
    }

    public function testSupportsMethod(): void
    {
        $this->assertTrue($this->categoryStateHandler->supports('WAITING_CATEGORIES_ACTION'));
        $this->assertTrue($this->categoryStateHandler->supports('WAITING_CATEGORY_SELECTION'));
        $this->assertTrue($this->categoryStateHandler->supports('WAITING_CATEGORY_MAPPING'));
        $this->assertFalse($this->categoryStateHandler->supports('INVALID_STATE'));
    }

    public function testHandleMethodWithUnsupportedState(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        $user->setState('INVALID_STATE');
        $this->userRepository->save($user, true);

        $result = $this->categoryStateHandler->handle(self::TEST_CHAT_ID, $user, 'test message');
        $this->assertFalse($result);
    }

    public function testHandleCategorySelectionWithInvalidCategory(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        // Set up test data
        $user->setState('WAITING_CATEGORY_SELECTION');
        $user->setTempData([
            'pending_transaction' => [
                'date' => new \DateTime(),
                'amount' => 1500,
                'description' => 'test transaction',
                'isIncome' => false,
            ],
        ]);
        $this->userRepository->save($user, true);

        // Test with invalid category
        $result = $this->categoryStateHandler->handle(self::TEST_CHAT_ID, $user, 'NonExistentCategory');
        $this->assertTrue($result);

        // Verify that state remains unchanged
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEquals('WAITING_CATEGORY_SELECTION', $updatedUser->getState());
    }

    public function testHandleCategoryMappingWithInvalidFormat(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        // Set up test data
        $user->setState('WAITING_CATEGORY_MAPPING');
        $user->setTempData([
            'pending_transaction' => [
                'date' => new \DateTime(),
                'amount' => 1500,
                'description' => 'test transaction',
                'isIncome' => false,
            ],
        ]);
        $this->userRepository->save($user, true);

        // Test with invalid mapping format
        $result = $this->categoryStateHandler->handle(self::TEST_CHAT_ID, $user, 'invalid mapping format');
        $this->assertTrue($result);

        // Verify that state remains unchanged
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEquals('WAITING_CATEGORY_MAPPING', $updatedUser->getState());
    }

    public function testHandleCategoryMappingWithNonExistentCategory(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        // Set up test data
        $user->setState('WAITING_CATEGORY_MAPPING');
        $user->setTempData([
            'pending_transaction' => [
                'date' => new \DateTime(),
                'amount' => 1500,
                'description' => 'test transaction',
                'isIncome' => false,
            ],
        ]);
        $this->userRepository->save($user, true);

        // Test with non-existent category
        $result = $this->categoryStateHandler->handle(self::TEST_CHAT_ID, $user, 'keyword = NonExistentCategory');
        $this->assertTrue($result);

        // Verify that state remains unchanged
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEquals('WAITING_CATEGORY_MAPPING', $updatedUser->getState());
    }

    public function testHandleCategoryMappingWithValidMapping(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        // Set up test data
        $user->setState('WAITING_CATEGORY_MAPPING');
        $user->setTempData([
            'pending_transaction' => [
                'date' => new \DateTime(),
                'amount' => 1500,
                'description' => 'test transaction',
                'isIncome' => false,
            ],
        ]);
        $this->userRepository->save($user, true);

        // Test with valid mapping
        $result = $this->categoryStateHandler->handle(self::TEST_CHAT_ID, $user, 'test = Питание');
        $this->assertTrue($result);

        // Verify that state is cleared
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEmpty($updatedUser->getState());
        $this->assertEmpty($updatedUser->getTempData());
    }

    public function testHandleCategoriesActionWithInvalidAction(): void
    {
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        // Set up test data
        $user->setState('WAITING_CATEGORIES_ACTION');
        $this->userRepository->save($user, true);

        // Test with invalid action
        $result = $this->categoryStateHandler->handle(self::TEST_CHAT_ID, $user, 'InvalidAction');
        $this->assertTrue($result);

        // Verify that state is cleared
        $updatedUser = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($updatedUser);
        $this->assertEmpty($updatedUser->getState());
    }
}
