<?php

namespace App\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use Longman\TelegramBot\Entities\Message;
use Longman\TelegramBot\Entities\Update;
use Longman\TelegramBot\Exception\TelegramException;
use Longman\TelegramBot\Request;
use Longman\TelegramBot\Telegram;
use Psr\Log\LoggerInterface;

class TelegramBotService
{
    private GoogleSheetsService $sheetsService;
    private MessageParserService $messageParser;
    private UserRepository $userRepository;
    private CategoryService $categoryService;
    private LoggerInterface $logger;

    public function __construct(
        string $botToken,
        string $botUsername,
        GoogleSheetsService $sheetsService,
        MessageParserService $messageParser,
        UserRepository $userRepository,
        CategoryService $categoryService,
        LoggerInterface $logger,
    ) {
        $this->sheetsService = $sheetsService;
        $this->messageParser = $messageParser;
        $this->userRepository = $userRepository;
        $this->categoryService = $categoryService;
        $this->logger = $logger;

        try {
            $telegram = new Telegram($botToken, $botUsername);
            Request::initialize($telegram);
        } catch (TelegramException $e) {
            $this->logger->error('Failed to initialize Telegram bot: '.$e->getMessage(), [
                'exception' => $e,
            ]);
            throw $e;
        }
    }

    /**
     * @param array<string, mixed> $updateData
     */
    public function handleUpdate(array $updateData): void
    {
        try {
            $this->logger->info('Processing update', ['update' => $updateData]);
            $update = new Update($updateData);
            $message = $update->getMessage();

            if (!$message instanceof Message) {
                $this->logger->info('Update does not contain a message');

                return;
            }

            $chatId = $message->getChat()->getId();
            $text = $message->getText();

            if (null === $text) {
                $this->logger->info('Message does not contain text', ['chat_id' => $chatId]);

                return;
            }

            $this->logger->info('Processing message', [
                'chat_id' => $chatId,
                'text' => $text,
            ]);

            if (str_starts_with($text, '/remove ')) {
                $this->logger->info('Handling /remove command', [
                    'chat_id' => $chatId,
                    'text' => $text,
                ]);
                $this->handleRemoveCommand($chatId, substr($text, 8));

                return;
            }

            switch ($text) {
                case '/start':
                    $this->logger->info('Handling /start command', ['chat_id' => $chatId]);
                    $this->handleStartCommand($chatId);

                    return;
                case '/list':
                    $this->logger->info('Handling /list command', ['chat_id' => $chatId]);
                    $this->handleListCommand($chatId);

                    return;
                case '/add':
                    $this->logger->info('Handling /add command', ['chat_id' => $chatId]);
                    $this->handleAddCommand($chatId, $text);

                    return;
                case '/categories':
                    $this->logger->info('Handling /categories command', ['chat_id' => $chatId]);
                    $this->handleCategoriesCommand($chatId);

                    return;
            }

            // Handle regular message
            $this->handleMessage($chatId, $text);
        } catch (TelegramException $e) {
            $this->logger->error('Error handling update: '.$e->getMessage(), [
                'exception' => $e,
                'update' => $updateData,
            ]);
        }
    }

    private function handleStartCommand(int $chatId): void
    {
        $user = $this->userRepository->findByTelegramId($chatId);

        if (!$user) {
            $user = new User();
            $user->setTelegramId($chatId);

            $this->userRepository->save($user, true);
            $this->logger->info('New user registered', [
                'telegram_id' => $chatId,
            ]);
        }

        $this->sendMessage($chatId, 'Привет! Я помогу вести учет доходов и расходов в Google Таблицах. '.
            'Отправляйте сообщения в формате: "[дата] [+]сумма описание"'.
            "\n\nДоступные команды:\n".
            "/list - список доступных таблиц\n".
            "/add - добавить таблицу\n".
            '/categories - управление категориями'
        );
    }

