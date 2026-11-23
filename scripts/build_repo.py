import os
import subprocess
import time
from datetime import datetime, timedelta

BASE_DIR = r"c:\yared-projects\pulse-api-gateway"
GIT_AUTHOR_NAME = "yaya2127"
GIT_AUTHOR_EMAIL = "kinetibebyared@gmail.com"

def run_git(cmd, env_vars=None):
    env = os.environ.copy()
    if env_vars:
        env.update(env_vars)
    res = subprocess.run(cmd, cwd=BASE_DIR, shell=True, capture_output=True, text=True, env=env)
    if res.returncode != 0:
        print(f"Git command failed: {cmd}\nStderr: {res.stderr}")
        raise RuntimeError(res.stderr)
    return res.stdout.strip()

def commit(date_str, message):
    env_vars = {
        "GIT_AUTHOR_NAME": GIT_AUTHOR_NAME,
        "GIT_AUTHOR_EMAIL": GIT_AUTHOR_EMAIL,
        "GIT_AUTHOR_DATE": date_str,
        "GIT_COMMITTER_NAME": GIT_AUTHOR_NAME,
        "GIT_COMMITTER_EMAIL": GIT_AUTHOR_EMAIL,
        "GIT_COMMITTER_DATE": date_str,
    }
    run_git("git add .", env_vars)
    # Check if there are changes staged
    status = run_git("git status --porcelain")
    if not status:
        # If nothing changed, touch a micro file or allow empty commit
        run_git(f'git commit --allow-empty -m "{message}"', env_vars)
    else:
        run_git(f'git commit -m "{message}"', env_vars)
    print(f"Committed: {message} @ {date_str}")

print("Git helper ready.")
