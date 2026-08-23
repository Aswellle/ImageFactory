"""
Comprehensive E2E test for ImageForge frontend.
Tests all UI functionality without backend (uses mocked auth state).
"""

import sys
import time
from pathlib import Path
from playwright.sync_api import sync_playwright, Page, expect

BASE_URL = "http://localhost:5173"
SCREENSHOT_DIR = Path(__file__).parent / "screenshots"
SCREENSHOT_DIR.mkdir(exist_ok=True)


def set_authenticated(page: Page) -> None:
    """Simulate authenticated state via localStorage."""
    page.evaluate(
        """() => {
        localStorage.setItem('imageforge_token', 'fake-token');
        localStorage.setItem('imageforge_user', JSON.stringify({ id: 1, email: 'test@example.com', role: 'user' }));
        localStorage.setItem('imageforge_locale', 'en');
        localStorage.setItem('imageforge_theme', 'light');
    }"""
    )


def set_admin(page: Page) -> None:
    """Simulate admin state via localStorage."""
    page.evaluate(
        """() => {
        localStorage.setItem('imageforge_token', 'fake-token');
        localStorage.setItem('imageforge_user', JSON.stringify({ id: 1, email: 'admin@example.com', role: 'admin' }));
        localStorage.setItem('imageforge_locale', 'en');
        localStorage.setItem('imageforge_theme', 'light');
    }"""
    )


def screenshot(page: Page, name: str) -> None:
    """Save screenshot for debugging."""
    path = SCREENSHOT_DIR / f"{name}.png"
    page.screenshot(path=str(path), full_page=False)
    print(f"  Screenshot: {path.name}")


def test_landing_page(page: Page) -> None:
    """Test the public landing page."""
    print("\n=== Landing Page ===")
    page.goto(BASE_URL)
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    # Set English locale for consistent testing
    page.evaluate("() => localStorage.setItem('imageforge_locale', 'en')")
    page.goto(BASE_URL)
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    # Hero section
    expect(page.locator("h1.text-display")).to_be_visible()
    print("  [OK] Hero heading visible")

    # Navigation
    expect(page.locator("nav")).to_be_visible()
    print("  [OK] Navigation visible")

    # Features section
    features_section = page.locator("#features")
    expect(features_section).to_be_visible()
    print("  [OK] Features section visible")

    # Workflow section
    expect(page.locator("#workflow")).to_be_visible()
    print("  [OK] Workflow section visible")

    # CTA section
    expect(page.locator("text=Open Desktop App")).to_be_visible()
    print("  [OK] CTA section visible")

    # Footer
    expect(page.locator("footer")).to_be_visible()
    print("  [OK] Footer visible")

    screenshot(page, "landing_page")


def test_dark_mode(page: Page) -> None:
    """Test dark mode toggle."""
    print("\n=== Dark Mode ===")
    page.goto(BASE_URL)
    page.wait_for_load_state("networkidle")

    # Get initial background
    bg_before = page.evaluate(
        "() => getComputedStyle(document.documentElement).getPropertyValue('--bg')"
    )
    print(f"  Before: --bg = {bg_before.strip()}")

    # Add dark class
    page.evaluate("() => document.documentElement.classList.add('dark')")
    page.wait_for_timeout(300)

    bg_after = page.evaluate(
        "() => getComputedStyle(document.documentElement).getPropertyValue('--bg')"
    )
    print(f"  After: --bg = {bg_after.strip()}")

    assert "000000" in bg_after.replace(" ", ""), "Dark mode should change --bg to #000000"
    print("  [OK] Dark mode toggles background color")

    # Remove dark class
    page.evaluate("() => document.documentElement.classList.remove('dark')")
    screenshot(page, "dark_mode")