    private function handleListCommand(int $chatId): void
    {
        $user = $this->userRepository->findByTelegramId($chatId);
        if (!$user) {
            $this->logger->warning('User not found for /list command', ['chat_id' => $chatId]);
            $this->sendMessage($chatId, 'Пользалуйста, используйте /start для начала работы.');

            return;
        }

        $spreadsheets = $this->sheetsService->getSpreadsheetsList($user);
        $this->logger->info('Retrieved spreadsheets list', [
            'chat_id' => $chatId,
            'count' => count($spreadsheets),
        ]);

        if (empty($spreadsheets)) {
            $this->sendMessage($chatId, 'У вас пока нет подключенных таблиц. Используйте /add чтобы добавить таблицу.');

            return;
        }

        $message = "Ваши таблицы:\n\n";
        foreach ($spreadsheets as $spreadsheet) {
            $message .= sprintf(
                "%s %d: %s\n",
                $spreadsheet['month'],
                $spreadsheet['year'],
                $spreadsheet['url']
            );
        }

        $this->sendMessage($chatId, $message);
    }

    private function handleAddCommand(int $chatId, string $text): void
    {
        $user = $this->userRepository->findByTelegramId($chatId);
        if (!$user) {
            $this->logger->warning('User not found for /add command', ['chat_id' => $chatId]);
            $this->sendMessage($chatId, 'Пользователь не найден');

            return;
        }

        $parts = explode(' ', $text);
        if (1 === count($parts)) {
            $this->logger->info('Setting user state to WAITING_SPREADSHEET_ID', ['chat_id' => $chatId]);
            $this->sendMessage($chatId, 'Отправьте ID или URL таблицы Google Sheets');
            $this->userRepository->setUserState($user, 'WAITING_SPREADSHEET_ID');

            return;
        }

        $spreadsheetId = $parts[1];
        $this->logger->info('Processing spreadsheet ID from command', [
            'chat_id' => $chatId,
            'spreadsheet_id' => $spreadsheetId,
        ]);
        $this->handleSpreadsheetId($chatId, $spreadsheetId);
    }

    private function handleSpreadsheetId(int $chatId, string $spreadsheetId): void
    {
        $user = $this->userRepository->findByTelegramId($chatId);
        if (!$user) {
            $this->logger->warning('User not found for spreadsheet handling', ['chat_id' => $chatId]);

            return;
        }

        try {
            $this->logger->info('Validating spreadsheet access', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
            ]);
            $spreadsheetId = $this->sheetsService->handleSpreadsheetId($spreadsheetId);

            $this->logger->info('Setting user state to WAITING_MONTH', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
            ]);

            $user->setTempData(['spreadsheet_id' => $spreadsheetId]);
            $this->userRepository->save($user, true);
            $this->userRepository->setUserState($user, 'WAITING_MONTH');

            $months = $this->getMonthsList();
            $keyboard = [];
            $row = [];
            $i = 0;

            foreach ($months as $label => $month) {
                $row[] = ['text' => $label];
                ++$i;
                if (0 === $i % 3) {
                    $keyboard[] = $row;
                    $row = [];
                }
            }

            if (!empty($row)) {
                $keyboard[] = $row;
            }

