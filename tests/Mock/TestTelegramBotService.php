<?php

namespace App\Tests\Mock;

use App\Repository\UserRepository;
use App\Service\CommandRegistry;
use App\Service\MessageParserService;
use App\Service\StateHandler\StateHandlerRegistry;
use App\Service\TelegramBotService;
use App\Service\TransactionHandler;
use Psr\Log\LoggerInterface;

class TestTelegramBotService extends TelegramBotService
{
    private ResponseCollector $responseCollector;

    public function __construct(
        string $botToken,
        string $botUsername,
        UserRepository $userRepository,
        CommandRegistry $commandRegistry,
        StateHandlerRegistry $stateHandlerRegistry,
        TransactionHandler $transactionHandler,
        MessageParserService $messageParser,
        LoggerInterface $logger,
    ) {
        parent::__construct(
            $botToken,
            $botUsername,
            $userRepository,
            $commandRegistry,
            $stateHandlerRegistry,
            $transactionHandler,
            $messageParser,
            $logger
        );
        $this->responseCollector = ResponseCollector::getInstance();
    }

    /**
     * @param array<int, array<string, string>>|null $keyboard
     */
    protected function sendMessage(int $chatId, string $text, ?array $keyboard = null): void
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

        error_log('Telegram API Request: '.json_encode($data, JSON_UNESCAPED_UNICODE));
        $response = ['ok' => true, 'result' => null, 'description' => null, 'error_code' => null];
        error_log('Telegram API Response: '.json_encode($response, JSON_UNESCAPED_UNICODE));

        $this->responseCollector->addResponse($text);
    }
}
