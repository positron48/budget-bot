<?php

namespace App\Tests\Integration;

use App\Repository\UserCategoryRepository;
use App\Repository\UserRepository;
use App\Tests\Integration\DataFixtures\TestFixtures;
use Doctrine\Common\DataFixtures\Executor\ORMExecutor;
use Doctrine\Common\DataFixtures\Loader;
use Doctrine\Common\DataFixtures\Purger\ORMPurger;

class TransactionIntegrationTest extends AbstractBotIntegrationTestCase
{
    private const TEST_CHAT_ID = 123456;
    private UserRepository $userRepository;
    private UserCategoryRepository $categoryRepository;

    protected function setUp(): void
    {
        parent::setUp();

        $container = self::getContainer();
        $this->userRepository = $container->get(UserRepository::class);
        $this->categoryRepository = $container->get(UserCategoryRepository::class);

        // Load test fixtures using DoctrineFixturesBundle
        $loader = new Loader();
        $loader->addFixture(new TestFixtures());

        $executor = new ORMExecutor($this->entityManager, new ORMPurger());
        $executor->execute($loader->getFixtures());
    }

    public function testCategoryListDisplay(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Check categories list
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Выберите действие');

        // Check expense categories
        $this->executeCommand('Категории расходов', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Питание');
        $this->assertLastMessageContains('Кафе/Ресторан');
        $this->assertLastMessageContains('Транспорт');

        // Check income categories
        $this->executeCommand('/categories', self::TEST_CHAT_ID);
        $this->executeCommand('Категории доходов', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Зарплата');
        $this->assertLastMessageContains('Премия');

        // Verify that all expected categories are present
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        $expenseCategories = $this->categoryRepository->findBy([
            'user' => $user,
            'type' => 'expense',
        ]);
        $this->assertCount(3, $expenseCategories);

        $incomeCategories = $this->categoryRepository->findBy([
            'user' => $user,
            'type' => 'income',
        ]);
        $this->assertCount(2, $incomeCategories);
    }

    public function testAddExpenseWithExistingCategory(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add expense with a keyword that matches existing category
        $this->executeCommand('1500 продукты пятерочка', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 1500');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: продукты пятерочка');
        $this->assertLastMessageContains('Категория: Питание');

        // Add another expense with a different keyword for the same category
        $this->executeCommand('2000 еда', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 2000');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: еда');
        $this->assertLastMessageContains('Категория: Питание');

        // Add expense for cafe category
        $this->executeCommand('3000 кафе обед', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 3000');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: кафе обед');
        $this->assertLastMessageContains('Категория: Кафе/Ресторан');

        // Add expense for transport category
        $this->executeCommand('500 такси', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 500');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: такси');
        $this->assertLastMessageContains('Категория: Транспорт');
    }

    public function testAddExpenseWithUnknownCategory(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add expense with unknown category keyword
        $this->executeCommand('1500 неизвестная категория', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Не удалось определить категорию для "неизвестная категория"');
        $this->assertLastMessageContains('Выберите категорию из списка');

        // Select category from list
        $this->executeCommand('Питание', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 1500');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: неизвестная категория');
        $this->assertLastMessageContains('Категория: Питание');

        // Verify that the keyword is now mapped to the category
        $this->executeCommand('/map неизвестная категория', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Описание "неизвестная категория" соответствует категории "Питание"');

        // Try adding another expense with the same keyword
        $this->executeCommand('2000 неизвестная категория', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 2000');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: неизвестная категория');
        $this->assertLastMessageContains('Категория: Питание');
    }

    public function testAddIncome(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add income with a keyword that matches existing category
        $this->executeCommand('+50000 зарплата', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Доход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 50000');
        $this->assertLastMessageContains('Тип: доход');
        $this->assertLastMessageContains('Описание: зарплата');
        $this->assertLastMessageContains('Категория: Зарплата');

        // Add another income with a different keyword for the same category
        $this->executeCommand('+10000 зп', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Доход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 10000');
        $this->assertLastMessageContains('Тип: доход');
        $this->assertLastMessageContains('Описание: зп');
        $this->assertLastMessageContains('Категория: Зарплата');

        // Add income with unknown category
        $this->executeCommand('+5000 подработка', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Не удалось определить категорию для "подработка"');
        $this->assertLastMessageContains('Выберите категорию из списка');

        // Select category from list
        $this->executeCommand('Зарплата', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Доход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 5000');
        $this->assertLastMessageContains('Тип: доход');
        $this->assertLastMessageContains('Описание: подработка');
        $this->assertLastMessageContains('Категория: Зарплата');

        // Add another income with the same keyword (should be automatically mapped)
        $this->executeCommand('+7500 подработка', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Доход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 7500');
        $this->assertLastMessageContains('Тип: доход');
        $this->assertLastMessageContains('Описание: подработка');
        $this->assertLastMessageContains('Категория: Зарплата');
    }

    public function testAddExpenseWithDifferentDates(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add expense with explicit date
        $this->executeCommand('12.01.2025 1500 продукты', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Дата: 12.01.2025');
        $this->assertLastMessageContains('Сумма: 1500');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: продукты');
        $this->assertLastMessageContains('Категория: Питание');

        // Add expense with short date format
        $this->executeCommand('15.01 2000 кафе', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Дата: 15.01.2025');
        $this->assertLastMessageContains('Сумма: 2000');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: кафе');
        $this->assertLastMessageContains('Категория: Кафе/Ресторан');

        // Add expense with "вчера"
        $this->executeCommand('вчера 500 такси', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        // Дата будет вчерашней, поэтому не проверяем конкретное значение
        $this->assertLastMessageContains('Сумма: 500');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: такси');
        $this->assertLastMessageContains('Категория: Транспорт');
    }

    public function testAddExpenseWithDecimalAmounts(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add expense with decimal point
        $this->executeCommand('1500.50 продукты', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 1500.50');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: продукты');
        $this->assertLastMessageContains('Категория: Питание');

        // Add expense with decimal comma
        $this->executeCommand('2000,75 кафе', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 2000.75');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: кафе');
        $this->assertLastMessageContains('Категория: Кафе/Ресторан');

        // Add expense with date and decimal amount
        $this->executeCommand('15.01 99.90 такси', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        $this->assertLastMessageContains('Дата: 15.01.2025');
        $this->assertLastMessageContains('Сумма: 99.90');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: такси');
        $this->assertLastMessageContains('Категория: Транспорт');

        // Add expense with "вчера" and decimal comma
        $this->executeCommand('вчера 199,99 продукты', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Расход успешно добавлен');
        // Дата будет вчерашней, поэтому не проверяем конкретное значение
        $this->assertLastMessageContains('Сумма: 199.99');
        $this->assertLastMessageContains('Тип: расход');
        $this->assertLastMessageContains('Описание: продукты');
        $this->assertLastMessageContains('Категория: Питание');
    }

    public function testAddIncomeWithDifferentFormats(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Add income with date and decimal point
        $this->executeCommand('15.01 +50000.50 зарплата', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Доход успешно добавлен');
        $this->assertLastMessageContains('Дата: 15.01.2025');
        $this->assertLastMessageContains('Сумма: 50000.50');
        $this->assertLastMessageContains('Тип: доход');
        $this->assertLastMessageContains('Описание: зарплата');
        $this->assertLastMessageContains('Категория: Зарплата');

        // Add income with decimal comma and known category
        $this->executeCommand('+1234,56 премия', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Доход успешно добавлен');
        $this->assertLastMessageContains('Сумма: 1234.56');
        $this->assertLastMessageContains('Тип: доход');
        $this->assertLastMessageContains('Описание: премия');
        $this->assertLastMessageContains('Категория: Премия');

        // Add income with full date format and unknown category
        $this->executeCommand('31.12.2024 +10000 подработка', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Не удалось определить категорию для "подработка"');
        $this->assertLastMessageContains('Выберите категорию из списка');

        // Select category from list
        $this->executeCommand('Зарплата', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('Доход успешно добавлен');
        $this->assertLastMessageContains('Дата: 31.12.2024');
        $this->assertLastMessageContains('Сумма: 10000');
        $this->assertLastMessageContains('Тип: доход');
        $this->assertLastMessageContains('Описание: подработка');
        $this->assertLastMessageContains('Категория: Зарплата');
    }

    public function testHandleSpreadsheetNotFound(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Try to add expense for a month without spreadsheet
        $this->executeCommand('01.03.2025 1500 продукты', self::TEST_CHAT_ID);
        $this->assertLastMessageContains('У вас нет таблицы за Март 2025');
        $this->assertLastMessageContains('Пожалуйста, добавьте её с помощью команды /add');
    }

    public function testHandleNullSpreadsheetId(): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Create spreadsheet with empty ID
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        // Remove any existing spreadsheets for January 2025
        $existingSpreadsheets = $this->entityManager->getRepository(\App\Entity\UserSpreadsheet::class)
            ->findBy([
                'user' => $user,
                'year' => 2025,
                'month' => 1,
            ]);

        foreach ($existingSpreadsheets as $spreadsheet) {
            $this->entityManager->remove($spreadsheet);
        }
        $this->entityManager->flush();

        // Create a new spreadsheet with empty ID
        $spreadsheet = new \App\Entity\UserSpreadsheet();
        $spreadsheet->setUser($user);
        $spreadsheet->setYear(2025);
        $spreadsheet->setMonth(1);
        $spreadsheet->setSpreadsheetId('');
        $spreadsheet->setTitle('Test Spreadsheet');
        $this->entityManager->persist($spreadsheet);
        $this->entityManager->flush();

        // Clear entity manager to force reload
        $this->entityManager->clear();

        // Get the transaction handler service
        $transactionHandler = self::getContainer()->get(\App\Service\TransactionHandler::class);

        // Try to add transaction directly - should throw RuntimeException
        $this->expectException(\RuntimeException::class);
        $this->expectExceptionMessage('Spreadsheet ID is null');

        $transactionHandler->handle(self::TEST_CHAT_ID, $user, [
            'date' => new \DateTime('2025-01-01'),
            'amount' => 1500.0,
            'description' => 'продукты',
            'isIncome' => false,
        ]);
    }

    /**
     * @dataProvider monthNameProvider
     */
    public function testMonthNameTranslations(int $month, string $expectedName): void
    {
        // Execute /start command to ensure user exists
        $this->executeCommand('/start', self::TEST_CHAT_ID);

        // Remove any existing spreadsheets for the test month
        $user = $this->userRepository->findOneBy(['telegramId' => self::TEST_CHAT_ID]);
        $this->assertNotNull($user);

        $spreadsheets = $this->entityManager->getRepository(\App\Entity\UserSpreadsheet::class)
            ->findBy([
                'user' => $user,
                'year' => 2025,
                'month' => $month,
            ]);

        foreach ($spreadsheets as $spreadsheet) {
            $this->entityManager->remove($spreadsheet);
        }
        $this->entityManager->flush();

        // Clear entity manager to force reload
        $this->entityManager->clear();

        // Try to add expense for each month
        $this->executeCommand(sprintf('01.%02d.2025 1500 продукты', $month), self::TEST_CHAT_ID);
        $this->assertLastMessageContains(sprintf('У вас нет таблицы за %s 2025', $expectedName));
    }

    /**
     * @return array<string, array{month: int, name: string}>
     */
    public function monthNameProvider(): array
    {
        return [
            'January' => ['month' => 1, 'name' => 'Январь'],
            'February' => ['month' => 2, 'name' => 'Февраль'],
            'March' => ['month' => 3, 'name' => 'Март'],
            'April' => ['month' => 4, 'name' => 'Апрель'],
            'May' => ['month' => 5, 'name' => 'Май'],
            'June' => ['month' => 6, 'name' => 'Июнь'],
            'July' => ['month' => 7, 'name' => 'Июль'],
            'August' => ['month' => 8, 'name' => 'Август'],
            'September' => ['month' => 9, 'name' => 'Сентябрь'],
            'October' => ['month' => 10, 'name' => 'Октябрь'],
            'November' => ['month' => 11, 'name' => 'Ноябрь'],
            'December' => ['month' => 12, 'name' => 'Декабрь'],
        ];
    }
}
