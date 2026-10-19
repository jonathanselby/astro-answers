# Release checklist

## Contract

- [x] `VERSION` is `0.1.0`.
- [x] PRD is final for V0.1.0 and records V0.2 mobile direction without mobile implementation.
- [x] EDD is final and aligned to the V0.1.0 implementation.
- [x] V0.1 is Linux-only, full-screen and keyboard-only.
- [x] No network, telemetry, accounts, advertising or runtime AI.

## Engineering gate

- [x] `gofmt` clean.
- [ ] `go vet ./...` passes in release CI.
- [ ] `go test -race -count=1 ./...` passes in release CI.
- [ ] Linux/amd64 build passes in release CI.
- [ ] Linux/arm64 build passes in release CI.
- [x] core gameplay tests pass independently of the renderer.
- [x] full-screen Linux gameplay/audio smoke test completed during pre-release testing.
- [x] final visual-only change is limited to dark-grey styled-text outlines; text fill and cyan structural lines are unchanged.

## Evidence

- [x] PRD acceptance criteria map to tests or observable play evidence.
- [ ] generated EDD pins the exact release source snapshot.
- [ ] generated comparison explains design/implementation matches and corrections.
- [x] design process is recorded in `DESIGN-NOTES.md`.
- [x] automated, play-test and failed-harness evidence is recorded in `TESTING.md`.
- [ ] constrained-hardware benchmark recorded on the standard Garage target.

## Release rule

The source may be published before the constrained-hardware run, but the Garage Definition of Done is not fully closed until the benchmark evidence is added. The release tag should follow green CI and code-derived evidence; the physical benchmark remains an explicit open evidence item rather than a fabricated claim.
