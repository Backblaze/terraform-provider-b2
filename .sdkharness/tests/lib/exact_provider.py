"""Terraform CLI configuration for the exact provider revision under test."""

import json
import os
from pathlib import Path


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
        "  direct {}\n"
        "}\n",
        encoding="utf-8",
    )
