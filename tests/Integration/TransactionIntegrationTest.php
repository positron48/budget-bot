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
}
