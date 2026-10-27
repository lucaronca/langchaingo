# langchaingo fork

This repository is a fork of [tmc/langchaingo](https://github.com/tmc/langchaingo). It carries changes that are not merged in tmc/langchaingo yet, and the services that need them consume it as a version-tagged module through a `replace` directive:

```
replace github.com/tmc/langchaingo => github.com/lucaronca/langchaingo v0.0.N
```

The fork adds APIs that tmc/langchaingo does not have, such as `chains.WithSafetyConfig`, `llms.SafetyConfig`, and `StartSync`, `UploadDocuments`, `UploadNamedDocuments` and `ErrIngestionJobInProgress` in `vectorstores/bedrockknowledgebases`. Pointing the `replace` directive at a tmc/langchaingo tag or at a local path therefore breaks the build of those services.

## Branches

```
main ─┬─ introduce-safety-configuration ── bedrock-cache-support ─┐
      │                                                           ├─ release (tags)
      └─ improve-bedrock-ingestion ────────────────────────────────┘
```

1. `main` mirrors tmc/langchaingo `main` and has no commits of its own, so it is always updated by fast-forward.
2. Each change lives on its own branch, which is the head of a pull request to tmc/langchaingo:
   1. `introduce-safety-configuration` is tmc/langchaingo#1489, the safety configuration of LLM calls, implemented in the Bedrock LLM through Bedrock Guardrails.
   2. `bedrock-cache-support` is tmc/langchaingo#1498, the Anthropic prompt caching on Bedrock. It is built on top of `introduce-safety-configuration` and contains its commits, so the pull request shows them too.
   3. `improve-bedrock-ingestion` reworks the ingestion of `vectorstores/bedrockknowledgebases`. `AddDocuments` and `AddNamedDocuments` ingest the documents directly and wait until they are indexed, run the first sync of a data source that was never synced, and return `ErrIngestionJobInProgress` before uploading when a sync is running. It adds `StartSync`, `UploadDocuments` and `UploadNamedDocuments`, and returns numeric metadata as `float32`, which recent AWS SDK versions decode as `float64`. It sits directly on `main`, because it depends on neither of the other two, and its pull request to tmc/langchaingo is not opened yet.
3. `release` merges the change branches and is the only branch that contains all of them. Every version tag is cut on it.

## Versions

Tags are lightweight, named `v0.0.N`, and cut on `release`. Each one has a GitHub release whose description gives the tmc/langchaingo commit it is based on and the changes it carries that are not merged in tmc/langchaingo yet.

## Updating from tmc/langchaingo

1. Run `git fetch upstream`, where `upstream` is tmc/langchaingo, and fast-forward `main` to `upstream/main`.
2. Merge `main` into `introduce-safety-configuration`, then `introduce-safety-configuration` into `bedrock-cache-support`, then `main` into `improve-bedrock-ingestion`. The change branches are updated with merges, never with rebases, so that the open pull requests never need a force-push. tmc/langchaingo squashes external pull requests, so these merge commits do not reach its history.
3. Merge `bedrock-cache-support` and `improve-bedrock-ingestion` into `release`.
4. Run the checks below on `release`, cut the next tag on it, and publish the GitHub release.
5. Push the branches and the tag by name (`git push origin refs/tags/v0.0.N`), never with `--tags`, because the fetch from tmc/langchaingo also brings its own `v0.1.x` tags into the clone.

## Adding a change

Create a branch from `main`, open its pull request to tmc/langchaingo, then merge the branch into `release` and cut a new tag.

## Checks

1. Lint with golangci-lint v2.9.0, the version used by the tmc/langchaingo CI: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0 run ./...`.
2. Test with `make test`, with Docker running, because several vector store tests start containers through testcontainers.
3. `TestInvokeModelInputs` in `llms/bedrock/internal/bedrockclient` fails since tmc/langchaingo#1489 and is a known failure. The test expects "bedrock safety" error texts, the guardrail `trace` detail as a map and a content filter error for a Titan `CONTENT_FILTERED` response, while the code writes "bedrock guardrail" texts, stores `trace` as a JSON string and returns no error for the Titan case. The tmc/langchaingo CI does not report it, because its test step pipes `go test` into `tee` without `pipefail`.
