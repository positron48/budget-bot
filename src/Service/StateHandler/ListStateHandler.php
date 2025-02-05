<?php

namespace App\Service\StateHandler;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\Google\GoogleApiClientInterface;
use App\Service\TelegramApiServiceInterface;
use App\Utility\MonthUtility;
use Psr\Log\LoggerInterface;

class ListStateHandler implements StateHandlerInterface
{
    private const SUPPORTED_STATES = [
        'WAITING_LIST_ACTION',
        'WAITING_LIST_PAGE',
    ];

    private readonly int $transactionsPerPage;

    public function __construct(
        private readonly UserRepository $userRepository,
        private readonly TelegramApiServiceInterface $telegramApi,
        private readonly GoogleApiClientInterface $googleApiClient,
        private readonly LoggerInterface $logger,
        int $transactionsPerPage,
    ) {
        $this->transactionsPerPage = $transactionsPerPage;
    }

    public function supports(string $state): bool
    {
        return in_array($state, self::SUPPORTED_STATES, true);
    }

    public function handle(int $chatId, User $user, string $message): bool
    {
        $state = $user->getState();
        $tempData = $user->getTempData();

        $this->logger->debug('ListStateHandler handling message', [
            'chat_id' => $chatId,
            'state' => $state,
            'message' => $message,
            'temp_data' => $tempData,
        ]);

        if ('WAITING_LIST_PAGE' === $state) {
            return $this->handlePageNavigation($chatId, $user, $message);
        }

        if ('WAITING_LIST_ACTION' !== $state) {
            $this->logger->debug('ListStateHandler: state is not WAITING_LIST_ACTION', [
                'state' => $state,
            ]);

            return false;
        }

        if (!in_array($message, ['Расходы', 'Доходы'])) {
            $this->logger->debug('ListStateHandler: message is not Расходы or Доходы', [
                'message' => $message,
            ]);

            return false;
        }

        if (!isset($tempData['list_month'], $tempData['list_year'], $tempData['spreadsheet_id'])) {
            $this->logger->error('Missing required temp data', [
                'chat_id' => $chatId,
                'temp_data' => $tempData,
            ]);
            throw new \RuntimeException('Missing required temp data');
        }

        $month = $tempData['list_month'];
        $year = $tempData['list_year'];
        $spreadsheetId = $tempData['spreadsheet_id'];

        // Get transactions from the spreadsheet
        $range = 'Расходы' === $message ? 'Транзакции!B5:E' : 'Транзакции!G5:J';
        $values = $this->googleApiClient->getValues($spreadsheetId, $range);

        if (!$values) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf('Нет %s за %s %d', 'Расходы' === $message ? 'расходов' : 'доходов', MonthUtility::getMonthName($month), $year),
                'parse_mode' => 'HTML',
            ]);

            $user->setState('');
            $user->setTempData([]);
            $this->userRepository->save($user, true);

            return true;
        }

        // Filter transactions by month and year
        $transactions = [];
        foreach ($values as $row) {
            if (count($row) < 4) {
                continue;
            }

            [$date, $amount, $description, $category] = array_map('strval', $row);

            // Parse date
            try {
                $transactionDate = \DateTime::createFromFormat('d.m.Y', $date);
                if (!$transactionDate) {
                    continue;
                }

                $transactionMonth = (int) $transactionDate->format('m');
                $transactionYear = (int) $transactionDate->format('Y');

                if ($transactionMonth === $month && $transactionYear === $year) {
                    $transactions[] = [
                        'date' => $date,
                        'amount' => $amount,
                        'description' => $description,
                        'category' => $category,
                    ];
                }
            } catch (\Exception $e) {
                $this->logger->warning('Failed to parse transaction date', [
                    'chat_id' => $chatId,
                    'date' => $date,
                    'error' => $e->getMessage(),
                ]);
                continue;
            }
        }

        if (empty($transactions)) {
            $this->telegramApi->sendMessage([
                'chat_id' => $chatId,
                'text' => sprintf('Нет %s за %s %d', 'Расходы' === $message ? 'расходов' : 'доходов', MonthUtility::getMonthName($month), $year),
                'parse_mode' => 'HTML',
            ]);

            $user->setState('');
            $user->setTempData([]);
            $this->userRepository->save($user, true);

            return true;
        }

        // Sort transactions by date in reverse order
        usort($transactions, function (array $a, array $b): int {
            return strcmp((string) $b['date'], (string) $a['date']);
        });

        // Save transactions in temp data for pagination
        $tempData['transactions'] = $transactions;
        $tempData['current_page'] = 1;
        $tempData['type'] = $message;
        $user->setTempData($tempData);
        $user->setState('WAITING_LIST_PAGE');
        $this->userRepository->save($user, true);

        return $this->showTransactionsPage($chatId, $user);
    }

    private function handlePageNavigation(int $chatId, User $user, string $message): bool
    {
        $tempData = $user->getTempData();
        if (!isset($tempData['transactions'], $tempData['current_page'], $tempData['type'])) {
            $user->setState('');
            $user->setTempData([]);
            $this->userRepository->save($user, true);

            return false;
        }

        $totalPages = ceil(count($tempData['transactions']) / $this->transactionsPerPage);

        switch ($message) {
            case '⬅️ Назад':
                if ($tempData['current_page'] > 1) {
                    --$tempData['current_page'];
                    $user->setTempData($tempData);
                    $this->userRepository->save($user, true);
                }
                break;
            case '➡️ Вперед':
                if ($tempData['current_page'] < $totalPages) {
                    ++$tempData['current_page'];
                    $user->setTempData($tempData);
                    $this->userRepository->save($user, true);
                }
                break;
            case '❌ Закрыть':
                $user->setState('');
                $user->setTempData([]);
                $this->userRepository->save($user, true);
                $this->telegramApi->sendMessage([
                    'chat_id' => $chatId,
                    'text' => 'Просмотр транзакций завершен',
                    'parse_mode' => 'HTML',
                    'reply_markup' => json_encode([
                        'remove_keyboard' => true,
                    ]),
                ]);

                return true;
            default:
                return false;
        }

        return $this->showTransactionsPage($chatId, $user);
    }

    private function showTransactionsPage(int $chatId, User $user): bool
    {
        $tempData = $user->getTempData();
        $transactions = $tempData['transactions'];
        $currentPage = $tempData['current_page'];
        $totalPages = ceil(count($transactions) / $this->transactionsPerPage);

        $start = ($currentPage - 1) * $this->transactionsPerPage;
        $pageTransactions = array_slice($transactions, $start, $this->transactionsPerPage);

        // Format message
        $text = sprintf(
            "%s за %s %d (страница %d из %d):\n\n",
            $tempData['type'],
            MonthUtility::getMonthName($tempData['list_month']),
            $tempData['list_year'],
            $currentPage,
            $totalPages
        );

        $total = 0;
        foreach ($pageTransactions as $t) {
            $text .= sprintf(
                "%s | %s руб. | [%s] %s\n",
                $t['date'],
                number_format((float) $t['amount'], 2, '.', ' '),
                $t['category'],
                $t['description']
            );
            $total += (float) $t['amount'];
        }

        // Add total for all transactions, not just current page
        $totalAll = array_sum(array_map(fn ($t) => (float) $t['amount'], $transactions));
        $text .= sprintf("\nИтого за страницу: %.2f руб.", $total);
        $text .= sprintf("\nОбщий итог: %.2f руб.", $totalAll);

        // Prepare navigation buttons
        $keyboard = [];
        $row = [];

        if ($currentPage > 1) {
            $row[] = ['text' => '⬅️ Назад'];
        }
        if ($currentPage < $totalPages) {
            $row[] = ['text' => '➡️ Вперед'];
        }
        if (!empty($row)) {
            $keyboard[] = $row;
        }
        $keyboard[] = [['text' => '❌ Закрыть']];

        $this->telegramApi->sendMessage([
            'chat_id' => $chatId,
            'text' => $text,
            'parse_mode' => 'HTML',
            'reply_markup' => json_encode([
                'keyboard' => $keyboard,
                'resize_keyboard' => true,
                'one_time_keyboard' => false,
            ]),
        ]);

        return true;
    }
}
