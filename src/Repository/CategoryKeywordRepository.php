<?php

namespace App\Repository;

use App\Entity\CategoryKeyword;
use App\Entity\User;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

class CategoryKeywordRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, CategoryKeyword::class);
    }

    public function findMatchingKeywords(string $text, string $type, ?User $user = null): array
    {
        $qb = $this->createQueryBuilder('k')
            ->leftJoin('k.category', 'c')
            ->leftJoin('k.userCategory', 'uc')
            ->where('LOWER(k.keyword) LIKE LOWER(:text)')
            ->andWhere('(c.type = :type OR uc.type = :type)')
            ->setParameter('text', '%' . mb_strtolower($text) . '%')
            ->setParameter('type', $type);

        if ($user) {
            $qb->andWhere('(uc.user = :user OR c.isDefault = true)')
               ->setParameter('user', $user);
        } else {
            $qb->andWhere('c.isDefault = true');
        }

        return $qb->getQuery()->getResult();
    }

    public function save(CategoryKeyword $keyword, bool $flush = false): void
    {
        $this->getEntityManager()->persist($keyword);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }
} 