<?php

namespace App\Tests\Repository;

use App\Entity\User;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\ORM\EntityManagerInterface;
use Doctrine\ORM\Mapping\ClassMetadata;
use Doctrine\ORM\Mapping\ClassMetadataFactory;
use Doctrine\Persistence\ManagerRegistry;
use PHPUnit\Framework\MockObject\MockObject;
use PHPUnit\Framework\TestCase;
use Psr\Log\LoggerInterface;

/** @extends ServiceEntityRepository<User> */
class TestUserRepository extends ServiceEntityRepository
{
    private LoggerInterface $logger;

    public function __construct(ManagerRegistry $registry, LoggerInterface $logger)
    {
        parent::__construct($registry, User::class);
        $this->logger = $logger;
    }

    public function findByTelegramId(int $telegramId): ?User
    {
        $this->logger->info('Finding user by Telegram ID', [
            'telegram_id' => $telegramId,
        ]);

        /** @var User|null $user */
        $user = $this->findOneBy(['telegramId' => $telegramId]);

        $this->logger->info('Found user', [
            'telegram_id' => $telegramId,
            'found' => null !== $user,
            'state' => $user?->getState(),
            'temp_data' => $user?->getTempData(),
        ]);

        return $user;
    }

    public function save(User $user, bool $flush = false): void
    {
        $this->logger->info('Saving user', [
            'telegram_id' => $user->getTelegramId(),
            'state' => $user->getState(),
            'temp_data' => $user->getTempData(),
            'flush' => $flush,
        ]);

        $this->getEntityManager()->persist($user);

        if ($flush) {
            $this->logger->info('Flushing changes to database');
            $this->getEntityManager()->flush();

            // Verify the save by re-reading from database
            $telegramId = $user->getTelegramId();
            if (null !== $telegramId) {
                $savedUser = $this->findByTelegramId($telegramId);
                $this->logger->info('Verified saved user state', [
                    'telegram_id' => $telegramId,
                    'saved_state' => $savedUser?->getState(),
                    'saved_temp_data' => $savedUser?->getTempData(),
                ]);
            }
        }
    }

    public function setUserState(User $user, string $state): void
    {
        $this->logger->info('Setting user state', [
            'telegram_id' => $user->getTelegramId(),
            'old_state' => $user->getState(),
            'new_state' => $state,
        ]);

        $user->setState($state);
        $this->save($user, true);

        // Double check the state was saved
        $this->getEntityManager()->refresh($user);
        $this->logger->info('Verified user state after save', [
            'telegram_id' => $user->getTelegramId(),
            'current_state' => $user->getState(),
        ]);
    }

    public function clearUserState(User $user): void
    {
        $this->logger->info('Clearing user state', [
            'telegram_id' => $user->getTelegramId(),
            'old_state' => $user->getState(),
            'temp_data' => $user->getTempData(),
        ]);

        $user->setState('');
        $user->setTempData([]);
        $this->save($user, true);

        // Double check the state was cleared
        $this->getEntityManager()->refresh($user);
        $this->logger->info('Verified user state after clear', [
            'telegram_id' => $user->getTelegramId(),
            'current_state' => $user->getState(),
            'current_temp_data' => $user->getTempData(),
        ]);
    }
}

class UserRepositoryTest extends TestCase
{
    /** @var TestUserRepository&MockObject */
    private TestUserRepository $repository;
    /** @var EntityManagerInterface&MockObject */
    private EntityManagerInterface $entityManager;
    /** @var LoggerInterface&MockObject */
    private LoggerInterface $logger;
    /** @var ManagerRegistry&MockObject */
    private ManagerRegistry $registry;
    /** @var ClassMetadataFactory&MockObject */
    private ClassMetadataFactory $metadataFactory;
    /** @var ClassMetadata<User> */
    private ClassMetadata $metadata;

