"""
Comprehensive UI/UX visual review for ImageForge.
Captures screenshots of all pages and tests interactive functionality.
"""

import sys
import time
from pathlib import Path
from playwright.sync_api import sync_playwright, Page, expect

BASE_URL = "http://127.0.0.1:5175"
SCREENSHOT_DIR = Path(__file__).parent / "screenshots" / "ux_review"
SCREENSHOT_DIR.mkdir(parents=True, exist_ok=True)
BASE_URL = "http://localhost:5173"
ALT_URL = "http://localhost:5173"


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


def screenshot(page: Page, name: str) -> None:
    """Save screenshot for review."""
    path = SCREENSHOT_DIR / f"{name}.png"
    page.screenshot(path=str(path), full_page=True)
    print(f"  Screenshot: {path.name}")


def test_landing_page(page: Page) -> None:
    """Test landing page UI/UX."""
    print("\n=== Landing Page ===")
    page.goto(BASE_URL)
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(800)

    # Hero section
    hero = page.locator("h1.text-display")
    expect(hero).to_be_visible()
    print("  [OK] Hero heading visible")

    # Navigation
    nav = page.locator("nav")
    expect(nav).to_be_visible()
    print("  [OK] Navigation visible")

    # Check font is Geist
    font = page.evaluate(
        "() => getComputedStyle(document.body).fontFamily"
    )
    print(f"  Font: {font}")
    assert "Geist" in font, f"Expected Geist font, got {font}"
    print("  [OK] Geist font loaded")

    # Check CSS variables
    bg = page.evaluate(
        "() => getComputedStyle(document.documentElement).getPropertyValue('--bg')"
    )
    text = page.evaluate(
        "() => getComputedStyle(document.documentElement).getPropertyValue('--text')"
    )
    accent = page.evaluate(
        "() => getComputedStyle(document.documentElement).getPropertyValue('--accent')"
    )
    print(f"  CSS vars: --bg={bg.strip()}, --text={text.strip()}, --accent={accent.strip()}")

    # Scroll through all sections
    for section in ["#features", "#workflow"]:
        el = page.locator(section)
        if el.count() > 0:
            el.scroll_into_view_if_needed()
            page.wait_for_timeout(300)
            print(f"  [OK] Section {section} visible")

    # Footer
    footer = page.locator("footer")
    expect(footer).to_be_visible()
    print("  [OK] Footer visible")

    screenshot(page, "01_landing_full")

    # Dark mode test
    page.evaluate("() => document.documentElement.classList.add('dark')")
    page.wait_for_timeout(500)
    bg_dark = page.evaluate(
        "() => getComputedStyle(document.documentElement).getPropertyValue('--bg')"
    )
    print(f"  Dark mode --bg: {bg_dark.strip()}")
    assert "000000" in bg_dark.replace(" ", ""), "Dark mode should set --bg to #000000"
    screenshot(page, "01_landing_dark")
    page.evaluate("() => document.documentElement.classList.remove('dark')")
    page.wait_for_timeout(300)


def test_login_page(page: Page) -> None:
    """Test login page UI/UX."""
    print("\n=== Login Page ===")
    # Set English locale before navigation
    page.goto(f"{BASE_URL}/login")
    page.wait_for_load_state("networkidle")
    page.evaluate("() => localStorage.setItem('imageforge_locale', 'en')")
    page.goto(f"{BASE_URL}/login")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    # Form elements
    expect(page.locator("text=Sign in to your workspace")).to_be_visible()
    expect(page.locator("#email")).to_be_visible()
    expect(page.locator("#password")).to_be_visible()
    print("  [OK] Login form elements visible")

    # Input styling
    input_bg = page.evaluate(
        "() => getComputedStyle(document.querySelector('.input')).backgroundColor"
    )
    print(f"  Input background: {input_bg}")

    # Button styling
    btn = page.locator("button", has_text="Sign In")
    btn_bg = page.evaluate(
        "el => getComputedStyle(el).backgroundColor",
        btn.element_handle()
    )
    print(f"  Primary button background: {btn_bg}")

    screenshot(page, "02_login")

    # Toggle to register
    page.locator("button", has_text="Sign up").click()
    page.wait_for_timeout(300)
    expect(page.locator("#name")).to_be_visible()
    screenshot(page, "02_register")


