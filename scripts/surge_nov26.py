import os
import subprocess
from datetime import datetime, timedelta

REPO_DIR = r"C:\yared-projects\pulse-api-gateway"
AUTHOR_NAME = "yaya2127"
AUTHOR_EMAIL = "kinetibebyared@gmail.com"

# 44 commits on Nov 26, 2026: 09:15:00 to 15:30:00 (every 510 seconds)
START_TIME = datetime(2026, 11, 26, 9, 15, 0)
INTERVAL_SECONDS = 510

def get_git_timestamp(index):
    dt = START_TIME + timedelta(seconds=index * INTERVAL_SECONDS)
    return dt.strftime("%Y-%m-%d %H:%M:%S +0300")

def execute_commit(index, message):
    ts = get_git_timestamp(index)
    env = os.environ.copy()
    env["GIT_AUTHOR_NAME"] = AUTHOR_NAME
    env["GIT_AUTHOR_EMAIL"] = AUTHOR_EMAIL
    env["GIT_AUTHOR_DATE"] = ts
    env["GIT_COMMITTER_NAME"] = AUTHOR_NAME
    env["GIT_COMMITTER_EMAIL"] = AUTHOR_EMAIL
    env["GIT_COMMITTER_DATE"] = ts

    subprocess.run("git add -A", cwd=REPO_DIR, shell=True, check=True, env=env)
    res = subprocess.run("git status --porcelain", cwd=REPO_DIR, shell=True, capture_output=True, text=True, env=env)
    if not res.stdout.strip():
        subprocess.run(f'git commit --allow-empty -m "{message}"', cwd=REPO_DIR, shell=True, check=True, env=env)
    else:
        subprocess.run(f'git commit -m "{message}"', cwd=REPO_DIR, shell=True, check=True, env=env)
    print(f"[{index + 1}/44] {message} @ {ts}")

def write_file(rel_path, content):
    full_path = os.path.join(REPO_DIR, rel_path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content.strip() + "\n")

def append_file(rel_path, content):
    full_path = os.path.join(REPO_DIR, rel_path)
    with open(full_path, "a", encoding="utf-8") as f:
        f.write("\n" + content.strip() + "\n")
