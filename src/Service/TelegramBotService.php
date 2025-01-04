<?php

namespace App\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use Longman\TelegramBot\Entities\Keyboard;
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

            // Handle commands
            if ('/start' === $text) {
                $this->handleStartCommand($chatId, $message);

                return;
            }

            if ('/list' === $text) {
                $this->handleListCommand($chatId);

                return;
            }

            if ('/categories' === $text) {
                $this->handleCategoriesCommand($chatId);

                return;
            }

            // Handle regular message
            $this->handleMessage($chatId, $message);
        } catch (TelegramException $e) {
            $this->logger->error('Error handling update: '.$e->getMessage(), [
                'exception' => $e,
                'update' => $updateData,
            ]);
        }
    }

    private function handleStartCommand(int $chatId, Message $message): void
    {
        $user = $this->userRepository->findByTelegramId($chatId);

        if (!$user) {
            $user = new User();
            $user->setTelegramId($chatId)
                ->setUsername($message->getFrom()->getUsername())
                ->setFirstName($message->getFrom()->getFirstName())
                ->setLastName($message->getFrom()->getLastName());

            $this->userRepository->save($user, true);
        }

        Request::sendMessage([
            'chat_id' => $chatId,
            'text' => 'Привет! Я помогу вести учет доходов и расходов в Google Таблицах. '.
                     'Отправляйте сообщения в формате: "[дата] [+]сумма описание"'.
                     "\n\nДоступные команды:\n".
                     "/list - список доступных таблиц\n".
                     '/categories - управление категориями',
        ]);
    }

    private function handleListCommand(int $chatId): void
    {
        $spreadsheets = $this->sheetsService->getSpreadsheetsList();

        if (empty($spreadsheets)) {
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Нет доступных таблиц.',
            ]);

            return;
        }

        $keyboard = new Keyboard(...array_chunk($spreadsheets, 2));
        $keyboard->setResizeKeyboard(true)
            ->setOneTimeKeyboard(true)
            ->setSelective(false);

        Request::sendMessage([
            'chat_id' => $chatId,
            'text' => 'Выберите таблицу:',
            'reply_markup' => $keyboard,
        ]);
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

    private function handleMessage(int $chatId, Message $message): void
    {
        $user = $this->userRepository->findByTelegramId($chatId);
        if (!$user || !$user->getCurrentSpreadsheetId()) {
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, выберите таблицу с помощью команды /list',
            ]);

            return;
        }

        $parsedData = $this->messageParser->parseMessage($message->getText() ?? '');
        if (!$parsedData) {
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Неверный формат сообщения. Используйте: "[дата] [+]сумма описание"',
            ]);

            return;
        }

        $type = $parsedData['isIncome'] ? 'income' : 'expense';
        $category = $this->categoryService->detectCategory(
            $parsedData['description'],
            $type,
            $user
        );

        if (null === $category) {
            // Ask user to select category
            $categories = $this->categoryService->getCategories($parsedData['isIncome'], $user);
            $keyboard = new Keyboard(...array_chunk($categories, 2));
            $keyboard->setResizeKeyboard(true)
                ->setOneTimeKeyboard(true)
                ->setSelective(false);

            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Выберите категорию:',
                'reply_markup' => $keyboard,
            ]);

            return;
        }

        $this->saveTransaction($chatId, $user->getCurrentSpreadsheetId(), $parsedData, $category);

        Request::sendMessage([
            'chat_id' => $chatId,
            'text' => 'Транзакция сохранена.',
        ]);
    }

    /**
     * @param array{date: \DateTime, amount: float, description: string, isIncome: bool} $parsedData
     */
    private function saveTransaction(int $chatId, string $spreadsheetId, array $parsedData, string $category): void
    {
        try {
            if ($parsedData['isIncome']) {
                $this->sheetsService->addIncome(
                    $spreadsheetId,
                    $parsedData['date']->format('d.m.Y'),
                    $parsedData['amount'],
                    $parsedData['description'],
                    $category
                );
            } else {
                $this->sheetsService->addExpense(
                    $spreadsheetId,
                    $parsedData['date']->format('d.m.Y'),
                    $parsedData['amount'],
                    $parsedData['description'],
                    $category
                );
            }
        } catch (\Exception $e) {
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Ошибка при сохранении транзакции.',
            ]);
        }
    }
}