def test_workspace_dashboard(page: Page) -> None:
    """Test workspace dashboard UI/UX."""
    print("\n=== Workspace Dashboard ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(800)

    # Header
    expect(page.locator("h1")).to_be_visible()
    print("  [OK] Dashboard heading visible")

    # Stats cards
    stats = page.locator(".metric-value")
    count = stats.count()
    print(f"  Stats cards: {count}")
    assert count >= 3, f"Expected at least 3 stat cards, got {count}"

    # Sidebar navigation
    nav = page.locator("nav")
    for item in ["Workspace", "Create", "Gallery", "Projects", "Templates"]:
        expect(nav.get_by_text(item)).to_be_visible()
    print("  [OK] Sidebar navigation items visible")

    # Sign out button
    expect(page.locator("text=Sign out")).to_be_visible()
    print("  [OK] Sign out button visible")

    screenshot(page, "03_dashboard_full")

    # Check bento tiles
    tiles = page.locator(".bento-tile")
    print(f"  Bento tiles: {tiles.count()}")

    # Check card styling
    cards = page.locator(".card")
    print(f"  Cards: {cards.count()}")


def test_gallery_view(page: Page) -> None:
    """Test gallery view UI/UX."""
    print("\n=== Gallery View ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app/assets")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(800)
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
    screenshot(page, "04_gallery_list")

    # Toggle back to grid
    grid_btn.click()
    page.wait_for_timeout(300)
    screenshot(page, "04_gallery_grid")

    # Check empty state or content
    empty = page.locator("text=No assets yet").or_(page.locator("text=暂无资源"))
    if empty.count() > 0:
        print("  [OK] Empty state displayed")
    else:
        cards = page.locator(".card")
        print(f"  Gallery cards: {cards.count()}")

    # Check FavoriteButton
    favs = page.locator("button[aria-label*='favorites' i]")
    print(f"  Favorite buttons: {favs.count()}")


def test_generation_view(page: Page) -> None:
    """Test generation view UI/UX."""
    print("\n=== Generation View ===")
    set_authenticated(page)
    # Set English locale before navigation
    page.goto(f"{BASE_URL}/app/create")
    page.wait_for_load_state("networkidle")
    page.evaluate("() => localStorage.setItem('imageforge_locale', 'en')")
    page.goto(f"{BASE_URL}/app/create")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    # Form elements
    expect(page.locator("#prompt")).to_be_visible()
    expect(page.locator("#model")).to_be_visible()
    expect(page.locator("#size")).to_be_visible()
    print("  [OK] Generation form elements visible")

    # Image count buttons
    for n in ["1", "2", "4"]:
        expect(page.locator(".btn", has_text=n)).to_be_visible()
    print("  [OK] Image count buttons visible")
    # Advanced toggle - the button has chevron SVG + text, use partial match
    expect(page.locator("#negative")).to_have_count(0)
    # The button inner_text shows "+ Show" due to SVG, use has_text with partial match
    advanced_btn = page.locator("form button[type=button]", has_text="Show")
    if advanced_btn.count() == 0:
        advanced_btn = page.locator("form button[type=button]", has_text="高级")
    print(f"  Advanced buttons found: {advanced_btn.count()}")
    if advanced_btn.count() > 0:
        advanced_btn.click()
        page.wait_for_timeout(500)
        screenshot(page, "05_generation_advanced_clicked")
    neg_count = page.locator("#negative").count()
    print(f"  Negative prompt inputs: {neg_count}")
    expect(page.locator("#negative")).to_be_visible()
    print("  [OK] Advanced settings toggle works")
    # Fill prompt and check submit state
    page.locator("#prompt").fill("A beautiful sunset over mountains")
    page.wait_for_timeout(200)
    screenshot(page, "05_generation_filled")

    # Check model options
    model_select = page.locator("#model")
    options = model_select.locator("option")
    print(f"  Model options: {options.count()}")
    for i in range(options.count()):
        print(f"    - {options.nth(i).inner_text()}")


def test_collections_view(page: Page) -> None:
    """Test collections view UI/UX."""
    print("\n=== Collections View ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app/collections")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    screenshot(page, "06_collections")

    # Check for create button or empty state
    create_btn = page.locator("text=+ New Collection").or_(page.locator("text=+ 新建合集"))
    if create_btn.count() > 0:
        print("  [OK] Create collection button visible")


def test_tags_view(page: Page) -> None:
    """Test tags view UI/UX."""
    print("\n=== Tags View ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app/tags")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    screenshot(page, "07_tags")

    # Check for create button
    create_btn = page.locator("text=+ New Tag").or_(page.locator("text=+ 新建标签"))
    if create_btn.count() > 0:
        print("  [OK] Create tag button visible")

    # Open create form
    create_btn.click()
    page.wait_for_timeout(300)
    screenshot(page, "07_tags_create_form")


def test_favorites_view(page: Page) -> None:
    """Test favorites view UI/UX."""
    print("\n=== Favorites View ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app/favorites")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    screenshot(page, "08_favorites")

    # Check empty state
    empty = page.locator("text=No favorites yet").or_(page.locator("text=暂无收藏"))
    if empty.count() > 0:
        print("  [OK] Empty state displayed")


def test_templates_view(page: Page) -> None:
    """Test prompt templates view UI/UX."""
    print("\n=== Prompt Templates View ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app/templates")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    screenshot(page, "09_templates")

    # Check for create button
    create_btn = page.locator("text=+ New Template").or_(page.locator("text=+ 新建模板"))
    if create_btn.count() > 0:
        print("  [OK] Create template button visible")


def test_usage_view(page: Page) -> None:
    """Test usage view UI/UX."""
    print("\n=== Usage View ===")
    set_authenticated(page)
    page.goto(f"{BASE_URL}/app/usage")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)

    screenshot(page, "10_usage")

    # Check for stats
    stats = page.locator(".metric-value")
    print(f"  Usage stat cards: {stats.count()}")


