"""
Conftest for tests/web directory.
Provides shared fixtures for Playwright tests.
"""

import subprocess
import time
import socket
from pathlib import Path

import pytest
from playwright.sync_api import sync_playwright, Page, Browser

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
FRONTEND_DIR = PROJECT_ROOT / "frontend"
DEV_SERVER_URL = "http://localhost:5173"


def wait_for_server(port: int, timeout: int = 30) -> bool:
    start = time.time()
    while time.time() - start < timeout:
        try:
            with socket.create_connection(("localhost", port), timeout=1):
                return True
        except (ConnectionRefusedError, socket.timeout):
            time.sleep(0.5)
    return False


@pytest.fixture(scope="session")
def browser():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        yield browser
        browser.close()


@pytest.fixture(scope="session")
def dev_server():
    proc = None
    if not wait_for_server(5173, timeout=3):
        proc = subprocess.Popen(
            ["pnpm", "dev"],
            cwd=str(FRONTEND_DIR),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            shell=True,
        )
        if not wait_for_server(5173, timeout=30):
            proc.terminate()
            raise RuntimeError("Failed to start dev server")
    yield proc
    if proc:
        proc.terminate()
        proc.wait(timeout=5)


@pytest.fixture
def page(browser: Browser, dev_server):
    context = browser.new_context()
    pg = context.new_page()
    yield pg
    context.close()
