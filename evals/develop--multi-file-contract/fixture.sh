#!/usr/bin/env bash
set -euo pipefail
python3 "$(cd "$(dirname "${BASH_SOURCE[0]}")/../_support" && pwd)/scaffold.py" develop--multi-file-contract
