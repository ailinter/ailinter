# Parallel Inheritance — Merge Parallel Hierarchies

## What

Two or more class hierarchies that must stay in sync — every new subclass in one hierarchy requires a matching subclass in another. In Go, this appears as parallel interface or struct hierarchies that must be extended together.

## Why It's Harmful

- **Multiplication of types**: Adding one concept requires creating N types (one per hierarchy).
- **Synchronization burden**: Every developer must remember to update all parallel hierarchies.
- **AI confusion**: LLMs create a new type in one hierarchy but forget the corresponding type in others.

## Step-by-Step Refactoring

### 1. Identify Parallel Hierarchies

Look for hierarchies that always change together:

```go
// BEFORE: Two parallel hierarchies — every Animal needs an AnimalSound
type Animal interface {
    Speak() string
    Name() string
}

type Dog struct{}
func (d Dog) Speak() string { return "woof" }
func (d Dog) Name() string  { return "Dog" }

type Cat struct{}
func (c Cat) Speak() string { return "meow" }
func (c Cat) Name() string  { return "Cat" }

type AnimalSound interface {
    Sound() string
    AnimalName() string
}

type DogSound struct {
    animal *Dog
}
func (d *DogSound) Sound() string      { return d.animal.Speak() }
func (d *DogSound) AnimalName() string { return d.animal.Name() }

type CatSound struct {
    animal *Cat
}
func (c *CatSound) Sound() string      { return c.animal.Speak() }
func (c *CatSound) AnimalName() string { return c.animal.Name() }
```

### 2. Merge into a Single Hierarchy

```go
// AFTER: Single hierarchy — Sound is a method on Animal itself
type Animal interface {
    Speak() string
    Name() string
}

type Dog struct{}
func (d Dog) Speak() string { return "woof" }
func (d Dog) Name() string  { return "Dog" }

type Cat struct{}
func (c Cat) Speak() string { return "meow" }
func (c Cat) Name() string  { return "Cat" }

// Everything that needs sound just calls Animal.Speak()
// No separate sound hierarchy needed.
```

### 3. Use Composition over Parallel Hierarchies

If the hierarchies can't fully merge, use a single config struct instead:

```go
type Fish struct {
    sound string // "blub" — no need for FishSound interface
}
func (f Fish) Speak() string { return f.sound }
func (f Fish) Name() string  { return "Fish" }
```

## Verification

- [ ] No two type hierarchies share the same extension pattern
- [ ] New functionality requires adding types in one place, not N
- [ ] Re-run `analyze_code` — Parallel Inheritance smell should be resolved
