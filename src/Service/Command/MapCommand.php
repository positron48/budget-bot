<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\CategoryRepository;
use App\Repository\UserCategoryRepository;
use App\Service\CategoryService;
use App\Service\TelegramApiServiceInterface;

class MapCommand implements CommandInterface
{
    public function __construct(
        private readonly CategoryService $categoryService,
        private readonly CategoryRepository $categoryRepository,
        private readonly UserCategoryRepository $userCategoryRepository,
        private readonly TelegramApiServiceInterface $telegramApi,
    ) {
    }

    public function getName(): string
    {
        return '/map';
    }

    public function supports(string $message): bool
    {
        return str_starts_with($message, $this->getName());
    }

    public function execute(int $chatId, ?User $user, string $message): void
    {
        if (!$user) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, начните с команды /start',
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        $parts = preg_split('/\s+/', trim($message), 2);
        if (!is_array($parts) || count($parts) < 2) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, укажите описание расхода после команды /map. Например: /map еда',
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        $description = $parts[1];
        if ('--all' === $description) {
            $this->showAllMappings($chatId, $user);

            return;
        }

        // Check if it's a mapping request (contains '=')
        if (str_contains($description, '=')) {
            $this->handleMapping($chatId, $user, $description);

            return;
        }

        $category = $this->categoryService->detectCategory($description, 'expense', $user);

        if (!$category) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf('Для описания "%s" категория не найдена', $description),
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        $this->telegramApi->sendMessage([
            'chat_id' => $chatId,
            'text' => sprintf('Описание "%s" соответствует категории "%s"', $description, $category),
            'parse_mode' => 'HTML',
        ]);
    }

    private function handleMapping(int $chatId, User $user, string $input): void
    {
        $parts = array_map('trim', explode('=', $input));
        if (2 !== count($parts)) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => 'Неверный формат. Используйте: слово = категория',
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        $keyword = mb_strtolower($parts[0]);
        $categoryName = $parts[1];

        // Get sorted list of categories
        $categories = $this->categoryService->getCategories(false, $user);
        if (!in_array($categoryName, $categories, true)) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf(
                    'Категория "%s" не найдена. Доступные категории:%s%s',
                    $categoryName,
                    PHP_EOL,
                    implode(PHP_EOL, $categories)
                ),
                'parse_mode' => 'HTML',
            ]);

            return;
        }

        $this->categoryService->addKeywordToCategory($keyword, $categoryName, 'expense', $user);
        $this->telegramApi->sendMessage([
            'chat_id' => $chatId,
            'text' => sprintf('Добавлено сопоставление: "%s" → "%s"', $keyword, $categoryName),
            'parse_mode' => 'HTML',
        ]);
    }

    private function showAllMappings(int $chatId, User $user): void
    {
        $defaultCategories = $this->categoryRepository->findByType('expense');
        $userCategories = $this->userCategoryRepository->findByUserAndType($user, 'expense');

        $message = "Справочник категорий расходов:\n\n";
        $hasAnyMappings = false;

        // Sort categories by name
        usort($defaultCategories, fn ($a, $b) => strcmp($a->getName() ?? '', $b->getName() ?? ''));
        usort($userCategories, fn ($a, $b) => strcmp($a->getName() ?? '', $b->getName() ?? ''));

        foreach ($defaultCategories as $category) {
            $keywords = [];
            foreach ($category->getKeywords() as $keyword) {
                $keywords[] = $keyword->getKeyword();
            }
            if (!empty($keywords)) {
                $hasAnyMappings = true;
                $message .= sprintf("📍 %s:\n%s\n\n", $category->getName(), implode(', ', $keywords));
            }
        }

        if (!empty($userCategories)) {
            $message .= "\nВаши категории:\n\n";
            foreach ($userCategories as $category) {
                $keywords = [];
                foreach ($category->getKeywords() as $keyword) {
                    $keywords[] = $keyword->getKeyword();
                }
                if (!empty($keywords)) {
                    $hasAnyMappings = true;
                    $message .= sprintf("📍 %s:\n%s\n\n", $category->getName(), implode(', ', $keywords));
                }
            }
        }

        if (!$hasAnyMappings) {
            $message .= "Сопоставлений пока нет. Чтобы добавить сопоставление, используйте команду:\n/map слово = категория\n\nНапример:\n/map еда = Питание";
        }

        $this->telegramApi->sendMessage([
            'chat_id' => $chatId,
            'text' => $message,
            'parse_mode' => 'HTML',
        ]);
    }
}
