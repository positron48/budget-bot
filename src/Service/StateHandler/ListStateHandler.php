<?php

namespace App\Service\StateHandler;

use App\Entity\User;
use App\Repository\UserRepository;
use App\Service\Google\GoogleApiClientInterface;
use App\Service\TelegramApiServiceInterface;
use Psr\Log\LoggerInterface;

class ListStateHandler implements StateHandlerInterface
{
    public function __construct(
        private readonly UserRepository $userRepository,
        private readonly TelegramApiServiceInterface $telegramApi,
        private readonly GoogleApiClientInterface $googleApiClient,
        private readonly LoggerInterface $logger,
    ) {
    }

    public function supports(string $state): bool
    {
        return 'WAITING_LIST_ACTION' === $state;
    }

    public function handle(int $chatId, User $user, string $message): bool
    {
        if (!in_array($message, ['Расходы', 'Доходы'])) {
            return false;
        }

        $tempData = $user->getTempData();
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
                'text' => sprintf('Нет %s за %s %d', 'Расходы' === $message ? 'расходов' : 'доходов', $this->getMonthName($month), $year),
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
                'text' => sprintf('Нет %s за %s %d', 'Расходы' === $message ? 'расходов' : 'доходов', $this->getMonthName($month), $year),
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

        // Format message
        $text = sprintf("%s за %s %d:\n\n", $message, $this->getMonthName($month), $year);
        $total = 0;

        foreach ($transactions as $t) {
            $text .= sprintf(
                "%s | %s руб. | [%s] %s\n",
                $t['date'],
                number_format((float) $t['amount'], 2, '.', ' '),
                $t['category'],
                $t['description']
            );
            $total += (float) $t['amount'];
        }

        $text .= sprintf("\nИтого: %.2f руб.", $total);

        $this->telegramApi->sendMessage([
            'chat_id' => $chatId,
            'text' => $text,
            'parse_mode' => 'HTML',
        ]);

        $user->setState('');
        $user->setTempData([]);
        $this->userRepository->save($user, true);

        return true;
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
}
