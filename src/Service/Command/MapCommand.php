<?php

namespace App\Service\Command;

use App\Entity\User;
use App\Repository\CategoryRepository;
use App\Repository\UserCategoryRepository;
use App\Repository\UserRepository;
use App\Service\CategoryService;
use Psr\Log\LoggerInterface;

class MapCommand extends AbstractCommand
{
    public function __construct(
        private readonly CategoryService $categoryService,
        private readonly CategoryRepository $categoryRepository,
        private readonly UserCategoryRepository $userCategoryRepository,
        UserRepository $userRepository,
        LoggerInterface $logger,
    ) {
        parent::__construct($userRepository, $logger);
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
            $this->sendMessage($chatId, 'Пожалуйста, начните с команды /start');

            return;
        }

        $parts = preg_split('/\s+/', trim($message), 2);
        if (!is_array($parts) || count($parts) < 2) {
            $this->sendMessage($chatId, 'Пожалуйста, укажите описание расхода после команды /map. Например: /map еда');

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
            $this->sendMessage($chatId, sprintf('Для описания "%s" категория не найдена', $description));

            return;
        }

        $this->sendMessage(
            $chatId,
            sprintf('Описание "%s" соответствует категории "%s"', $description, $category)
        );
    }

    private function handleMapping(int $chatId, User $user, string $input): void
    {
        $parts = array_map('trim', explode('=', $input));
        if (2 !== count($parts)) {
            $this->sendMessage($chatId, 'Неверный формат. Используйте: слово = категория');

            return;
        }

        $keyword = mb_strtolower($parts[0]);
        $categoryName = $parts[1];

        // Get sorted list of categories
        $categories = $this->categoryService->getCategories(false, $user);
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

        $this->categoryService->addKeywordToCategory($keyword, $categoryName, 'expense', $user);
        $this->sendMessage(
            $chatId,
            sprintf('Добавлено сопоставление: "%s" → "%s"', $keyword, $categoryName)
        );
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

        $this->sendMessage($chatId, $message);
    }
}