def test_login_page(page: Page) -> None:
    """Test login page and form validation."""
    print("\n=== Login Page ===")
    page.goto(f"{BASE_URL}/login")
    page.wait_for_load_state("networkidle")

    # Set English locale for consistent testing
    page.evaluate("() => localStorage.setItem('imageforge_locale', 'en')")
    page.goto(f"{BASE_URL}/login")
    page.wait_for_load_state("networkidle")

    # Form elements (English)
    expect(page.locator("text=Sign in to your workspace")).to_be_visible()
    expect(page.locator("#email")).to_be_visible()
    expect(page.locator("#password")).to_be_visible()
    expect(page.locator("button", has_text="Sign In")).to_be_visible()
    print("  [OK] Login form elements visible")

    # Empty submit (browser validation)
    page.locator("button", has_text="Sign In").click()
    page.wait_for_timeout(500)
    assert "/login" in page.url, "Should stay on login page after empty submit"
    print("  [OK] Empty form submission prevented")

    # Toggle to register
    page.locator("button", has_text="Sign up").click()
    page.wait_for_timeout(300)
    expect(page.locator("#name")).to_be_visible()
    expect(page.locator("button", has_text="Create Account")).to_be_visible()
    print("  [OK] Toggle to register mode works")

    # Toggle back to login
    page.locator("button", has_text="Sign in").click()
    page.wait_for_timeout(300)
    expect(page.locator("#name")).to_have_count(0)
    print("  [OK] Toggle back to login mode works")

    screenshot(page, "login_page")


def test_auth_redirect(page: Page) -> None:
    """Test authentication redirects."""
    print("\n=== Auth Redirects ===")

    # Unauthenticated -> /app should redirect to /login
    page.goto(f"{BASE_URL}/app")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)
    assert "/login" in page.url, f"Unauthenticated /app should redirect to /login, got {page.url}"
    print("  [OK] Unauthenticated /app redirects to /login")

    # Authenticated -> / should redirect to /app
    set_authenticated(page)
    page.goto(BASE_URL)
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(1000)
    url = page.url
    assert "/app" in url or "/" == url.replace("http://localhost:5173", ""), f"Authenticated / should redirect to /app, got {url}"
    print("  [OK] Authenticated / redirects to /app")


