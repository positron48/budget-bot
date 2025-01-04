<?php

namespace App\Repository;

use App\Entity\User;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;
use Psr\Log\LoggerInterface;

/**
 * @extends ServiceEntityRepository<User>
 */
class UserRepository extends ServiceEntityRepository
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
