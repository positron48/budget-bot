<?php

namespace App\Service\StateHandler;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\CategoryService;
use App\Service\TelegramApiServiceInterface;
use App\Service\TransactionHandler;
use Psr\Log\LoggerInterface;

class CategoryStateHandler implements StateHandlerInterface
{
    private const SUPPORTED_STATES = [
        'WAITING_CATEGORIES_ACTION',
        'WAITING_CATEGORY_SELECTION',
        'WAITING_CATEGORY_MAPPING',
    ];

    public function __construct(
        private readonly UserRepository $userRepository,
        private readonly CategoryService $categoryService,
        private readonly TransactionHandler $transactionHandler,
        private readonly LoggerInterface $logger,
        private readonly TelegramApiServiceInterface $telegramApi,
    ) {
    }

    public function supports(string $state): bool
    {
        return in_array($state, self::SUPPORTED_STATES, true);
    }

    public function handle(int $chatId, User $user, string $message): bool
    {
        $state = $user->getState();
        $tempData = $user->getTempData();

        if ('WAITING_CATEGORY_SELECTION' === $state && isset($tempData['pending_transaction'])) {
            $this->handleCategorySelection($chatId, $user, $message);

            return true;
        }

        if ('WAITING_CATEGORY_MAPPING' === $state && isset($tempData['pending_transaction'])) {
            $this->handleCategoryMapping($chatId, $user, $message);

            return true;
        }

        if ('WAITING_CATEGORIES_ACTION' === $state) {
            $this->handleCategoriesAction($chatId, $user, $message);

            return true;
        }

        return false;
    }

    private function handleCategorySelection(int $chatId, User $user, string $message): void
    {
        $tempData = $user->getTempData();
        $transaction = $tempData['pending_transaction'];

        if ('Добавить сопоставление' === $message) {
            $user->setState('WAITING_CATEGORY_MAPPING');
            $this->userRepository->save($user, true);

            $this->sendMessage(
                $chatId,
                sprintf(
                    'Введите сопоставление в формате "слово = категория". Например: "%s = Питание"',
                    $transaction['description']
                )
            );

            return;
        }

        // Check if selected category exists
        $categories = $this->categoryService->getCategories($transaction['isIncome'], $user);
        $categories = array_unique($categories);

        if (!in_array($message, $categories, true)) {
            $this->sendMessage(
                $chatId,
                sprintf('Категория "%s" не найдена. Выберите категорию из списка:', $message),
                array_merge($categories, ['Добавить сопоставление'])
            );

            return;
        }

        // Add mapping for the full description
        $this->categoryService->addKeywordToCategory(
            mb_strtolower($transaction['description']),
            $message,
            $transaction['isIncome'] ? 'income' : 'expense',
            $user
        );

        // Add transaction with selected category
        $this->transactionHandler->addTransaction($chatId, $user, [
            'date' => $transaction['date'] instanceof \DateTime ? $transaction['date'] : new \DateTime(),
            'amount' => $transaction['amount'],
            'description' => $transaction['description'],
            'isIncome' => $transaction['isIncome'],
        ], $message);

        // Clear state and temp data
        $user->setState('');
        $user->setTempData([]);
        $this->userRepository->save($user, true);
    }

    private function handleCategoryMapping(int $chatId, User $user, string $message): void
    {
        $tempData = $user->getTempData();
        $transaction = $tempData['pending_transaction'];

        // Parse mapping
        $parts = array_map('trim', explode('=', $message));
        if (2 !== count($parts)) {
            $this->sendMessage($chatId, 'Неверный формат. Используйте: слово = категория');

            return;
        }

        $keyword = mb_strtolower($parts[0]);
        $categoryName = $parts[1];

        // Check if category exists
        $categories = $this->categoryService->getCategories($transaction['isIncome'], $user);
        if (!in_array($categoryName, $categories, true)) {
            $this->sendMessage(
                $chatId,
                sprintf(
                    'Категория "%s" не найдена. Доступные категории:%s%s',
                    $categoryName,
                    PHP_EOL,
                    implode(PHP_EOL, $categories)
                )
            );

            return;
        }

        // Add mapping
        $this->categoryService->addKeywordToCategory(
            $keyword,
            $categoryName,
            $transaction['isIncome'] ? 'income' : 'expense',
            $user
        );

        $this->sendMessage($chatId, sprintf('Добавлено сопоставление: "%s" → "%s"', $keyword, $categoryName));

        // Try to detect category again
        $category = $this->categoryService->detectCategory(
            $transaction['description'],
            $transaction['isIncome'] ? 'income' : 'expense',
            $user
        );

        if ($category) {
            // Add transaction with detected category
            $this->transactionHandler->addTransaction($chatId, $user, [
                'date' => $transaction['date'] instanceof \DateTime ? $transaction['date'] : new \DateTime(),
                'amount' => $transaction['amount'],
                'description' => $transaction['description'],
                'isIncome' => $transaction['isIncome'],
            ], $category);

            // Clear state and temp data
            $user->setState('');
            $user->setTempData([]);
            $this->userRepository->save($user, true);
        } else {
            // Show categories list again
            $user->setState('WAITING_CATEGORY_SELECTION');
            $this->userRepository->save($user, true);

            $this->sendMessage(
                $chatId,
                sprintf(
                    'Категория для "%s" все еще не определена. Выберите категорию из списка или добавьте еще одно сопоставление:',
                    $transaction['description']
                ),
                array_merge($categories, ['Добавить сопоставление'])
            );
        }
    }

    /**
     * @param array<string>|null $keyboard
     */
    private function sendMessage(int $chatId, string $text, ?array $keyboard = null): void
    {
        try {
            if (null !== $keyboard) {
                $this->telegramApi->sendMessageWithKeyboard($chatId, $text, $keyboard);
            } else {
                $this->telegramApi->sendMessage([
                    'chat_id' => $chatId,
                    'text' => $text,
                    'parse_mode' => 'HTML',
                ]);
            }
        } catch (\Exception $e) {
            $this->logger->error('Failed to send message: {error}', [
                'error' => $e->getMessage(),
                'exception' => $e,
                'chat_id' => $chatId,
                'text' => $text,
            ]);
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
                $this->sendMessage($chatId, 'Категории расходов:'.PHP_EOL.implode(PHP_EOL, array_unique($categories)));
                $user->setState('');
                $this->userRepository->save($user, true);
                break;
            case 'Категории доходов':
                $categories = $this->categoryService->getCategories(true, $user);
                $this->sendMessage($chatId, 'Категории доходов:'.PHP_EOL.implode(PHP_EOL, array_unique($categories)));
                $user->setState('');
                $this->userRepository->save($user, true);
                break;
            default:
                $this->sendMessage($chatId, 'Неизвестное действие');
                $user->setState('');
                $this->userRepository->save($user, true);
        }
    }
}
