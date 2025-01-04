<?php

namespace App\Repository;

use App\Entity\User;
use App\Entity\UserCategory;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

class UserCategoryRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, UserCategory::class);
    }

    public function findByUserAndType(User $user, string $type): array
    {
        return $this->createQueryBuilder('uc')
            ->where('uc.user = :user')
            ->andWhere('uc.type = :type')
            ->setParameter('user', $user)
            ->setParameter('type', $type)
            ->getQuery()
            ->getResult();
    }

    public function save(UserCategory $category, bool $flush = false): void
    {
        $this->getEntityManager()->persist($category);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }
} 