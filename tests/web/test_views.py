"""
ImageForge Tool Views Tests
Tests the main application views (dashboard, gallery, generation, etc.).
"""

import pytest
from playwright.sync_api import Page, expect

from conftest import DEV_SERVER_URL


def authenticate_page(page: Page):
    """Helper to simulate authenticated state with English locale."""
    page.evaluate(
        """() => {
        localStorage.setItem('imageforge_token', 'fake-token');
        localStorage.setItem('auth_user', JSON.stringify({ id: 1, email: 'test@example.com', role: 'user' }));
        localStorage.setItem('imageforge_locale', 'en');
    }"""
    )


class TestDashboard:
    """Tests for the dashboard view."""

    def test_dashboard_loads(self, page: Page):
        """Dashboard loads with hero, stats, and recent images section."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app")
        page.wait_for_load_state("networkidle")
        expect(page.locator("h1")).to_be_visible()

    def test_dashboard_shows_stats_cards(self, page: Page):
        """Dashboard shows stat cards (Total Images, Projects, Storage)."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app")
        page.wait_for_load_state("networkidle")
        expect(page.locator(".metric-value").first).to_be_visible()


class TestGalleryView:
    """Tests for the gallery/asset library view."""

    def test_gallery_loads(self, page: Page):
        """Gallery loads with header, search, and view mode toggle."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app/assets")
        page.wait_for_load_state("networkidle")
        # Search input placeholder may be en/zh depending on locale.
        expect(page.locator("input[placeholder*='Search']").or_(page.locator("input[placeholder*='搜索']"))).to_be_visible()

    def test_gallery_view_mode_toggle(self, page: Page):
        """Can toggle between grid and list view."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app/assets")
        page.wait_for_load_state("networkidle")

        # View mode buttons: Grid/List or 网格/列表
        grid_btn = page.locator("text=Grid").or_(page.locator("text=网格"))
        list_btn = page.locator("text=List").or_(page.locator("text=列表"))
        expect(grid_btn).to_be_visible()
        expect(list_btn).to_be_visible()

        list_btn.click()
        page.wait_for_timeout(300)
        grid_btn.click()
        page.wait_for_timeout(300)


class TestGenerationView:
    """Tests for the image generation view."""

    def test_generation_form_elements(self, page: Page):
        """Generation form has prompt, model, size, and submit."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app/create")
        page.wait_for_load_state("networkidle")

        expect(page.locator("#prompt")).to_be_visible()
        expect(page.locator("#model")).to_be_visible()
        expect(page.locator("#size")).to_be_visible()
        expect(page.locator("button", has_text="Generate")).to_be_visible()

    def test_generation_advanced_toggle(self, page: Page):
        """Advanced settings toggle shows/hides negative prompt."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app/create")
        page.wait_for_load_state("networkidle")

        expect(page.locator("#negative")).to_have_count(0)
        page.click("text=Advanced")
        expect(page.locator("#negative")).to_be_visible()

    def test_generation_image_count_buttons(self, page: Page):
        """Image count selector buttons work."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app/create")
        page.wait_for_load_state("networkidle")

        expect(page.locator(".btn", has_text="1")).to_be_visible()
        expect(page.locator(".btn", has_text="2")).to_be_visible()
        expect(page.locator(".btn", has_text="4")).to_be_visible()


class TestNavigation:
    """Tests for sidebar navigation."""

    def test_sidebar_nav_links(self, page: Page):
        """Sidebar has links to all main sections."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app")
        page.wait_for_load_state("networkidle")

        # Use nav scope to avoid matching headings with same text.
        nav = page.locator("nav")
        expect(nav.locator("text=Create")).to_be_visible()
        expect(nav.locator("text=Gallery")).to_be_visible()
        expect(nav.locator("text=Projects")).to_be_visible()
        expect(nav.locator("text=Templates")).to_be_visible()

    def test_navigate_to_gallery(self, page: Page):
        """Clicking Gallery nav navigates to /app/assets."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app")
        page.wait_for_load_state("networkidle")

        page.locator("nav").get_by_text("Gallery").click()
        page.wait_for_url("**/app/assets")

    def test_navigate_to_create(self, page: Page):
        """Clicking Create nav navigates to /app/create."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app")
        page.wait_for_load_state("networkidle")

        page.locator("nav").get_by_text("Create").click()
        page.wait_for_url("**/app/create")


class TestI18n:
    """Tests for internationalization."""

    def test_language_switcher_visible(self, page: Page):
        """Language switcher is visible in sidebar."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app")
        page.wait_for_load_state("networkidle")
        # Language switcher shows "EN" in the sidebar footer.
        expect(page.locator("aside button", has_text="EN")).to_be_visible()

    def test_language_switch_to_chinese(self, page: Page):
        """Switching to Chinese updates UI text."""
        page.goto(f"{DEV_SERVER_URL}/login")
        authenticate_page(page)
        page.goto(f"{DEV_SERVER_URL}/app")
        page.wait_for_load_state("networkidle")

        page.locator("aside button", has_text="ZH").click()
        page.wait_for_timeout(500)
        expect(page.locator("text=工作台")).to_be_visible()
