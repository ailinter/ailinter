# Refactoring Strategies

AILINTER provides step-by-step refactoring guidance for 8+ common code smells.

> **This page is under construction.** For now, use the `get_refactoring_strategy` MCP tool or see the [README](https://github.com/ailinter/ailinter#the-refactoring-loop).

## Available Strategies

| Smell | Strategy |
|-------|----------|
| Deep Nesting | Guard Clauses + Extract Method |
| Brain Method | Extract Method + SRP |
| Bumpy Road | Extract Method + Flatten |
| Complex Conditional | Guard Clauses + Strategy Pattern |
| God Class | Extract Class + SRP |
| Long Parameter List | Parameter Object Pattern |
| Primitive Obsession | Type Wrapper + Value Object |
| Duplicated Code | Extract Method + Template Method |

## Usage

```bash
# Via CLI (direct)
ailinter check --refactor-strategy deep_nesting

# Via MCP (from AI assistant)
# → get_refactoring_strategy("deep_nesting")
```
