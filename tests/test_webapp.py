"""
ImageForge Web Application Test Suite
Tests the key user flows of the ImageForge web application.
"""
import sys
from playwright.sync_api import sync_playwright, expect

BASE_URL = "http://localhost:5173"
SCREENSHOT_DIR = "/tmp/imageforge_tests"

# Test results tracking
results = {"passed": 0, "failed": 0, "errors": []}


def take_screenshot(page, name: str):
    """Take a screenshot for debugging."""
    import os
    os.makedirs(SCREENSHOT_DIR, exist_ok=True)
    path = f"{SCREENSHOT_DIR}/{name}.png"
    page.screenshot(path=path, full_page=True)
    return path


def set_locale(page, locale: str):
    """Set locale in localStorage."""
    page.evaluate(f"localStorage.setItem('imageforge_locale', '{locale}')")


def run_test(test_func):
    """Decorator to run a test and track results."""
    def wrapper(page):
        test_name = test_func.__name__
        try:
            test_func(page)
            results["passed"] += 1
            print(f"  ✅ {test_name}")
            return True
        except Exception as e:
            results["failed"] += 1
            results["errors"].append((test_name, str(e)))
            print(f"  ❌ {test_name}: {e}")
            take_screenshot(page, f"error_{test_name}")
            return False
    return wrapper


# ============================================================
# Landing Page Tests
# ============================================================