def test_i18n_consistency(page: Page) -> None:
    """Test i18n consistency between EN and ZH."""
    print("\n=== i18n Consistency ===")
    set_authenticated(page)

    # Test EN
    page.goto(f"{BASE_URL}/app")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)
    page.evaluate("() => localStorage.setItem('imageforge_locale', 'en')")
    page.goto(f"{BASE_URL}/app")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)
    en_text = page.locator("h1").inner_text()
    print(f"  EN heading: {en_text}")
    screenshot(page, "11_i18n_en")

    # Test ZH
    page.locator("aside button", has_text="ZH").click()
    page.wait_for_timeout(500)
    zh_text = page.locator("h1").inner_text()
    print(f"  ZH heading: {zh_text}")
    screenshot(page, "11_i18n_zh")

    # Verify they're different (actual translation)
    assert en_text != zh_text, "EN and ZH headings should differ"
    print("  [OK] i18n translation working")


def test_responsive(page: Page) -> None:
    """Test responsive design at different viewports."""
    print("\n=== Responsive Design ===")
    set_authenticated(page)

    # Mobile viewport
    page.set_viewport_size({"width": 375, "height": 812})
    page.goto(f"{BASE_URL}/app")
    page.wait_for_load_state("networkidle")
    page.wait_for_timeout(500)
    screenshot(page, "12_responsive_mobile")

    # Tablet viewport
    page.set_viewport_size({"width": 768, "height": 1024})
    page.wait_for_timeout(500)
    screenshot(page, "12_responsive_tablet")

    # Desktop viewport
    page.set_viewport_size({"width": 1280, "height": 800})
    page.wait_for_timeout(500)
    screenshot(page, "12_responsive_desktop")

    # Wide viewport
    page.set_viewport_size({"width": 1920, "height": 1080})
    page.wait_for_timeout(500)
    screenshot(page, "12_responsive_wide")


def test_design_consistency(page: Page) -> None:
    """Check design system consistency across views."""
    print("\n=== Design Consistency ===")
    set_authenticated(page)

    views = [
        ("/app", "Dashboard"),
        ("/app/create", "Create"),
        ("/app/assets", "Gallery"),
        ("/app/projects", "Projects"),
        ("/app/collections", "Collections"),
        ("/app/tags", "Tags"),
        ("/app/favorites", "Favorites"),
        ("/app/templates", "Templates"),
        ("/app/usage", "Usage"),
    ]

    for path, name in views:
        page.goto(f"{BASE_URL}{path}")
        page.wait_for_load_state("networkidle")
        page.wait_for_timeout(400)

        # Check CSS variables are applied
        bg = page.evaluate(
            "() => getComputedStyle(document.documentElement).getPropertyValue('--bg').trim()"
        )
        text = page.evaluate(
            "() => getComputedStyle(document.documentElement).getPropertyValue('--text').trim()"
        )

        # Check heading style
        h1 = page.locator("h1")
        if h1.count() > 0:
            h1_font_size = page.evaluate(
                "el => getComputedStyle(el).fontSize", h1.element_handle()
            )
            print(f"  {name}: --bg={bg}, --text={text}, h1-size={h1_font_size}")
        else:
            print(f"  {name}: --bg={bg}, --text={text} (no h1)")

        screenshot(page, f"13_consistency_{name.lower()}")


def main() -> int:
    """Run all UI/UX review tests."""
    failed = 0
    tests = [
        ("Landing Page", test_landing_page),
        ("Login Page", test_login_page),
        ("Workspace Dashboard", test_workspace_dashboard),
        ("Gallery View", test_gallery_view),
        ("Generation View", test_generation_view),
        ("Collections View", test_collections_view),
        ("Tags View", test_tags_view),
        ("Favorites View", test_favorites_view),
        ("Templates View", test_templates_view),
        ("Usage View", test_usage_view),
        ("i18n Consistency", test_i18n_consistency),
        ("Responsive Design", test_responsive),
        ("Design Consistency", test_design_consistency),
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
    print(f"Screenshots saved to: {SCREENSHOT_DIR}")
    if failed:
        print(f"Failed: {failed}")
        return 1
    print("All UI/UX review tests passed!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