            $this->logger->info('Sending month selection keyboard', [
                'chat_id' => $chatId,
                'months' => array_keys($months),
            ]);
            $this->sendMessage(
                $chatId,
                'Выберите месяц из списка или введите в формате "Месяц Год" (например: Январь 2025):',
                ['reply_markup' => json_encode([
                    'keyboard' => $keyboard,
                    'one_time_keyboard' => true,
                    'resize_keyboard' => true,
                ])]
            );
        } catch (\RuntimeException $e) {
            $this->logger->warning('Failed to handle spreadsheet', [
                'chat_id' => $chatId,
                'spreadsheet_id' => $spreadsheetId,
                'error' => $e->getMessage(),
            ]);
            $this->sendMessage($chatId, $e->getMessage());
            $this->userRepository->clearUserState($user);
        }
    }

    private function handleCategoriesCommand(int $chatId): void
    {
        $user = $this->userRepository->findByTelegramId($chatId);
        if (!$user) {
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, используйте /start для начала работы.',
            ]);

            return;
        }

        $expenseCategories = $this->categoryService->getCategories(false, $user);
        $incomeCategories = $this->categoryService->getCategories(true, $user);

        $message = "Категории расходов:\n".implode("\n", $expenseCategories);
        $message .= "\n\nКатегории доходов:\n".implode("\n", $incomeCategories);

        Request::sendMessage([
            'chat_id' => $chatId,
            'text' => $message,
        ]);
    }

    private function handleMessage(int $chatId, string $text): void
    {
        $user = $this->userRepository->findByTelegramId($chatId);
        if (!$user) {
            $this->logger->warning('User not found for message handling', ['chat_id' => $chatId]);

            return;
        }

        $state = $user->getState();
        $tempData = $user->getTempData();
        $this->logger->info('Processing message with state', [
            'chat_id' => $chatId,
            'state' => $state,
            'temp_data' => $tempData,
            'text' => $text,
        ]);

        if ('WAITING_SPREADSHEET_ID' === $state) {
            $this->logger->info('Handling spreadsheet ID input', [
                'chat_id' => $chatId,
                'text' => $text,
            ]);
            $this->handleSpreadsheetId($chatId, $text);

            return;
        }

        if ('WAITING_MONTH' === $state) {
            $this->logger->info('Handling month selection', [
                'chat_id' => $chatId,
                'text' => $text,
                'temp_data' => $tempData,
            ]);

            // Try to parse month and year from text
            if (preg_match('/^(\p{L}+)\s+(\d{4})$/u', $text, $matches)) {
                $monthName = $matches[1];
                $year = (int) $matches[2];

                $this->logger->info('Parsed month and year from text', [
                    'chat_id' => $chatId,
                    'month_name' => $monthName,
                    'year' => $year,
                ]);

                $months = [
                    'Январь' => 1,
                    'Февраль' => 2,
                    'Март' => 3,
                    'Апрель' => 4,
                    'Май' => 5,
                    'Июнь' => 6,
                    'Июль' => 7,
                    'Август' => 8,
                    'Сентябрь' => 9,
                    'Октябрь' => 10,
                    'Ноябрь' => 11,
                    'Декабрь' => 12,
                ];

                if (!isset($months[$monthName])) {
                    $this->logger->warning('Invalid month name provided', [
                        'chat_id' => $chatId,
                        'month_name' => $monthName,
                        'valid_months' => array_keys($months),
                    ]);
                    $this->sendMessage($chatId, 'Пожалуйста, используйте корректное название месяца (например: Январь 2025)');

                    return;
                }

                $monthNumber = $months[$monthName];
            } else {
                // Try to parse from the keyboard selection
                $months = $this->getMonthsList();
                $this->logger->info('Trying to parse from keyboard selection', [
                    'chat_id' => $chatId,
                    'text' => $text,
                    'available_months' => array_keys($months),
                ]);

                if (!isset($months[$text])) {
                    $this->logger->warning('Invalid month selected', [
                        'chat_id' => $chatId,
                        'text' => $text,
                        'available_months' => array_keys($months),
                    ]);
                    $this->sendMessage($chatId, 'Пожалуйста, выберите месяц из предложенных вариантов или введите в формате "Месяц Год" (например: Январь 2025)');

                    return;
                }

                $monthNumber = $months[$text];
                if (!preg_match('/^.*\s+(\d{4})$/', $text, $matches)) {
                    $this->logger->error('Failed to parse year from month selection', [
                        'chat_id' => $chatId,
                        'text' => $text,
                    ]);
                    $this->sendMessage($chatId, 'Произошла ошибка при обработке выбранного месяца. Попробуйте ввести месяц и год вручную (например: Январь 2025)');

                    return;
                }
                $year = (int) $matches[1];

                $this->logger->info('Parsed month and year from keyboard selection', [
                    'chat_id' => $chatId,
                    'month_number' => $monthNumber,
                    'year' => $year,
                ]);
            }

            $spreadsheetId = $tempData['spreadsheet_id'] ?? null;
            if (!$spreadsheetId) {
                $this->logger->error('Missing spreadsheet_id in temp data', [
                    'chat_id' => $chatId,
                    'temp_data' => $tempData,
                    'state' => $state,
                ]);
                $this->sendMessage($chatId, 'Произошла ошибка. Попробуйте начать сначала с команды /add');
                $this->userRepository->clearUserState($user);

                return;
            }

            try {
                $this->logger->info('Adding spreadsheet for month', [
                    'chat_id' => $chatId,
                    'spreadsheet_id' => $spreadsheetId,
                    'month' => $monthNumber,
                    'year' => $year,
                ]);

                $this->sheetsService->addSpreadsheet($user, $spreadsheetId, (int) $monthNumber, $year);
                $this->sendMessage($chatId, "Таблица успешно привязана к месяцу {$this->getMonthName((int) $monthNumber)} $year");

                $this->logger->info('Successfully added spreadsheet', [
                    'chat_id' => $chatId,
                    'spreadsheet_id' => $spreadsheetId,
                    'month' => $monthNumber,
                    'year' => $year,
                ]);
            } catch (\RuntimeException $e) {
                $this->logger->error('Failed to add spreadsheet', [
                    'chat_id' => $chatId,
                    'spreadsheet_id' => $spreadsheetId,
                    'month' => $monthNumber,
                    'year' => $year,
                    'error' => $e->getMessage(),
                ]);
                $this->sendMessage($chatId, $e->getMessage());
            }

            $this->logger->info('Clearing user state', [
                'chat_id' => $chatId,
                'previous_state' => $state,
            ]);
            $this->userRepository->clearUserState($user);

            return;
        }

        // Check if the message is a spreadsheet link
        if (str_contains($text, 'docs.google.com/spreadsheets/d/')) {
            $this->logger->info('Received spreadsheet link, setting state and handling', [
                'chat_id' => $chatId,
                'text' => $text,
                'previous_state' => $state,
            ]);
            $this->userRepository->setUserState($user, 'WAITING_SPREADSHEET_ID');
            $this->handleSpreadsheetId($chatId, $text);

            return;
        }

        // Handle expense message
        try {
            $result = $this->messageParser->parseMessage($text);
            if (!$result) {
                return;
            }

            $this->logger->info('Parsed expense/income message', [
                'chat_id' => $chatId,
                'result' => $result,
            ]);

            $spreadsheet = $this->sheetsService->findSpreadsheetByDate($user, $result['date']);
            if (!$spreadsheet) {
                $this->logger->warning('Spreadsheet not found for date', [
                    'chat_id' => $chatId,
                    'date' => $result['date']->format('Y-m-d'),
                ]);
                $this->sendMessage($chatId, sprintf(
                    'У вас нет таблицы за %s %d. Пожалуйста, добавьте её с помощью команды /add',
                    $this->getMonthName((int) $result['date']->format('m')),
                    (int) $result['date']->format('Y')
                ));

                return;
            }

            $category = $this->categoryService->detectCategory(
                $result['description'],
                $result['isIncome'] ? 'income' : 'expense',
                $user
            );

            if (!$category) {
                $this->logger->warning('Category not detected', [
                    'chat_id' => $chatId,
                    'description' => $result['description'],
                    'type' => $result['isIncome'] ? 'income' : 'expense',
                ]);
                $this->sendMessage($chatId, 'Не удалось определить категорию');

                return;
            }

            $spreadsheetId = $spreadsheet->getSpreadsheetId();
            if (!$spreadsheetId) {
                $this->logger->error('Spreadsheet ID is null', [
                    'chat_id' => $chatId,
                    'spreadsheet' => $spreadsheet,
                ]);
                throw new \RuntimeException('Spreadsheet ID is null');
            }

            if ($result['isIncome']) {
                $this->logger->info('Adding income', [
                    'chat_id' => $chatId,
                    'spreadsheet_id' => $spreadsheetId,
                    'date' => $result['date']->format('d.m.Y'),
                    'amount' => $result['amount'],
                    'description' => $result['description'],
                    'category' => $category,
                ]);
                $this->sheetsService->addIncome(
                    $spreadsheetId,
                    $result['date']->format('d.m.Y'),
                    $result['amount'],
                    $result['description'],
                    $category
                );
                $this->sendMessage($chatId, 'Доход успешно добавлен');
            } else {
                $this->logger->info('Adding expense', [
                    'chat_id' => $chatId,
                    'spreadsheet_id' => $spreadsheetId,
                    'date' => $result['date']->format('d.m.Y'),
                    'amount' => $result['amount'],
                    'description' => $result['description'],
                    'category' => $category,
                ]);
                $this->sheetsService->addExpense(
                    $spreadsheetId,
                    $result['date']->format('d.m.Y'),
                    $result['amount'],
                    $result['description'],
                    $category
                );
                $this->sendMessage($chatId, 'Расход успешно добавлен');
            }
        } catch (\RuntimeException $e) {
            $this->logger->error('Failed to process expense/income', [
                'chat_id' => $chatId,
                'text' => $text,
                'error' => $e->getMessage(),
            ]);
            $this->sendMessage($chatId, $e->getMessage());
        }
    }

    /**
     * @return array<string, int>
     */
    private function getMonthsList(): array
    {
        $months = [
            1 => 'Январь',
            2 => 'Февраль',
            3 => 'Март',
            4 => 'Апрель',
            5 => 'Май',
            6 => 'Июнь',
            7 => 'Июль',
            8 => 'Август',
            9 => 'Сентябрь',
            10 => 'Октябрь',
            11 => 'Ноябрь',
            12 => 'Декабрь',
        ];

        $currentMonth = (int) date('m');
        $currentYear = (int) date('Y');
        $result = [];

        // Get next month and 5 previous months
        $nextMonth = $currentMonth + 1;
        $nextMonthYear = $currentYear;
        if ($nextMonth > 12) {
            $nextMonth = 1;
            ++$nextMonthYear;
        }

        // Add next month first
        $result[$months[$nextMonth].' '.$nextMonthYear] = $nextMonth;

        // Then add 5 previous months
        for ($i = 0; $i < 5; ++$i) {
            $month = $nextMonth - 1 - $i;
            $year = $nextMonthYear;

            if ($month <= 0) {
                $month += 12;
                --$year;
            }

            $result[$months[$month].' '.$year] = $month;
        }

        return $result;
    }

    private function getMonthName(int $month): string
    {
        $months = [
            1 => 'Январь',
            2 => 'Февраль',
            3 => 'Март',
            4 => 'Апрель',
            5 => 'Май',
            6 => 'Июнь',
            7 => 'Июль',
            8 => 'Август',
            9 => 'Сентябрь',
            10 => 'Октябрь',
            11 => 'Ноябрь',
            12 => 'Декабрь',
        ];

        return $months[$month] ?? '';
    }

    private function handleRemoveCommand(int $chatId, string $text): void
    {
        $text = trim($text); // Remove leading/trailing whitespace
        $parts = array_values(array_filter(explode(' ', $text))); // Split by spaces and remove empty parts

        if (2 !== count($parts)) {
            $this->sendMessage($chatId, 'Неверный формат команды. Используйте: /remove Месяц Год');

            return;
        }

        $monthName = trim($parts[0]);
        $year = (int) $parts[1];

        $months = [
            'Январь' => 1,
            'Февраль' => 2,
            'Март' => 3,
            'Апрель' => 4,
            'Май' => 5,
            'Июнь' => 6,
            'Июль' => 7,
            'Август' => 8,
            'Сентябрь' => 9,
            'Октябрь' => 10,
            'Ноябрь' => 11,
            'Декабрь' => 12,
        ];

        if (!isset($months[$monthName])) {
            $this->sendMessage($chatId, 'Неверный формат месяца. Используйте русское название месяца, например: Январь');

            return;
        }

        $monthNumber = $months[$monthName];

        $user = $this->userRepository->findByTelegramId($chatId);
        if (!$user) {
            $this->sendMessage($chatId, 'Пользователь не найден');

            return;
        }

        try {
            $this->sheetsService->removeSpreadsheet($user, $monthNumber, $year);
            $this->logger->info('Sending message to chat {chat_id}: {message}', [
                'chat_id' => $chatId,
                'message' => "Таблица за $monthName $year успешно удалена",
            ]);
            $this->sendMessage($chatId, "Таблица за $monthName $year успешно удалена");
        } catch (\RuntimeException $e) {
            $this->sendMessage($chatId, $e->getMessage());
        }
    }

    /**
     * @param array<string, mixed> $options
     */
    private function sendMessage(int $chatId, string $text, array $options = []): void
    {
        $params = array_merge([
            'chat_id' => $chatId,
            'text' => $text,
        ], $options);

        Request::sendMessage($params);
    }
}
