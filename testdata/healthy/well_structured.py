"""A well-structured module with clean code."""

from dataclasses import dataclass


@dataclass
class User:
    name: str
    email: str

    def is_valid(self) -> bool:
        return bool(self.name and self.email and "@" in self.email)


def create_greeting(user: User) -> str:
    if not user.is_valid():
        return "Hello, stranger!"
    return f"Hello, {user.name}!"


def calculate_discount(price: float, loyalty_years: int) -> float:
    if loyalty_years <= 0:
        return 0.0
    if loyalty_years > 10:
        return price * 0.2
    return price * (loyalty_years / 100.0)
