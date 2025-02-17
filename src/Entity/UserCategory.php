<?php

namespace App\Entity;

use App\Repository\UserCategoryRepository;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity(repositoryClass: UserCategoryRepository::class)]
#[ORM\UniqueConstraint(name: 'UNIQ_E9B0D9895E237E06979B1AD6A76ED395', columns: ['name', 'type', 'user_id'])]
class UserCategory
{
    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column]
    private ?int $id = null;

    #[ORM\ManyToOne(targetEntity: User::class)]
    #[ORM\JoinColumn(nullable: false)]
    private ?User $user = null;

    #[ORM\Column(length: 255)]
    private ?string $name = null;

    #[ORM\Column(length: 10)]
    private ?string $type = null;

    #[ORM\Column]
    private bool $isIncome = false;

    /** @var Collection<int, CategoryKeyword> */
    #[ORM\OneToMany(mappedBy: 'userCategory', targetEntity: CategoryKeyword::class, cascade: ['persist', 'remove'])]
    private Collection $keywords;

    public function __construct()
    {
        $this->keywords = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getUser(): ?User
    {
        return $this->user;
    }

    public function setUser(?User $user): static
    {
        $this->user = $user;

        return $this;
    }

    public function getName(): ?string
    {
        return $this->name;
    }

    public function setName(string $name): static
    {
        $this->name = $name;

        return $this;
    }

    public function getType(): ?string
    {
        return $this->type;
    }

    public function setType(string $type): static
    {
        $this->type = $type;

        return $this;
    }

    public function isIncome(): bool
    {
        return $this->isIncome;
    }

    public function setIsIncome(bool $isIncome): static
    {
        $this->isIncome = $isIncome;

        return $this;
    }

    /** @return Collection<int, CategoryKeyword> */
    public function getKeywords(): Collection
    {
        return $this->keywords;
    }

    public function addKeyword(CategoryKeyword $keyword): static
    {
        if (!$this->keywords->contains($keyword)) {
            $this->keywords->add($keyword);
            $keyword->setUserCategory($this);
        }

        return $this;
    }

    public function removeKeyword(CategoryKeyword $keyword): static
    {
        if ($this->keywords->removeElement($keyword)) {
            if ($keyword->getUserCategory() === $this) {
                $keyword->setUserCategory(null);
            }
        }

        return $this;
    }
}
