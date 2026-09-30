# Releases

UniQUE uses Release Please to prepare releases from Conventional Commits. A
push to `main` creates or updates one release pull request containing:

- the next semantic version in `.release-please-manifest.json`;
- matching `version` and `appVersion` values in `deploy/helm/unique/Chart.yaml`;
- generated release notes in `CHANGELOG.md`.

Merging the release pull request creates the corresponding GitHub Release and
`vX.Y.Z` tag. The tag-triggered workflow then publishes all six versioned
container images and the OCI Helm chart. Do not create release tags manually.

## Repository setup

Create a GitHub App for release automation with these repository permissions:

- Contents: read and write
- Issues: read and write
- Pull requests: read and write
- Metadata: read

Install the App on this repository, generate a private key, and configure:

- Actions variable `RELEASE_APP_CLIENT_ID`: the App client ID
- Actions secret `RELEASE_APP_PRIVATE_KEY`: the complete PEM private key
- Actions secret `HARBOR_TOKEN`: password for `robot$github-publisher`

The App token is required because pull requests and tags created with the
built-in `GITHUB_TOKEN` do not trigger subsequent GitHub Actions workflows.

## Version selection

- `fix:` produces a patch release.
- `feat:` produces a minor release.
- `feat!:`, `fix!:`, or a `BREAKING CHANGE:` footer produces a major release.

Other Conventional Commit types may appear in the notes without changing the
version. The release pull request remains open and is updated as releasable
commits reach `main`; merging it is the explicit approval to publish.
