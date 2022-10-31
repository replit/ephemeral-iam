# Releasing a new version

1. Increment the version number in `flake.nix`.
2. Create a new git tag for the new version. Ex: `v0.0.23`.
3. Push the git tag to GitHub.
4. Build `eiam` locally for both `amd64` (`GOOS=linux GOARCH=amd64`) and `arm64` (`GOOS=darwin GOARCH=arm64`).
5. Upload the build to the release for the tag in GitHub.
