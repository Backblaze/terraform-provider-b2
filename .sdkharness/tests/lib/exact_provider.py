"""Terraform CLI configuration for the exact provider revision under test."""

import json
import os
from pathlib import Path
from urllib.parse import urlsplit

DEAD_PROXY = "http://127.0.0.1:1"


def is_loopback_origin(value: str) -> bool:
    """A bare IPv4 loopback HTTP origin (http://127.0.0.1:<port>), nothing else."""
    url = urlsplit(value or "")
    try:
        port = url.port
    except ValueError:
        return False
    return (url.scheme == "http" and url.hostname == "127.0.0.1" and port is not None
            and url.username is None and url.password is None and url.path in ("", "/")
            and not url.query and not url.fragment)


def network_guard_env() -> dict:
    """Assertions may talk to the loopback simulator only: every proxy-aware client gets a dead
    proxy and loopback is exempt, so a connection to anything else fails at once. (`terraform
    init` would contact the provider registry even with dev_overrides, so checks never run it.)"""
    env = {"CHECKPOINT_DISABLE": "1", "NO_PROXY": "127.0.0.1,localhost,::1", "no_proxy": "127.0.0.1,localhost,::1"}
    for name in ("HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY"):
        env[name] = DEAD_PROXY
        env[name.lower()] = DEAD_PROXY
    return env


def write_cli_config(destination: str) -> None:
    """Point Terraform at the provider binary supplied by the orchestrator."""
    provider_dir = os.environ.get("SDKHARNESS_TERRAFORM_PROVIDER_DIR", "")
    provider = Path(provider_dir) / "terraform-provider-b2"
    if not provider.is_file() or not os.access(provider, os.X_OK):
        raise RuntimeError("exact-revision provider executable is unavailable")
    Path(destination).write_text(
        "provider_installation {\n"
        "  dev_overrides {\n"
        f'    "Backblaze/b2" = {json.dumps(provider_dir)}\n'
        "  }\n"
        "}\n",
        encoding="utf-8",
    )