@run_test
def test_landing_page_loads(page):
    """Test that the landing page loads correctly."""
    page.goto(BASE_URL, timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    # Check main elements
    expect(page.locator(".nav-content a[href='/']")).to_be_visible(timeout=5000)
    expect(page.locator("text=Create images")).to_be_visible(timeout=5000)


@run_test
def test_landing_navigation(page):
    """Test landing page navigation links."""
    page.goto(BASE_URL, timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    # Check navigation links (use nav container to be specific)
    nav = page.locator(".nav-content")
    expect(nav.locator("a[href='#features']")).to_be_visible(timeout=5000)
    expect(nav.locator("a[href='#workflow']")).to_be_visible(timeout=5000)


@run_test
def test_landing_features_section(page):
    """Test that features section is accessible."""
    page.goto(BASE_URL, timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    # Scroll to features
    page.locator("#features").scroll_into_view_if_needed()
    expect(page.locator("#features")).to_be_visible(timeout=5000)


# ============================================================
# Login Page Tests
# ============================================================

@run_test
def test_login_page_loads(page):
    """Test that the login page loads correctly."""
    page.goto(f"{BASE_URL}/login", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    expect(page.locator("h1")).to_contain_text("ImageForge")
    expect(page.locator("label[for='email']")).to_be_visible(timeout=5000)
    expect(page.locator("label[for='password']")).to_be_visible(timeout=5000)
    expect(page.locator("button[type='submit']")).to_be_visible(timeout=5000)


@run_test
def test_login_password_visibility_toggle(page):
    """Test password visibility toggle button."""
    page.goto(f"{BASE_URL}/login", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    password_input = page.locator("#password")
    toggle_btn = page.locator(".relative button[aria-label*='password']")
    
    # Initially password should be hidden
    expect(password_input).to_have_attribute("type", "password", timeout=5000)
    expect(toggle_btn).to_have_attribute("aria-pressed", "false", timeout=5000)
    
    # Click to show password
    toggle_btn.click()
    expect(password_input).to_have_attribute("type", "text", timeout=5000)
    expect(toggle_btn).to_have_attribute("aria-pressed", "true", timeout=5000)
    
    # Click to hide password
    toggle_btn.click()
    expect(password_input).to_have_attribute("type", "password", timeout=5000)
    expect(toggle_btn).to_have_attribute("aria-pressed", "false", timeout=5000)


@run_test
def test_login_forgot_password_link(page):
    """Test forgot password link navigation."""
    page.goto(f"{BASE_URL}/login", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    forgot_link = page.locator("a[href='/forgot-password']")
    expect(forgot_link).to_be_visible(timeout=5000)
    forgot_link.click()
    
    # Should navigate to forgot password page
    expect(page).to_have_url(f"{BASE_URL}/forgot-password")
    expect(page.locator("text=Reset password")).to_be_visible(timeout=5000)


@run_test
def test_login_empty_submit(page):
    """Test login form validation on empty submit."""
    page.goto(f"{BASE_URL}/login", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    # Submit empty form
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(1000)
    
    # Browser built-in validation should prevent submission
    # Check that we're still on the login page (form not submitted)
    expect(page).to_have_url(f"{BASE_URL}/login")


# ============================================================
# Register Page Tests
# ============================================================

@run_test
def test_register_page_loads(page):
    """Test that the register page loads correctly."""
    page.goto(f"{BASE_URL}/login", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    # Switch to register mode
    page.locator(".text-center button:has-text('Sign up')").click()
    
    expect(page.locator("label[for='name']")).to_be_visible(timeout=5000)
    expect(page.locator("label[for='email']")).to_be_visible(timeout=5000)
    expect(page.locator("label[for='password']")).to_be_visible(timeout=5000)
    expect(page.locator("label[for='confirmPassword']")).to_be_visible(timeout=5000)
    expect(page.locator("button[type='submit']")).to_be_visible(timeout=5000)


@run_test
def test_register_password_strength_meter(page):
    """Test password strength meter appears in register mode."""
    page.goto(f"{BASE_URL}/login", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    # Switch to register mode
    page.locator(".text-center button:has-text('Sign up')").click()
    
    # Type a password
    password_input = page.locator("#password")
    password_input.fill("Test123!")
    
    # Strength meter should appear
    expect(page.locator(".password-strength")).to_be_visible(timeout=5000)


@run_test
def test_register_password_mismatch(page):
    """Test password mismatch validation."""
    page.goto(f"{BASE_URL}/login", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    # Switch to register mode
    page.locator(".text-center button:has-text('Sign up')").click()
    
    # Fill form with mismatched passwords
    page.locator("#name").fill("Test User")
    page.locator("#email").fill("test@example.com")
    page.locator("#password").fill("Test123!")
    page.locator("#confirmPassword").fill("Different123!")
    
    # Submit
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(1000)
    
    # Should show error
    expect(page.locator("text=Passwords do not match")).to_be_visible(timeout=5000)


# ============================================================
# Forgot Password Page Tests
# ============================================================

@run_test
def test_forgot_password_page_loads(page):
    """Test that the forgot password page loads correctly."""
    page.goto(f"{BASE_URL}/forgot-password", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    expect(page.locator("text=Reset password")).to_be_visible(timeout=5000)
    expect(page.locator("label[for='reset-email']")).to_be_visible(timeout=5000)
    expect(page.locator("button[type='submit']")).to_be_visible(timeout=5000)


@run_test
def test_forgot_password_back_to_login(page):
    """Test back to login link."""
    page.goto(f"{BASE_URL}/forgot-password", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    back_link = page.locator("text=← Back to login")
    expect(back_link).to_be_visible(timeout=5000)
    back_link.click()
    
    # Should navigate to login page
    expect(page).to_have_url(f"{BASE_URL}/login")


# ============================================================
# Language Switching Tests
# ============================================================

@run_test
def test_language_switch_to_chinese(page):
    """Test switching language to Chinese."""
    page.goto(f"{BASE_URL}/login", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "zh")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    # Check Chinese text
    expect(page.locator("button:has-text('登录')")).to_be_visible(timeout=5000)
    expect(page.locator("button:has-text('注册')")).to_be_visible(timeout=5000)


@run_test
def test_language_switch_to_english(page):
    """Test switching language back to English."""
    page.goto(f"{BASE_URL}/login", timeout=10000)
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    set_locale(page, "en")
    page.reload()
    page.wait_for_load_state("domcontentloaded", timeout=15000)
    
    # Check English text
    expect(page.locator("button:has-text('Sign In')")).to_be_visible(timeout=5000)
    expect(page.locator("button:has-text('Sign up')")).to_be_visible(timeout=5000)


# ============================================================
# Main Test Runner
# ============================================================

def main():
    """Run all tests."""
    print("=" * 60)
    print("ImageForge Web Application Test Suite")
    print("=" * 60)
    
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={"width": 1280, "height": 720})
        page = context.new_page()
        
        # Landing Page Tests
        print("\n📍 Landing Page Tests")
        print("-" * 40)
        run_test(test_landing_page_loads)(page)
        run_test(test_landing_navigation)(page)
        run_test(test_landing_features_section)(page)
        
        # Login Page Tests
        print("\n🔐 Login Page Tests")
        print("-" * 40)
        run_test(test_login_page_loads)(page)
        run_test(test_login_password_visibility_toggle)(page)
        run_test(test_login_forgot_password_link)(page)
        run_test(test_login_empty_submit)(page)
        
        # Register Page Tests
        print("\n📝 Register Page Tests")
        print("-" * 40)
        run_test(test_register_page_loads)(page)
        run_test(test_register_password_strength_meter)(page)
        run_test(test_register_password_mismatch)(page)
        
        # Forgot Password Page Tests
        print("\n🔑 Forgot Password Page Tests")
        print("-" * 40)
        run_test(test_forgot_password_page_loads)(page)
        run_test(test_forgot_password_back_to_login)(page)
        
        # Language Switching Tests
        print("\n🌐 Language Switching Tests")
        print("-" * 40)
        run_test(test_language_switch_to_chinese)(page)
        run_test(test_language_switch_to_english)(page)
        
        # Clean up
        browser.close()
    
    # Print summary
    print("\n" + "=" * 60)
    print("Test Summary")
    print("=" * 60)
    total = results["passed"] + results["failed"]
    print(f"Total: {total}")
    print(f"Passed: {results['passed']}")
    print(f"Failed: {results['failed']}")
    
    if results["errors"]:
        print("\n❌ Failed Tests:")
        for name, error in results["errors"]:
            print(f"  - {name}: {error}")
    
    print(f"\nScreenshots saved to: {SCREENSHOT_DIR}")
    
    # Exit with error code if any tests failed
    sys.exit(1 if results["failed"] > 0 else 0)


if __name__ == "__main__":
    main()
