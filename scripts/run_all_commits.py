import os
import sys
import subprocess
from datetime import datetime, timedelta

BASE_DIR = r"c:\yared-projects\pulse-api-gateway"
GIT_AUTHOR_NAME = "yaya2127"
GIT_AUTHOR_EMAIL = "kinetibebyared@gmail.com"

start_dt = datetime(2026, 11, 23, 9, 15, 0)
step_seconds = (5 * 3600 + 40 * 60) // 88  # ~231s per commit

def get_timestamp(index):
    dt = start_dt + timedelta(seconds=index * step_seconds)
    return dt.strftime("%Y-%m-%d %H:%M:%S +0300")

def run_cmd(cmd, env=None):
    res = subprocess.run(cmd, cwd=BASE_DIR, shell=True, capture_output=True, text=True, env=env)
    if res.returncode != 0:
        print(f"Error executing: {cmd}\n{res.stderr}")
        raise RuntimeError(res.stderr)
    return res.stdout.strip()

def commit(index, message):
    ts = get_timestamp(index)
    env = os.environ.copy()
    env["GIT_AUTHOR_NAME"] = GIT_AUTHOR_NAME
    env["GIT_AUTHOR_EMAIL"] = GIT_AUTHOR_EMAIL
    env["GIT_AUTHOR_DATE"] = ts
    env["GIT_COMMITTER_NAME"] = GIT_AUTHOR_NAME
    env["GIT_COMMITTER_EMAIL"] = GIT_AUTHOR_EMAIL
    env["GIT_COMMITTER_DATE"] = ts
    
    run_cmd("git add -A", env)
    status = run_cmd("git status --porcelain")
    if not status:
        run_cmd(f'git commit --allow-empty -m "{message}"', env)
    else:
        run_cmd(f'git commit -m "{message}"', env)
    print(f"[{index+1}/88] {message} @ {ts}")

def write_f(path, content):
    full = os.path.join(BASE_DIR, path)
    os.makedirs(os.path.dirname(full), exist_ok=True)
    with open(full, "w", encoding="utf-8") as f:
        f.write(content.strip() + "\n")

def append_f(path, content):
    full = os.path.join(BASE_DIR, path)
    with open(full, "a", encoding="utf-8") as f:
        f.write("\n" + content.strip() + "\n")

print("Initializing commit runner...")
