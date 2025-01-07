<?php

namespace App\Tests\Integration\DataFixtures;

use App\Entity\Category;
use App\Entity\CategoryKeyword;
use App\Entity\User;
use App\Entity\UserCategory;
use App\Entity\UserSpreadsheet;
use Doctrine\ORM\EntityManagerInterface;

class TestFixtures
{
    private EntityManagerInterface $entityManager;

    public function __construct(EntityManagerInterface $entityManager)
    {
        $this->entityManager = $entityManager;
    }

    public function load(): void
    {
        $this->loadCategories();
        $this->entityManager->flush();
    }

    private function loadCategories(): void
    {
        // Create test user
        $user = new User();
        $user->setTelegramId(123456);
        $this->entityManager->persist($user);

        // Create user categories
        $foodCategory = new UserCategory();
        $foodCategory->setUser($user);
        $foodCategory->setName('Питание');
        $foodCategory->setType('expense');
        $this->entityManager->persist($foodCategory);

        $cafeCategory = new UserCategory();
        $cafeCategory->setUser($user);
        $cafeCategory->setName('Кафе/Ресторан');
        $cafeCategory->setType('expense');
        $this->entityManager->persist($cafeCategory);

        $transportCategory = new UserCategory();
        $transportCategory->setUser($user);
        $transportCategory->setName('Транспорт');
        $transportCategory->setType('expense');
        $this->entityManager->persist($transportCategory);

        $salaryCategory = new UserCategory();
        $salaryCategory->setUser($user);
        $salaryCategory->setName('Зарплата');
        $salaryCategory->setType('income');
        $this->entityManager->persist($salaryCategory);

        // Create category keywords
        $foodKeywords = ['еда', 'продукты', 'магазин', 'супермаркет', 'пятерочка', 'перекресток', 'магнит', 'ашан', 'продуктовый', 'готовая еда'];
        foreach ($foodKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($foodCategory);
            $this->entityManager->persist($categoryKeyword);
        }

        $cafeKeywords = ['кафе', 'ресторан', 'столовая', 'кофейня', 'бар', 'кофе', 'обед', 'ланч', 'бизнес-ланч'];
        foreach ($cafeKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($cafeCategory);
            $this->entityManager->persist($categoryKeyword);
        }

        $transportKeywords = ['такси', 'метро', 'автобус', 'трамвай', 'маршрутка', 'транспорт', 'проезд', 'uber', 'яндекс.такси', 'ситимобил'];
        foreach ($transportKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($transportCategory);
            $this->entityManager->persist($categoryKeyword);
        }

        $salaryKeywords = ['зп', 'зарплата', 'аванс', 'получка'];
        foreach ($salaryKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($salaryCategory);
            $this->entityManager->persist($categoryKeyword);
        }

        // Create test spreadsheet for current month
        $now = new \DateTime();
        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setUser($user);
        $spreadsheet->setSpreadsheetId('test_spreadsheet_id');
        $spreadsheet->setTitle('Бюджет');
        $spreadsheet->setMonth((int) $now->format('n'));
        $spreadsheet->setYear((int) $now->format('Y'));
        $this->entityManager->persist($spreadsheet);

        // Flush all changes
        $this->entityManager->flush();
    }
}
