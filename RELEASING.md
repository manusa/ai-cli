# Release Process

## Active development

Active development happens on the `main` branch.
All new features and bug fixes should be merged into `main` first.

Main should remain stable and deployable at all times.

## Versioning

The project uses [Semantic Versioning](https://semver.org/).
While the project is in proof of concept mode, only patch versions will be released.
These versions can include new features and breaking changes.

## Creating a release

1. Ensure all changes are merged into `main`.
2. Check the latest released version in the [releases page](https://github.com/manusa/ai-cli/releases).
3. Create a new Release using the GitHub UI:
   - Go to the [releases page](https://github.com/manusa/ai-cli/releases).
   - Click "Draft a new release".
   - Set (+create) the appropriate tag version (e.g., `v0.0.1337`) following semantic versioning.
     > [!NOTE]
     > Note that the `v` prefix is required.
   - Press the "Generate release notes" button to auto-generate the changelog.
   - Review and edit the release notes as needed.
   - Click "Publish release" to finalize.
