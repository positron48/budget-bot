<?php

namespace App\Command;

use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;
use Longman\TelegramBot\Telegram;
use Longman\TelegramBot\Exception\TelegramException;

#[AsCommand(
    name: 'app:set-webhook',
    description: 'Sets up the Telegram webhook URL'
)]
class SetWebhookCommand extends Command
{
    private string $botToken;
    private string $botUsername;

    public function __construct(string $botToken, string $botUsername)
    {
        parent::__construct();
        $this->botToken = $botToken;
        $this->botUsername = $botUsername;
    }

    protected function configure(): void
    {
        $this->addArgument('url', InputArgument::REQUIRED, 'The webhook URL');
    }

    protected function execute(InputInterface $input, OutputInterface $output): int
    {
        try {
            $telegram = new Telegram($this->botToken, $this->botUsername);
            $result = $telegram->setWebhook($input->getArgument('url'));

            if ($result->isOk()) {
                $output->writeln('<info>Webhook was set successfully!</info>');
                return Command::SUCCESS;
            }

            $output->writeln('<error>Failed to set webhook: ' . $result->getDescription() . '</error>');
            return Command::FAILURE;
        } catch (TelegramException $e) {
            $output->writeln('<error>Error: ' . $e->getMessage() . '</error>');
            return Command::FAILURE;
        }
    }
} 