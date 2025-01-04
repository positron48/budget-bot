<?php

namespace App\Repository;

use App\Entity\CategoryKeyword;
use App\Entity\User;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<CategoryKeyword>
 */
class CategoryKeywordRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, CategoryKeyword::class);
    }

    /**
     * @return CategoryKeyword[]
     */
    public function findMatchingKeywords(string $text, string $type, ?User $user = null): array
    {
        $qb = $this->createQueryBuilder('k')
            ->leftJoin('k.category', 'c')
            ->leftJoin('k.userCategory', 'uc')
            ->where('LOWER(k.keyword) LIKE LOWER(:text)')
            ->andWhere('(c.type = :type OR uc.type = :type)')
            ->setParameter('text', '%'.mb_strtolower($text).'%')
            ->setParameter('type', $type);

        if ($user) {
            $qb->andWhere('(c.isDefault = true OR uc.user = :user)')
                ->setParameter('user', $user);
        } else {
            $qb->andWhere('c.isDefault = true')
                ->andWhere('uc.id IS NULL');
        }

        return $qb->getQuery()->getResult();
    }

    public function save(CategoryKeyword $entity, bool $flush = false): void
    {
        $this->getEntityManager()->persist($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }
}
