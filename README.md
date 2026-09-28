# pensieve-go-core
Shared code for the Go services in Pennsieve

## Realtime events (`pkg/realtime`)

Live updates (upload rows, workflow run status, app builds) go through `pkg/realtime`, not Pusher directly. The design is in [docs/realtime-appsync-design.md](docs/realtime-appsync-design.md).

```go
pub := realtime.FromEnv(ctx, realtime.WithPusher(buildPusherClient)) // WithPusher only if the service still supports Pusher

err := pub.Publish(ctx, realtime.Run(run.OrganizationId, run.CreatedBy, run.Uuid),
    "workflow-run-status", event)
if err != nil {
    log.Warn(err) // never fail the request over a live update
}
```

- **Channels:**
  - `realtime.Dataset(id)` → `/datasets/<uuid>`
  - `realtime.Run(org, createdBy, run)` → `/runs/org-<uuid>/<run>`, or `/runs/user-<uuid>/<run>` without a workspace
  - `realtime.Application(id)` → `/applications/<uuid>`
  - `realtime.RunScope(org, createdBy)` → `/runs/<scope>/*`, for subscribers only

  Node IDs (`N:dataset:…`) and bare UUIDs are both accepted.
- **Configuration:** `REALTIME_PROVIDER` selects the publisher:
  - `appsync`: the Event API. Needs `REALTIME_EVENTS_ENDPOINT` (its HTTP DNS name) and `appsync:EventPublish`; it signs with the Lambda's own credentials.
  - `pusher`: today's Pusher channels, through the client passed with `WithPusher`.
  - `noop`: the default.

  A missing or broken configuration returns `noop` with a warning. The service still starts and only loses live updates.
- **Payloads:** subscribers on AppSync receive `{"event": "<name>", "data": <payload>}`. The Pusher adapter sends the payload as before.

## Releasing a new version

1. Merge updates into the main branch
2. Create a new tag in main and name the tag: vx.x.x following [semantic versioning](https://semver.org/).

   e.g ```git tag -a v0.0.1 -m "Initial release"```

   Given a version number MAJOR.MINOR.PATCH, increment the:

    - MAJOR version when you make incompatible API changes,
    - MINOR version when you add functionality in a backwards compatible manner, and
    - PATCH version when you make backwards compatible bug fixes.

3. Push the tag to Gihub

   eg. ```git push origin v0.0.1```

This will trigger Github Actions to create a new release with the same name.