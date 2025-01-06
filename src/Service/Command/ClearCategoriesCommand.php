<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\CategoryService;
use Psr\Log\LoggerInterface;

class ClearCategoriesCommand extends AbstractCommand
{
    public function __construct(
        UserRepository $userRepository,
        LoggerInterface $logger,
        private readonly CategoryService $categoryService,
    ) {
        parent::__construct($userRepository, $logger);
    }

    public function getName(): string
    {
        return '/clear_categories';
    }

    protected function handleCommand(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $this->sendMessage($chatId, 'Пожалуйста, начните с команды /start');

            return;
        }

        try {
            $expenseCount = count($this->categoryService->getCategories(false, $user));
            $incomeCount = count($this->categoryService->getCategories(true, $user));

            $this->categoryService->clearUserCategories($user);

            $this->sendMessage(
                $chatId,
                sprintf(
                    'Пользовательские категории очищены:%s- Расходы: %d%s- Доходы: %d',
                    PHP_EOL,
                    $expenseCount,
                    PHP_EOL,
                    $incomeCount
                )
            );
        } catch (\Exception $e) {
            $this->logger->error('Failed to clear categories: '.$e->getMessage(), [
                'chat_id' => $chatId,
                'exception' => $e,
            ]);
            $this->sendMessage($chatId, 'Не удалось очистить категории. Попробуйте еще раз.');
        }
    }
}
