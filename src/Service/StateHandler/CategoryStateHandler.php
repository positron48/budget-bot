<?php

namespace App\Service\StateHandler;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\CategoryService;
use Psr\Log\LoggerInterface;

class CategoryStateHandler implements StateHandlerInterface
{
    private const SUPPORTED_STATES = [
        'WAITING_CATEGORIES_ACTION',
        'WAITING_CATEGORY_NAME',
        'WAITING_CATEGORY_TO_DELETE',
    ];

    public function __construct(
        private readonly UserRepository $userRepository,
        private readonly CategoryService $categoryService,
        private readonly LoggerInterface $logger,
    ) {
    }

    public function supports(string $state): bool
    {
        return in_array($state, self::SUPPORTED_STATES, true);
    }

    public function handle(int $chatId, User $user, string $message): void
    {
        $state = $user->getState();

        if ('WAITING_CATEGORIES_ACTION' === $state) {
            $this->handleCategoriesAction($chatId, $user, $message);

            return;
        }

        if ('WAITING_CATEGORY_NAME' === $state) {
            $this->handleCategoryName($chatId, $user, $message);

            return;
        }

        if ('WAITING_CATEGORY_TO_DELETE' === $state) {
            $this->handleCategoryToDelete($chatId, $user, $message);
        }
    }

    private function handleCategoriesAction(int $chatId, User $user, string $text): void
    {
        $this->logger->info('Handling categories action', [
            'chat_id' => $chatId,
            'text' => $text,
        ]);

        switch ($text) {
            case 'Категории расходов':
                $categories = $this->categoryService->getCategories(false, $user);
                $this->sendMessage($chatId, 'Категории расходов:'.PHP_EOL.implode(PHP_EOL, $categories));
                break;
            case 'Категории доходов':
                $categories = $this->categoryService->getCategories(true, $user);
                $this->sendMessage($chatId, 'Категории доходов:'.PHP_EOL.implode(PHP_EOL, $categories));
                break;
            case 'Добавить категорию':
                $user->setState('WAITING_CATEGORY_NAME');
                $this->userRepository->save($user, true);
                $keyboard = [
                    ['text' => 'Категория расходов'],
                    ['text' => 'Категория доходов'],
                ];
                $this->sendMessage($chatId, 'Выберите тип категории:', $keyboard);
                break;
            case 'Удалить категорию':
                $user->setState('WAITING_CATEGORY_TO_DELETE');
                $this->userRepository->save($user, true);
                $keyboard = [];
                $expenseCategories = $this->categoryService->getCategories(false, $user);
                $incomeCategories = $this->categoryService->getCategories(true, $user);
                foreach ($expenseCategories as $category) {
                    $keyboard[] = ['text' => $category];
                }
                foreach ($incomeCategories as $category) {
                    $keyboard[] = ['text' => $category];
                }
                $this->sendMessage($chatId, 'Выберите категорию для удаления:', $keyboard);
                break;
            default:
                $this->logger->warning('Unknown categories action', [
                    'chat_id' => $chatId,
                    'text' => $text,
                ]);
                $this->sendMessage($chatId, 'Неизвестное действие. Пожалуйста, используйте кнопки меню.');
                break;
        }

        if (!in_array($text, ['Добавить категорию', 'Удалить категорию'], true)) {
            $user->setState('');
            $this->userRepository->save($user, true);
        }
    }

    private function handleCategoryName(int $chatId, User $user, string $text): void
    {
        $this->logger->info('Handling category name', [
            'chat_id' => $chatId,
            'text' => $text,
        ]);

        $tempData = $user->getTempData();
        if (!isset($tempData['category_type'])) {
            if ('Категория расходов' === $text) {
                $tempData['category_type'] = 'expense';
                $user->setTempData($tempData);
                $this->userRepository->save($user, true);
                $this->sendMessage($chatId, 'Введите название категории расходов:');

                return;
            }

            if ('Категория доходов' === $text) {
                $tempData['category_type'] = 'income';
                $user->setTempData($tempData);
                $this->userRepository->save($user, true);
                $this->sendMessage($chatId, 'Введите название категории доходов:');

                return;
            }

            $keyboard = [
                ['text' => 'Категория расходов'],
                ['text' => 'Категория доходов'],
            ];
            $this->sendMessage($chatId, 'Пожалуйста, выберите тип категории:', $keyboard);

            return;
        }

        $isIncome = 'income' === $tempData['category_type'];
        $this->categoryService->addUserCategory($user, $text, $isIncome);
        $this->sendMessage($chatId, 'Категория успешно добавлена');

        $user->setState('');
        $user->setTempData([]);
        $this->userRepository->save($user, true);
    }

    private function handleCategoryToDelete(int $chatId, User $user, string $text): void
    {
        $this->logger->info('Handling category to delete', [
            'chat_id' => $chatId,
            'text' => $text,
        ]);

        $expenseCategories = $this->categoryService->getCategories(false, $user);
        $incomeCategories = $this->categoryService->getCategories(true, $user);

        $isIncome = in_array($text, $incomeCategories, true);
        $isExpense = in_array($text, $expenseCategories, true);

        if (!$isIncome && !$isExpense) {
            $keyboard = [];
            foreach ($expenseCategories as $category) {
                $keyboard[] = ['text' => $category];
            }
            foreach ($incomeCategories as $category) {
                $keyboard[] = ['text' => $category];
            }
            $this->sendMessage($chatId, 'Категория не найдена. Выберите категорию из списка:', $keyboard);

            return;
        }

        $this->categoryService->removeUserCategory($user, $text, $isIncome);
        $this->sendMessage($chatId, 'Категория успешно удалена');

        $user->setState('');
        $this->userRepository->save($user, true);
    }

    /**
     * @param array<int, array<string, string>>|null $keyboard
     */
    private function sendMessage(int $chatId, string $text, ?array $keyboard = null): void
    {
        $data = [
            'chat_id' => $chatId,
            'text' => $text,
            'parse_mode' => 'HTML',
        ];

        if ($keyboard) {
            $data['reply_markup'] = json_encode([
                'keyboard' => array_map(
                    static function (array $button) {
                        return [$button];
                    },
                    $keyboard
                ),
                'resize_keyboard' => true,
                'one_time_keyboard' => true,
            ]);
        }

        $this->logger->info('Sending message to Telegram API', [
            'request' => $data,
        ]);

        $response = \Longman\TelegramBot\Request::sendMessage($data);

        $this->logger->info('Received response from Telegram API', [
            'response' => [
                'ok' => $response->isOk(),
                'result' => $response->getResult(),
                'description' => $response->getDescription(),
                'error_code' => $response->getErrorCode(),
            ],
        ]);
    }
}
