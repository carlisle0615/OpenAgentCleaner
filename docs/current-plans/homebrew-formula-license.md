# Homebrew Formula License

## Goal

Add explicit MIT license metadata to the Homebrew formula so the published tap file is more complete for open-source distribution.

## Touched Files

- `docs/current-plan.md`
- `docs/current-plans/homebrew-formula-license.md`
- `Formula/oac.rb`

## Validation Steps

1. Read `Formula/oac.rb` to confirm the `license "MIT"` field is present in the class definition.
2. Check the formula syntax visually for a valid field order and unchanged release URLs and checksums.

## Doc-Sync Steps

- Keep `docs/current-plan.md` aligned with the task status.
- No additional doc updates are needed unless the Homebrew publishing workflow changes.
