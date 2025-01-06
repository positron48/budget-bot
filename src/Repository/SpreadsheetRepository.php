<?php

namespace App\Repository;

use App\Entity\Spreadsheet;
use App\Entity\User;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<Spreadsheet>
 *
 * @method Spreadsheet|null find($id, $lockMode = null, $lockVersion = null)
 * @method Spreadsheet|null findOneBy(array $criteria, array $orderBy = null)
 * @method Spreadsheet[]    findAll()
 * @method Spreadsheet[]    findBy(array $criteria, array $orderBy = null, $limit = null, $offset = null)
 */
class SpreadsheetRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, Spreadsheet::class);
    }

    /**
     * @return array<int, Spreadsheet>
     */
    public function findByUser(User $user): array
    {
        return $this->findBy(['user' => $user], ['year' => 'DESC', 'month' => 'DESC']);
    }

    public function findLatestByUser(User $user): ?Spreadsheet
    {
        return $this->findOneBy(['user' => $user], ['year' => 'DESC', 'month' => 'DESC']);
    }

    public function findByUserAndMonth(User $user, int $month, int $year): ?Spreadsheet
    {
        return $this->findOneBy([
            'user' => $user,
            'month' => $month,
            'year' => $year,
        ]);
    }
}
