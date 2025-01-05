<?php

namespace App\Repository;

use App\Entity\User;
use App\Entity\UserCategory;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<UserCategory>
 */
class UserCategoryRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, UserCategory::class);
    }

    /**
     * @return UserCategory[]
     */
    public function findByUserAndType(User $user, string $type): array
    {
        return $this->findBy(['user' => $user, 'type' => $type]);
    }

    public function save(UserCategory $entity, bool $flush = false): void
    {
        $this->getEntityManager()->persist($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function remove(UserCategory $entity, bool $flush = false): void
    {
        $this->getEntityManager()->remove($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }
}
