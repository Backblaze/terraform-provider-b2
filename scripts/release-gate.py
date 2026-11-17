#!/usr/bin/env python3
"""Fail-closed release gates. Report counts only, never private finding contents."""

import argparse
import json
import subprocess
import sys


def fetch(endpoint):
    try:
        result = subprocess.run(["gh", "api", "--paginate", "--slurp", endpoint],
                                check=True, capture_output=True, text=True)
        pages = json.loads(result.stdout)
        if not isinstance(pages, list) or any(not isinstance(p, list) for p in pages):
            raise ValueError("not a paginated list")
        rows = [row for page in pages for row in page]
        if any(not isinstance(row, dict) for row in rows):
            raise ValueError("not an alert object")
        return rows
    except (OSError, subprocess.CalledProcessError, ValueError) as from_error:
        raise RuntimeError("Security/provenance source unreadable; release held") from from_error


def blocking_count(source, rows):
    count = 0
    for row in rows:
        if source == "secrets":
            count += 1
            continue
        if source == "advisories" and row.get("state") in ("published", "closed", "withdrawn"):
            continue
        if source == "code":
            severity = (row.get("rule") or {}).get("security_severity_level")
        elif source == "dependencies":
            severity = (row.get("security_advisory") or {}).get("severity")
        else:
            severity = row.get("severity")
        # Unknown severity cannot establish that the release meets the policy.
        if severity not in ("low", "none"):
            count += 1
    return count


def matches_release(pr, repo, commit, tag):
    return bool(pr.get("merged_at") and pr.get("merge_commit_sha") == commit
                and pr.get("base", {}).get("ref") == "master"
                and (pr.get("base", {}).get("repo") or {}).get("full_name") == repo
                and pr.get("head", {}).get("ref") == "release/" + tag
                and (pr.get("head", {}).get("repo") or {}).get("full_name") == repo
                and (pr.get("milestone") or {}).get("title") == tag)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("security", "provenance"))
    parser.add_argument("--repo", default="Backblaze/terraform-provider-b2")
    parser.add_argument("--tag")
    args = parser.parse_args()
    try:
        if args.mode == "provenance":
            if not args.tag:
                raise RuntimeError("Release tag required")
            commit = subprocess.run(["git", "rev-parse", args.tag + "^{commit}"],
                                    check=True, capture_output=True, text=True).stdout.strip()
            subprocess.run(["git", "merge-base", "--is-ancestor", commit, "origin/master"],
                           check=True, capture_output=True)
            prs = fetch(f"repos/{args.repo}/commits/{commit}/pulls?per_page=100")
            if not any(matches_release(pr, args.repo, commit, args.tag) for pr in prs):
                raise RuntimeError("Tag must name the merge commit of a release/vX.Y.Z PR with its version milestone")
            print("Release provenance: PASS")
            return 0
        endpoints = {
            "code": "code-scanning/alerts?state=open&ref=refs%2Fheads%2Fmaster&per_page=100",
            "dependencies": "dependabot/alerts?state=open&per_page=100",
            "secrets": "secret-scanning/alerts?state=open&per_page=100",
            "advisories": "security-advisories?per_page=100",
        }
        held = False
        for source, path in endpoints.items():
            count = blocking_count(source, fetch(f"repos/{args.repo}/{path}"))
            print(f"{source}: {count} blocking finding(s)")
            held |= count > 0
        print("Security gate: HOLD" if held else "Security gate: PASS")
        return int(held)
    except (RuntimeError, subprocess.CalledProcessError):
        print("Release gate: HOLD (required verification failed or a source is unreadable)", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
