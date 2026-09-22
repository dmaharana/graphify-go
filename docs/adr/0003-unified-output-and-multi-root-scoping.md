# Unified Output Directory and Multi-Root Scoping

We decided to standardize all generated artifacts (`graph.json`, `GRAPH_REPORT.md`, `graph.html`, and `manifest.json`) in a local `graphify-out/` directory, and support multi-root directory scanning. Node identifiers and source paths are normalized relative to each scanned directory root, keeping outputs isolated from source trees while supporting multi-package repository layouts.
