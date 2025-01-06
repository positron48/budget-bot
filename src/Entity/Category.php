<?php

namespace App\Entity;

use App\Repository\CategoryRepository;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity(repositoryClass: CategoryRepository::class)]
#[ORM\UniqueConstraint(name: 'UNIQ_64C19C15E237E06979B1AD6', columns: ['name', 'type'])]
class Category
{
    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column]
    private ?int $id = null;

    #[ORM\Column(length: 255)]
    private ?string $name = null;

    #[ORM\Column(length: 10)]
    private ?string $type = null;

    #[ORM\Column]
    private bool $isDefault = true;

    /** @var Collection<int, CategoryKeyword> */
    #[ORM\OneToMany(mappedBy: 'category', targetEntity: CategoryKeyword::class, cascade: ['persist', 'remove'])]
    private Collection $keywords;

    public function __construct()
    {
        $this->keywords = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
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

    public function isDefault(): bool
    {
        return $this->isDefault;
    }

    public function setIsDefault(bool $isDefault): static
    {
        $this->isDefault = $isDefault;

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
            $keyword->setCategory($this);
        }

        return $this;
    }

    public function removeKeyword(CategoryKeyword $keyword): static
    {
        if ($this->keywords->removeElement($keyword)) {
            if ($keyword->getCategory() === $this) {
                $keyword->setCategory(null);
            }
        }

        return $this;
    }
}
