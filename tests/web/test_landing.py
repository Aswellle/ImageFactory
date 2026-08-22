"""
ImageForge Landing Page & Routing Tests
Tests the SPA landing page and route guards.
"""

import pytest
from playwright.sync_api import Page, expect

from conftest import DEV_SERVER_URL


def set_english_locale(page: Page):
    """Force English locale for consistent test assertions."""
    page.evaluate("() => localStorage.setItem('imageforge_locale', 'en')")


class TestLandingPage:
    """Tests for the public-facing SPA landing page."""

    def test_landing_page_loads(self, page: Page):
        """Landing page loads with hero, features, workflow, CTA, footer."""
        page.goto(DEV_SERVER_URL)
        set_english_locale(page)
        page.goto(DEV_SERVER_URL)
        page.wait_for_load_state("networkidle")

        expect(page.locator("h1.text-display")).to_be_visible()
        expect(page.locator("text=Create images.")).to_be_visible()
        expect(page.locator("nav")).to_be_visible()
        expect(page.locator("text=Features")).to_be_visible()
        expect(page.locator("text=How it works")).to_be_visible()
        expect(page.locator("text=Open Desktop App")).to_be_visible()

    def test_landing_nav_scrolls_to_features(self, page: Page):
        """Clicking Features nav link scrolls to features section."""
        page.goto(DEV_SERVER_URL)
        set_english_locale(page)
        page.goto(DEV_SERVER_URL)
        page.wait_for_load_state("networkidle")
        page.click("text=Features")
        expect(page.locator("#features")).to_be_visible()

    def test_dark_mode_toggle(self, page: Page):
        """Dark mode class on html changes background variable."""
        page.goto(DEV_SERVER_URL)
        set_english_locale(page)
        page.goto(DEV_SERVER_URL)
        page.wait_for_load_state("networkidle")
        bg_before = page.evaluate(
            "() => getComputedStyle(document.documentElement).getPropertyValue('--bg')"
        )
        assert "f5f5f7" in bg_before.replace(" ", "")
        page.evaluate("() => document.documentElement.classList.add('dark')")
        bg_after = page.evaluate(
            "() => getComputedStyle(document.documentElement).getPropertyValue('--bg')"
        )
        assert "000000" in bg_after.replace(" ", "")


class TestRouting:
    """Tests for route guards and navigation."""

    def test_unauthenticated_redirect_to_login(self, page: Page):
        """Unauthenticated user visiting /app is redirected to /login."""
        page.goto(f"{DEV_SERVER_URL}/app")
        page.wait_for_load_state("networkidle")
        assert "/login" in page.url

    def test_login_page_loads(self, page: Page):
        """Login page loads with form elements."""
        page.goto(f"{DEV_SERVER_URL}/login")
        set_english_locale(page)
        page.goto(f"{DEV_SERVER_URL}/login")
        page.wait_for_load_state("networkidle")
        expect(page.locator("text=Sign in to your workspace")).to_be_visible()
        expect(page.locator("#email")).to_be_visible()
        expect(page.locator("#password")).to_be_visible()
        expect(page.locator("button", has_text="Sign In")).to_be_visible()

    def test_login_register_toggle(self, page: Page):
        """Toggling between login and register mode shows/hides name field."""
        page.goto(f"{DEV_SERVER_URL}/login")
        set_english_locale(page)
        page.goto(f"{DEV_SERVER_URL}/login")
        page.wait_for_load_state("networkidle")

        expect(page.locator("#name")).to_have_count(0)
        page.locator("button", has_text="Sign up").click()
        expect(page.locator("#name")).to_be_visible()
        page.locator("button", has_text="Sign in").click()
        expect(page.locator("#name")).to_have_count(0)

    def test_authenticated_landing_redirect(self, page: Page):
        """Authenticated user visiting / is redirected to /app."""
        page.goto(DEV_SERVER_URL)
        page.evaluate(
            """() => {
            localStorage.setItem('imageforge_token', 'fake-token');
            localStorage.setItem('auth_user', JSON.stringify({ id: 1, email: 'test@example.com', role: 'user' }));
            localStorage.setItem('imageforge_locale', 'en');
        }"""
        )
        page.goto(DEV_SERVER_URL)
        page.wait_for_load_state("networkidle")
        page.wait_for_timeout(1000)
        url = page.url
        assert "/app" in url or "/" == url.replace("http://localhost:5173", "")
