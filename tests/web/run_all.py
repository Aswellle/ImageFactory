#!/usr/bin/env python3
"""
ImageForge Web Test Runner

Usage:
    python tests/web/run_all.py          # Run all tests
    python tests/web/run_all.py landing  # Run landing tests only
    python tests/web/run_all.py auth     # Run auth tests only
    python tests/web/run_all.py views    # Run view tests only
"""

import sys
import subprocess
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent


def main():
    test_filter = sys.argv[1] if len(sys.argv) > 1 else ""

    cmd = [sys.executable, "-m", "pytest", "tests/web/", "-v", "--tb=short"]

    if test_filter:
        cmd.extend(["-k", test_filter])

    print(f"Running: {' '.join(cmd)}")
    print(f"Project root: {PROJECT_ROOT}")
    print("-" * 60)

    result = subprocess.run(cmd, cwd=str(PROJECT_ROOT))
    sys.exit(result.returncode)


if __name__ == "__main__":
    main()
