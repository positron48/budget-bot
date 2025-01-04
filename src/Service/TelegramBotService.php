<?php

namespace App\Service;

use App\Entity\User;
use App\Repository\UserRepository;
use Longman\TelegramBot\Telegram;
use Longman\TelegramBot\Request;
use Longman\TelegramBot\Entities\Update;
use Longman\TelegramBot\Entities\Keyboard;
use Longman\TelegramBot\Exception\TelegramException;

class TelegramBotService
{
    private Telegram $telegram;
    private MessageParserService $messageParser;
    private CategoryService $categoryService;
    private GoogleSheetsService $sheetsService;
    private UserRepository $userRepository;

    public function __construct(
        string $botToken,
        string $botUsername,
        MessageParserService $messageParser,
        CategoryService $categoryService,
        GoogleSheetsService $sheetsService,
        UserRepository $userRepository
    ) {
        $this->telegram = new Telegram($botToken, $botUsername);
        $this->messageParser = $messageParser;
        $this->categoryService = $categoryService;
        $this->sheetsService = $sheetsService;
        $this->userRepository = $userRepository;
    }

    public function handleUpdate(array $updateData): void
    {
        try {
            $update = new Update($updateData);
            $message = $update->getMessage();
            
            if ($message === null) {
                return;
            }

            $chatId = $message->getChat()->getId();
            $text = $message->getText();
            
            if ($text === null) {
                return;
            }

            // Handle commands
            if ($text === '/start') {
                $this->handleStartCommand($chatId, $message);
                return;
            }

            if ($text === '/list') {
                $this->handleListCommand($chatId);
                return;
            }

            // Handle regular messages
            $this->handleMessage($chatId, $text, $message);
        } catch (TelegramException $e) {
            // Log error
        }
    }

    private function handleStartCommand(int $chatId, $message): void
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
            'text' => 'Привет! Я помогу вести учет доходов и расходов в Google Таблицах. ' .
                     'Отправляйте сообщения в формате: "[дата] [+]сумма описание"',
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

        $text = "Доступные таблицы:\n\n";
        foreach ($spreadsheets as $name => $id) {
            $text .= "{$name}: https://docs.google.com/spreadsheets/d/{$id}\n";
        }

        Request::sendMessage([
            'chat_id' => $chatId,
            'text' => $text,
            'disable_web_page_preview' => true,
        ]);
    }

    private function handleMessage(int $chatId, string $text, $message): void
    {
        $user = $this->userRepository->findByTelegramId($chatId);
        if (!$user || !$user->getCurrentSpreadsheetId()) {
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Пожалуйста, выберите таблицу с помощью команды /list',
            ]);
            return;
        }

        $parsedData = $this->messageParser->parseMessage($text);
        if (!$parsedData) {
            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Неверный формат сообщения. Используйте: "[дата] [+]сумма описание"',
            ]);
            return;
        }

        $category = $this->categoryService->detectCategory(
            $parsedData['description'],
            $parsedData['isIncome']
        );

        if ($category === null) {
            // Ask user to select category
            $categories = $this->categoryService->getCategories($parsedData['isIncome']);
            $keyboard = new Keyboard(...array_chunk($categories, 2));
            $keyboard->setResizeKeyboard(true)
                    ->setOneTimeKeyboard(true);

            Request::sendMessage([
                'chat_id' => $chatId,
                'text' => 'Выберите категорию:',
                'reply_markup' => $keyboard,
            ]);
            return;
        }

        $this->saveTransaction($user, $parsedData, $category);
    }

    private function saveTransaction(User $user, array $parsedData, string $category): void
    {
        try {
            if ($parsedData['isIncome']) {
                $this->sheetsService->addIncome(
                    $user->getCurrentSpreadsheetId(),
                    $parsedData['date']->format('d.m.Y'),
                    $parsedData['amount'],
                    $parsedData['description'],
                    $category
                );
            } else {
                $this->sheetsService->addExpense(
                    $user->getCurrentSpreadsheetId(),
                    $parsedData['date']->format('d.m.Y'),
                    $parsedData['amount'],
                    $parsedData['description'],
                    $category
                );
            }

            Request::sendMessage([
                'chat_id' => $user->getTelegramId(),
                'text' => 'Запись успешно добавлена!',
            ]);
        } catch (\Exception $e) {
            Request::sendMessage([
                'chat_id' => $user->getTelegramId(),
                'text' => 'Произошла ошибка при сохранении записи.',
            ]);
        }
    }
} 