# DP-1 Validator

Command-line validator for DP-1 playlists and capsules.

Use this tool to establish trust before integration work: validate a playlist first, then validate capsule asset integrity when needed.

## Build and run (shortest path)

```bash
git clone https://github.com/display-protocol/dp1-validator.git
cd dp1-validator
go build -o dp1-validator .
./dp1-validator --help
```

## Compatibility note

- Canonical DP-1 protocol spec: `v1.1.0` in `display-protocol/dp1`.
- This validator is clearly verified in this repo for:
  - DP-1 `1.0.x`-style playlist structure validation
  - legacy top-level `signature` (`ed25519:<hex>`) verification when `--pubkey` is provided
  - capsule extraction and SHA256 directory hash comparison
- Current `1.0.x` examples in this README are **transitional ecosystem examples** and remain useful for first trust checks.
- Unverified here and should not be assumed:
  - end-to-end verification of DP-1 `v1.1.0` multi-signature chains (`signatures[]`) in this CLI
  - cross-repo parity claims with feed/operator implementations

For protocol truth, use: <https://github.com/display-protocol/dp1/blob/main/docs/spec.md>

For guided integration flow, use: <https://docs.feralfile.com/dp1-protocol/overview/>

## Quickstart: validate a playlist first

### 1) Validate structure from URL

```bash
./dp1-validator playlist --playlist "https://example.com/playlist.json"
```

### 2) Validate structure from base64 payload

```bash
PLAYLIST_B64="$(cat playlist.json | base64 | tr -d '\n')"
./dp1-validator playlist --playlist "$PLAYLIST_B64"
```

### 3) Verify legacy signature (optional)

```bash
./dp1-validator playlist \
  --playlist "https://example.com/playlist.json" \
  --pubkey "a1b2c3d4e5f6..."
```

Flags:

- `--playlist` (required): URL or base64 playlist payload
- `--pubkey` (optional): Ed25519 public key hex for legacy top-level signature verification

What output means:

- Success path includes `Playlist structure is valid`.
- If `--pubkey` is provided and legacy signature verifies, output includes `Playlist signature verification successful`.
- Typical failure classes:
  - parse/fetch errors: input could not be read as URL/base64/JSON
  - structure validation errors: required fields or formats are invalid
  - signature verification errors: key/signature mismatch or invalid format

## Capsule validation (secondary path)

Use capsule validation when you need to verify extracted assets against expected SHA256 hashes.

```bash
# Use hashes from playlist repro.assetsSHA256
./dp1-validator capsule --path "artwork.dp1c"

# Override with explicit hashes
./dp1-validator capsule \
  --path "artwork.dp1c" \
  --hashes "hash1,hash2,hash3"
```

Flags:

- `--path` (required): `.dp1c` file path
- `--hashes` (optional): override expected hashes; accepts comma/colon/bracket formats

Capsule requirements:

- `.dp1c` extension
- tar+zstd archive
- `playlist.json` at archive root
- `assets/` directory at archive root

## Development checks

```bash
go build ./...
go test ./...
```

## License

Mozilla Public License 2.0. See [LICENSE](LICENSE).
