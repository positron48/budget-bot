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
        $foodCategory->setName('Food');
        $foodCategory->setType('expense');
        $this->entityManager->persist($foodCategory);

        $transportCategory = new UserCategory();
        $transportCategory->setUser($user);
        $transportCategory->setName('Transport');
        $transportCategory->setType('expense');
        $this->entityManager->persist($transportCategory);

        // Create category keywords
        $foodKeywords = ['grocery', 'restaurant', 'cafe', 'food', 'еда', 'ресторан', 'кафе', 'продукты'];
        foreach ($foodKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($foodCategory);
            $this->entityManager->persist($categoryKeyword);
        }

        $transportKeywords = ['taxi', 'bus', 'metro', 'transport', 'такси', 'автобус', 'метро', 'транспорт'];
        foreach ($transportKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($transportCategory);
            $this->entityManager->persist($categoryKeyword);
        }

        // Create test spreadsheet for current month
        $now = new \DateTime();
        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setUser($user);
        $spreadsheet->setSpreadsheetId('test_spreadsheet_id');
        $spreadsheet->setTitle('Budget');
        $spreadsheet->setMonth((int) $now->format('n'));
        $spreadsheet->setYear((int) $now->format('Y'));
        $this->entityManager->persist($spreadsheet);

        // Flush all changes
        $this->entityManager->flush();
    }
}
