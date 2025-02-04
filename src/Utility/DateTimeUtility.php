<?php

namespace App\Utility;

class DateTimeUtility
{
    protected ?\DateTime $mockedDate = null;

    public function getCurrentDate(): \DateTime
    {
        return $this->mockedDate ?? new \DateTime();
    }

    public function setCurrentDate(\DateTime $date): void
    {
        $this->mockedDate = clone $date;
    }

    public function resetCurrentDate(): void
    {
        $this->mockedDate = null;
    }
} 