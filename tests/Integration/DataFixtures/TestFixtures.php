<?php

namespace App\Tests\Integration\DataFixtures;

use App\Entity\Category;
use App\Entity\CategoryKeyword;
use App\Entity\User;
use App\Entity\UserCategory;
use App\Entity\UserSpreadsheet;
use Doctrine\Bundle\FixturesBundle\Fixture;
use Doctrine\Persistence\ObjectManager;

class TestFixtures extends Fixture
{
    public function load(ObjectManager $manager): void
    {
        // Create test user
        $user = new User();
        $user->setTelegramId(123456);
        $manager->persist($user);

        // Create user categories
        $foodCategory = new UserCategory();
        $foodCategory->setUser($user);
        $foodCategory->setName('Питание');
        $foodCategory->setType('expense');
        $manager->persist($foodCategory);

        $cafeCategory = new UserCategory();
        $cafeCategory->setUser($user);
        $cafeCategory->setName('Кафе/Ресторан');
        $cafeCategory->setType('expense');
        $manager->persist($cafeCategory);

        $transportCategory = new UserCategory();
        $transportCategory->setUser($user);
        $transportCategory->setName('Транспорт');
        $transportCategory->setType('expense');
        $manager->persist($transportCategory);

        $salaryCategory = new UserCategory();
        $salaryCategory->setUser($user);
        $salaryCategory->setName('Зарплата');
        $salaryCategory->setType('income');
        $manager->persist($salaryCategory);

        $bonusCategory = new UserCategory();
        $bonusCategory->setUser($user);
        $bonusCategory->setName('Премия');
        $bonusCategory->setType('income');
        $manager->persist($bonusCategory);

        // Create category keywords
        $foodKeywords = ['еда', 'продукты', 'магазин', 'супермаркет', 'пятерочка', 'перекресток', 'магнит', 'ашан', 'продуктовый', 'готовая еда'];
        foreach ($foodKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($foodCategory);
            $manager->persist($categoryKeyword);
        }

        $cafeKeywords = ['кафе', 'ресторан', 'столовая', 'кофейня', 'бар', 'кофе', 'обед', 'ланч', 'бизнес-ланч'];
        foreach ($cafeKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($cafeCategory);
            $manager->persist($categoryKeyword);
        }

        $transportKeywords = ['такси', 'метро', 'автобус', 'трамвай', 'маршрутка', 'транспорт', 'проезд', 'uber', 'яндекс.такси', 'ситимобил'];
        foreach ($transportKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($transportCategory);
            $manager->persist($categoryKeyword);
        }

        $salaryKeywords = ['зп', 'зарплата', 'аванс', 'получка'];
        foreach ($salaryKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($salaryCategory);
            $manager->persist($categoryKeyword);
        }

        $bonusKeywords = ['премия', 'бонус', 'квартальная', 'годовая'];
        foreach ($bonusKeywords as $keyword) {
            $categoryKeyword = new CategoryKeyword();
            $categoryKeyword->setKeyword(mb_strtolower($keyword));
            $categoryKeyword->setUserCategory($bonusCategory);
            $manager->persist($categoryKeyword);
        }

        // Create test spreadsheet for current month
        $now = new \DateTime();
        $spreadsheet = new UserSpreadsheet();
        $spreadsheet->setUser($user);
        $spreadsheet->setSpreadsheetId('test_spreadsheet_id');
        $spreadsheet->setTitle('Бюджет');
        $spreadsheet->setMonth((int) $now->format('n'));
        $spreadsheet->setYear((int) $now->format('Y'));
        $manager->persist($spreadsheet);

        // Create test spreadsheet for December 2024
        $december2024 = new UserSpreadsheet();
        $december2024->setUser($user);
        $december2024->setSpreadsheetId('test_spreadsheet_id');
        $december2024->setTitle('Бюджет');
        $december2024->setMonth(12);
        $december2024->setYear(2024);
        $manager->persist($december2024);

        // Create test spreadsheet for January 2025
        $january2025 = new UserSpreadsheet();
        $january2025->setUser($user);
        $january2025->setSpreadsheetId('test_spreadsheet_id');
        $january2025->setTitle('Бюджет');
        $january2025->setMonth(1);
        $january2025->setYear(2025);
        $manager->persist($january2025);

        $manager->flush();
    }
}
