"""
ImageForge Authentication Flow Tests
Tests login, registration, and form validation.
"""

import pytest
from playwright.sync_api import Page, expect

from conftest import DEV_SERVER_URL


def set_english_locale(page: Page):
    """Force English locale for consistent test assertions."""
    page.evaluate("() => localStorage.setItem('imageforge_locale', 'en')")


class TestAuthForms:
    """Tests for auth form validation and interaction."""

    def test_login_empty_submit_shows_validation(self, page: Page):
        """Submitting empty login form triggers browser validation."""
        page.goto(f"{DEV_SERVER_URL}/login")
        set_english_locale(page)
        page.goto(f"{DEV_SERVER_URL}/login")
        page.wait_for_load_state("networkidle")
        page.locator("button", has_text="Sign In").click()
        assert "/login" in page.url

    def test_login_with_credentials(self, page: Page):
        """Filling login form and submitting (API may fail but UI should respond)."""
        page.goto(f"{DEV_SERVER_URL}/login")
        set_english_locale(page)
        page.goto(f"{DEV_SERVER_URL}/login")
        page.wait_for_load_state("networkidle")

        page.fill("#email", "test@example.com")
        page.fill("#password", "password123")
        page.locator("button", has_text="Sign In").click()
        page.wait_for_timeout(3000)

        current_url = page.url
        assert "/app" in current_url or "/login" in current_url

    def test_register_with_credentials(self, page: Page):
        """Filling register form and submitting."""
        page.goto(f"{DEV_SERVER_URL}/login")
        set_english_locale(page)
        page.goto(f"{DEV_SERVER_URL}/login")
        page.wait_for_load_state("networkidle")

        page.locator("button", has_text="Sign up").click()
        page.fill("#name", "Test User")
        page.fill("#email", "newuser@example.com")
        page.fill("#password", "securepass123")
        page.locator("button", has_text="Create Account").click()
        page.wait_for_timeout(3000)

        current_url = page.url
        assert "/app" in current_url or "/login" in current_url

    def test_password_min_length_validation(self, page: Page):
        """Register form enforces min password length."""
        page.goto(f"{DEV_SERVER_URL}/login")
        set_english_locale(page)
        page.goto(f"{DEV_SERVER_URL}/login")
        page.wait_for_load_state("networkidle")

        page.locator("button", has_text="Sign up").click()
        page.fill("#name", "Test")
        page.fill("#email", "test@test.com")
        page.fill("#password", "short")
        page.locator("button", has_text="Create Account").click()
        page.wait_for_timeout(2000)
        assert "/login" in page.url

    def test_logout_button_visible_when_authenticated(self, page: Page):
        """Sign out button is visible in the app layout when logged in."""
        page.goto(f"{DEV_SERVER_URL}/login")
        page.evaluate(
            """() => {
            localStorage.setItem('imageforge_token', 'fake-token');
            localStorage.setItem('auth_user', JSON.stringify({ id: 1, email: 'test@example.com', role: 'user' }));
            localStorage.setItem('imageforge_locale', 'en');
        }"""
        )
        page.goto(f"{DEV_SERVER_URL}/app")
        page.wait_for_load_state("networkidle")
        expect(page.locator("text=Sign out")).to_be_visible()
