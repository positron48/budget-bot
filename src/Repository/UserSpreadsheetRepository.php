<?php

namespace App\Repository;

use App\Entity\User;
use App\Entity\UserSpreadsheet;
use Doctrine\Bundle\DoctrineBundle\Repository\ServiceEntityRepository;
use Doctrine\Persistence\ManagerRegistry;

/**
 * @extends ServiceEntityRepository<UserSpreadsheet>
 *
 * @method UserSpreadsheet|null find($id, $lockMode = null, $lockVersion = null)
 * @method UserSpreadsheet|null findOneBy(array $criteria, array $orderBy = null)
 * @method UserSpreadsheet[]    findAll()
 * @method UserSpreadsheet[]    findBy(array $criteria, array $orderBy = null, $limit = null, $offset = null)
 */
class UserSpreadsheetRepository extends ServiceEntityRepository
{
    public function __construct(ManagerRegistry $registry)
    {
        parent::__construct($registry, UserSpreadsheet::class);
    }

    public function save(UserSpreadsheet $entity, bool $flush = false): void
    {
        $this->getEntityManager()->persist($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    public function remove(UserSpreadsheet $entity, bool $flush = false): void
    {
        $this->getEntityManager()->remove($entity);

        if ($flush) {
            $this->getEntityManager()->flush();
        }
    }

    /**
     * @return array<string, string>
     */
    public function getSpreadsheetsList(User $user): array
    {
        $spreadsheets = $this->findBy(['user' => $user], ['year' => 'DESC', 'month' => 'DESC']);
        $result = [];

        foreach ($spreadsheets as $spreadsheet) {
            $result[$spreadsheet->getSpreadsheetId()] = sprintf(
                '%s (%s %d)',
                $spreadsheet->getTitle(),
                $spreadsheet->getMonth(),
                $spreadsheet->getYear()
            );
        }

        return $result;
    }

    public function findByDate(User $user, \DateTime $date): ?UserSpreadsheet
    {
        $month = (int) $date->format('n'); // Numeric month without leading zeros
        $year = (int) $date->format('Y');

        return $this->findOneBy([
            'user' => $user,
            'month' => $month,
            'year' => $year,
        ]);
    }

    public function findLatest(User $user): ?UserSpreadsheet
    {
        return $this->findOneBy(
            ['user' => $user],
            ['year' => 'DESC', 'month' => 'DESC']
        );
    }

    public function findByMonthAndYear(User $user, int $month, int $year): ?UserSpreadsheet
    {
        return $this->findOneBy([
            'user' => $user,
            'month' => $month,
            'year' => $year,
        ]);
    }
}
