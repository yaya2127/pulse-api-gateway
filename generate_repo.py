import os
import subprocess
from datetime import datetime, timedelta

ROOT = r"C:\yared-projects\pulse-api-gateway"
AUTHOR = "yaya2127"
EMAIL = "kinetibebyared@gmail.com"

start_time = datetime(2026, 11, 23, 9, 0, 0)
step_sec = 230

def git_commit(idx, msg):
    cur_time = start_time + timedelta(seconds=idx * step_sec)
    ts = cur_time.strftime("%Y-%m-%d %H:%M:%S +0300")
    env = os.environ.copy()
    env["GIT_AUTHOR_NAME"] = AUTHOR
    env["GIT_AUTHOR_EMAIL"] = EMAIL
    env["GIT_AUTHOR_DATE"] = ts
    env["GIT_COMMITTER_NAME"] = AUTHOR
    env["GIT_COMMITTER_EMAIL"] = EMAIL
    env["GIT_COMMITTER_DATE"] = ts

    subprocess.run("git add -A", cwd=ROOT, shell=True, check=True, env=env)
    res = subprocess.run("git status --porcelain", cwd=ROOT, shell=True, capture_output=True, text=True, env=env)
    if not res.stdout.strip():
        subprocess.run(f'git commit --allow-empty -m "{msg}"', cwd=ROOT, shell=True, check=True, env=env)
    else:
        subprocess.run(f'git commit -m "{msg}"', cwd=ROOT, shell=True, check=True, env=env)
    print(f"[{idx}/88] {msg}")

def write_code(rel_path, content):
    full_path = os.path.join(ROOT, rel_path)
    os.makedirs(os.path.dirname(full_path), exist_ok=True)
    with open(full_path, "w", encoding="utf-8") as f:
        f.write(content.strip() + "\n")

def append_code(rel_path, content):
    full_path = os.path.join(ROOT, rel_path)
    with open(full_path, "a", encoding="utf-8") as f:
        f.write("\n" + content.strip() + "\n")

print("Generator script template ready.")
