#!/bin/bash
set -euo pipefail

./build.sh all
exec ./go-cesi