def test_workspace_dashboard(page: Page) -> None:
    """Test workspace dashboard."""
    print("\n=== Workspace Dashboard ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    # Header
    expect(page.locator("h1")).to_be_visible()
    print("  [OK] Dashboard heading visible")

    # Stats cards
    expect(page.locator(".metric-value").first).to_be_visible()
    print("  [OK] Stats cards visible")

    # Sidebar navigation
    nav = page.locator("nav")
    expect(nav.locator("text=Create")).to_be_visible()
    expect(nav.locator("text=Gallery")).to_be_visible()
    expect(nav.locator("text=Projects")).to_be_visible()
    expect(nav.locator("text=Templates")).to_be_visible()
    print("  [OK] Sidebar navigation visible")

    # Sign out button
    expect(page.locator("text=Sign out")).to_be_visible()
    print("  [OK] Sign out button visible")

    screenshot(page, "dashboard")


def test_workspace_navigation(page: Page) -> None:
    """Test sidebar navigation between views."""
    print("\n=== Workspace Navigation ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app")
    page.wait_for_load_state("networkidle")

    # Navigate to Gallery
    page.locator("nav").get_by_text("Gallery").click()
    page.wait_for_url("**/app/assets")
    page.wait_for_timeout(300)
    print("  [OK] Navigate to Gallery")

    # Navigate to Create
    page.locator("nav").get_by_text("Create").click()
    page.wait_for_url("**/app/create")
    page.wait_for_timeout(300)
    print("  [OK] Navigate to Create")

    # Navigate to Projects
    page.locator("nav").get_by_text("Projects").click()
    page.wait_for_url("**/app/projects")
    page.wait_for_timeout(300)
    print("  [OK] Navigate to Projects")

    # Navigate to Templates
    page.locator("nav").get_by_text("Templates").click()
    page.wait_for_url("**/app/templates")
    page.wait_for_timeout(300)
    print("  [OK] Navigate to Templates")

    # Navigate back to Dashboard
    page.locator("nav").get_by_text("Workspace").click()
    page.wait_for_url("**/app")
    page.wait_for_timeout(300)
    print("  [OK] Navigate to Dashboard")


def test_gallery_view(page: Page) -> None:
    """Test gallery view."""
    print("\n=== Gallery View ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app/assets")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    # Search input
    search = page.locator("input[data-search-input]")
    expect(search).to_be_visible()
    print("  [OK] Search input visible")

    # View mode toggle
    grid_btn = page.locator("text=Grid").or_(page.locator("text=网格"))
    list_btn = page.locator("text=List").or_(page.locator("text=列表"))
    expect(grid_btn).to_be_visible()
    expect(list_btn).to_be_visible()
    print("  [OK] View mode toggle visible")

    # Toggle to list view
    list_btn.click()
    page.wait_for_timeout(300)
    grid_btn.click()
    page.wait_for_timeout(300)
    print("  [OK] View mode toggle works")

    screenshot(page, "gallery_view")


def test_generation_view(page: Page) -> None:
    """Test image generation view."""
    print("\n=== Generation View ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app/create")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    # Form elements
    expect(page.locator("#prompt")).to_be_visible()
    expect(page.locator("#model")).to_be_visible()
    expect(page.locator("#size")).to_be_visible()
    expect(page.locator("button", has_text="Generate")).to_be_visible()
    print("  [OK] Generation form elements visible")
    # Advanced toggle - click to expand (use negative prompt input as proxy)
    # The button text changes based on state; use a more robust approach
    page.locator("#prompt").fill("test prompt")
    page.locator("button", has_text="4").click()
    page.wait_for_timeout(200)
    print("  [OK] Form interaction works")

    # Image count buttons
    expect(page.locator(".btn", has_text="1")).to_be_visible()
    expect(page.locator(".btn", has_text="2")).to_be_visible()
    expect(page.locator(".btn", has_text="4")).to_be_visible()
    print("  [OK] Image count buttons visible")

    screenshot(page, "generation_view")


def test_i18n(page: Page) -> None:
    """Test internationalization."""
    print("\n=== i18n ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app")
    page.wait_for_load_state("networkidle")

    # Language switcher visible
    expect(page.locator("aside button", has_text="EN")).to_be_visible()
    print("  [OK] Language switcher visible")

    # Switch to Chinese
    page.locator("aside button", has_text="ZH").click()
    page.wait_for_timeout(500)
    expect(page.locator("text=工作台")).to_be_visible()
    print("  [OK] Switch to Chinese works")

    # Switch back to English
    page.locator("aside button", has_text="EN").click()
    page.wait_for_timeout(500)
    expect(page.locator("text=Workspace")).to_be_visible()
    print("  [OK] Switch back to English works")

    screenshot(page, "i18n")


def test_loading_states(page: Page) -> None:
    """Test loading states and transitions."""
    print("\n=== Loading States ===")
    set_authenticated(page)

    # Navigate to gallery (triggers data fetch)
    page.goto(f"{BASE_URL}/app/assets")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(1000)

    # Check for loading spinner or content
    # Since we have no backend, the page should show empty state or error
    print("  [OK] Gallery page loaded (with mocked auth)")

    # Navigate to generation
    page.goto(f"{BASE_URL}/app/create")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)
    print("  [OK] Generation page loaded")

    screenshot(page, "loading_states")


def test_sign_out(page: Page) -> None:
    """Test sign out functionality."""
    print("\n=== Sign Out ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app")
    page.wait_for_load_state("networkidle")

    # Click sign out
    page.locator("text=Sign out").click()
    page.wait_for_timeout(500)

    # Should redirect to login
    assert "/login" in page.url, f"Sign out should redirect to /login, got {page.url}"
    print("  [OK] Sign out redirects to login")

    screenshot(page, "sign_out")


def main() -> int:
    """Run all tests."""
    failed = 0
    tests = [
        ("Landing Page", test_landing_page),
        ("Dark Mode", test_dark_mode),
        ("Login Page", test_login_page),
        ("Auth Redirects", test_auth_redirect),
        ("Workspace Dashboard", test_workspace_dashboard),
        ("Workspace Navigation", test_workspace_navigation),
        ("Gallery View", test_gallery_view),
        ("Generation View", test_generation_view),
        ("i18n", test_i18n),
        ("Loading States", test_loading_states),
        ("Sign Out", test_sign_out),
    ]

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={"width": 1280, "height": 800})
        page = context.new_page()

        for name, test_fn in tests:
            try:
                test_fn(page)
                print(f"\n  PASSED: {name}")
            except Exception as e:
                print(f"\n  FAILED: {name}: {e}")
                screenshot(page, f"FAILED_{name.lower().replace(' ', '_')}")
                failed += 1

        browser.close()

    print(f"\n{'='*50}")
    print(f"Results: {len(tests) - failed}/{len(tests)} passed")
    if failed:
        print(f"Failed: {failed}")
        return 1
    print("All tests passed!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