    protected function setUp(): void
    {
        $this->entityManager = $this->createMock(EntityManagerInterface::class);
        $this->logger = $this->createMock(LoggerInterface::class);
        $this->registry = $this->createMock(ManagerRegistry::class);
        $this->metadataFactory = $this->createMock(ClassMetadataFactory::class);

        $this->metadata = new ClassMetadata(User::class);

        $this->registry->method('getManagerForClass')
            ->with(User::class)
            ->willReturn($this->entityManager);

        $this->entityManager->method('getMetadataFactory')
            ->willReturn($this->metadataFactory);

        $this->metadataFactory->method('getMetadataFor')
            ->with(User::class)
            ->willReturn($this->metadata);

        $this->entityManager->method('getClassMetadata')
            ->with(User::class)
            ->willReturn($this->metadata);

        $this->repository = $this->getMockBuilder(TestUserRepository::class)
            ->setConstructorArgs([$this->registry, $this->logger])
            ->onlyMethods(['findOneBy'])
            ->getMock();
    }

    public function testFindByTelegramId(): void
    {
        $telegramId = 123456;
        $user = new User();
        $user->setTelegramId($telegramId);
        $user->setState('test_state');
        $user->setTempData(['key' => 'value']);

        $this->repository->expects($this->once())
            ->method('findOneBy')
            ->with(['telegramId' => $telegramId])
            ->willReturn($user);

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context) use ($telegramId) {
                static $callCount = 0;
                ++$callCount;

                if (1 === $callCount) {
                    $this->assertEquals('Finding user by Telegram ID', $message);
                    $this->assertEquals(['telegram_id' => $telegramId], $context);
                } elseif (2 === $callCount) {
                    $this->assertEquals('Found user', $message);
                    $this->assertEquals([
                        'telegram_id' => $telegramId,
                        'found' => true,
                        'state' => 'test_state',
                        'temp_data' => ['key' => 'value'],
                    ], $context);
                }

                return null;
            });

        $result = $this->repository->findByTelegramId($telegramId);
        $this->assertSame($user, $result);
    }

    public function testFindByTelegramIdWhenNotFound(): void
    {
        $telegramId = 123456;

        $this->repository->expects($this->once())
            ->method('findOneBy')
            ->with(['telegramId' => $telegramId])
            ->willReturn(null);

        $this->logger->expects($this->exactly(2))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context) use ($telegramId) {
                static $callCount = 0;
                ++$callCount;

                if (1 === $callCount) {
                    $this->assertEquals('Finding user by Telegram ID', $message);
                    $this->assertEquals(['telegram_id' => $telegramId], $context);
                } elseif (2 === $callCount) {
                    $this->assertEquals('Found user', $message);
                    $this->assertEquals([
                        'telegram_id' => $telegramId,
                        'found' => false,
                        'state' => null,
                        'temp_data' => null,
                    ], $context);
                }

                return null;
            });

        $result = $this->repository->findByTelegramId($telegramId);
        $this->assertNull($result);
    }

    public function testSave(): void
    {
        $user = new User();
        $user->setTelegramId(123456);
        $user->setState('test_state');
        $user->setTempData(['key' => 'value']);

        $this->entityManager->expects($this->once())
            ->method('persist')
            ->with($user);

        $this->logger->expects($this->once())
            ->method('info')
            ->with('Saving user', [
                'telegram_id' => 123456,
                'state' => 'test_state',
                'temp_data' => ['key' => 'value'],
                'flush' => false,
            ]);

        $this->repository->save($user, false);
    }

    public function testSaveWithFlush(): void
    {
        $user = new User();
        $user->setTelegramId(123456);
        $user->setState('test_state');
        $user->setTempData(['key' => 'value']);

        $this->repository->expects($this->once())
            ->method('findOneBy')
            ->with(['telegramId' => 123456])
            ->willReturn($user);

        $this->entityManager->expects($this->once())
            ->method('persist')
            ->with($user);

        $this->entityManager->expects($this->once())
            ->method('flush');

        $this->logger->expects($this->exactly(5))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context = []) {
                static $callCount = 0;
                ++$callCount;

                if (1 === $callCount) {
                    $this->assertEquals('Saving user', $message);
                    $this->assertEquals([
                        'telegram_id' => 123456,
                        'state' => 'test_state',
                        'temp_data' => ['key' => 'value'],
                        'flush' => true,
                    ], $context);
                } elseif (2 === $callCount) {
                    $this->assertEquals('Flushing changes to database', $message);
                } elseif (3 === $callCount) {
                    $this->assertEquals('Finding user by Telegram ID', $message);
                    $this->assertEquals(['telegram_id' => 123456], $context);
                } elseif (4 === $callCount) {
                    $this->assertEquals('Found user', $message);
                    $this->assertEquals([
                        'telegram_id' => 123456,
                        'found' => true,
                        'state' => 'test_state',
                        'temp_data' => ['key' => 'value'],
                    ], $context);
                } elseif (5 === $callCount) {
                    $this->assertEquals('Verified saved user state', $message);
                    $this->assertEquals([
                        'telegram_id' => 123456,
                        'saved_state' => 'test_state',
                        'saved_temp_data' => ['key' => 'value'],
                    ], $context);
                }

                return null;
            });

        $this->repository->save($user, true);
    }

    public function testSetUserState(): void
    {
        $user = new User();
        $user->setTelegramId(123456);
        $user->setState('old_state');

        $this->repository->expects($this->once())
            ->method('findOneBy')
            ->with(['telegramId' => 123456])
            ->willReturn($user);

        $this->entityManager->expects($this->once())
            ->method('persist')
            ->with($user);

        $this->entityManager->expects($this->once())
            ->method('flush');

        $this->entityManager->expects($this->once())
            ->method('refresh')
            ->with($user);

        $this->logger->expects($this->exactly(7))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context) {
                static $callCount = 0;
                ++$callCount;

                if (1 === $callCount) {
                    $this->assertEquals('Setting user state', $message);
                    $this->assertEquals([
                        'telegram_id' => 123456,
                        'old_state' => 'old_state',
                        'new_state' => 'new_state',
                    ], $context);
                } elseif (2 === $callCount) {
                    $this->assertEquals('Saving user', $message);
                } elseif (3 === $callCount) {
                    $this->assertEquals('Flushing changes to database', $message);
                } elseif (4 === $callCount) {
                    $this->assertEquals('Finding user by Telegram ID', $message);
                } elseif (5 === $callCount) {
                    $this->assertEquals('Found user', $message);
                } elseif (6 === $callCount) {
                    $this->assertEquals('Verified saved user state', $message);
                } elseif (7 === $callCount) {
                    $this->assertEquals('Verified user state after save', $message);
                    $this->assertEquals([
                        'telegram_id' => 123456,
                        'current_state' => 'new_state',
                    ], $context);
                }

                return null;
            });

        $this->repository->setUserState($user, 'new_state');
        $this->assertEquals('new_state', $user->getState());
    }

    public function testClearUserState(): void
    {
        $user = new User();
        $user->setTelegramId(123456);
        $user->setState('old_state');
        $user->setTempData(['key' => 'value']);

        $this->repository->expects($this->once())
            ->method('findOneBy')
            ->with(['telegramId' => 123456])
            ->willReturn($user);

        $this->entityManager->expects($this->once())
            ->method('persist')
            ->with($user);

        $this->entityManager->expects($this->once())
            ->method('flush');

        $this->entityManager->expects($this->once())
            ->method('refresh')
            ->with($user);

        $this->logger->expects($this->exactly(7))
            ->method('info')
            ->willReturnCallback(function (string $message, array $context) {
                static $callCount = 0;
                ++$callCount;

                if (1 === $callCount) {
                    $this->assertEquals('Clearing user state', $message);
                    $this->assertEquals([
                        'telegram_id' => 123456,
                        'old_state' => 'old_state',
                        'temp_data' => ['key' => 'value'],
                    ], $context);
                } elseif (2 === $callCount) {
                    $this->assertEquals('Saving user', $message);
                } elseif (3 === $callCount) {
                    $this->assertEquals('Flushing changes to database', $message);
                } elseif (4 === $callCount) {
                    $this->assertEquals('Finding user by Telegram ID', $message);
                } elseif (5 === $callCount) {
                    $this->assertEquals('Found user', $message);
                } elseif (6 === $callCount) {
                    $this->assertEquals('Verified saved user state', $message);
                } elseif (7 === $callCount) {
                    $this->assertEquals('Verified user state after clear', $message);
                    $this->assertEquals([
                        'telegram_id' => 123456,
                        'current_state' => '',
                        'current_temp_data' => [],
                    ], $context);
                }

                return null;
            });

        $this->repository->clearUserState($user);
        $this->assertEquals('', $user->getState());
        $this->assertEquals([], $user->getTempData());
    }
}
