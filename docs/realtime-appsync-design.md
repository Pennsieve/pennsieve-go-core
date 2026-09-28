# Realtime events on AWS AppSync Events

Status: design, 2026-09-28. Tracks ClickUp [868makp65](https://app.clickup.com/t/868makp65); this document settles subtask 1/9 ([868mambm6](https://app.clickup.com/t/868mambm6)).

## Why

Pennsieve pushes live updates (upload rows, workflow status, app build status) through Pusher. That doesn't hold for a HIPAA environment:

- Pusher does not sign BAAs. Its HIPAA path is `private-encrypted-` channels only.
- Every channel is **public**: no subscription authorization. Anyone with the (public) app key and a channel's UUID can subscribe.
- `upload-event` payloads carry **file names**.

AppSync is on AWS's HIPAA-eligible services list, so events stay inside AWS under the BAA. AppSync Events adds the subscription authorization Pusher channels never had.

## Decisions

### 1. Authorization: one Lambda authorizer, reusing the api2 authorizer

The Event API uses **`AWS_LAMBDA` for connect and subscribe** and **`AWS_IAM` for publish**.

- A new entry point, `cmd/events-authorizer`, lives in **pennsieve-go-api** next to `cmd/websocket-authorizer`, with the same build, publish and Terraform pattern. Clin already runs `prod-pennsieve-go-api-websocket-authorizer-lambda-use1` in Pennsieve-VPC.
- It reuses what the HTTP and WebSocket authorizers already share:
  - `validateCognitoJWT` accepts **both** Cognito pools, the user pool (web app) and the token pool (agent, API key/secret). One Cognito provider on the Event API couldn't do this, which is why we don't use `AMAZON_COGNITO_USER_POOLS`.
  - `ClaimsManager.GetCurrentUser` resolves the Cognito identity to the Pennsieve user (Postgres).
  - `DatasetAuthorizer` and `WorkspaceAuthorizer` give the role checks.
- AppSync hands the Lambda the operation and the **exact channel** (`requestContext.operation` = `EVENT_CONNECT` | `EVENT_SUBSCRIBE` | `EVENT_PUBLISH`, `requestContext.channel`, `channelNamespaceName`), and caches the answer per API, channel, operation and token. So the per-channel check is a plain authorizer decision; no `onSubscribe` handler or data source is needed.
- **Rules:**

  | Operation | Allow when |
  |---|---|
  | `EVENT_CONNECT` | The token is a valid Pennsieve JWT for this environment |
  | `EVENT_SUBSCRIBE` `/datasets/<datasetUuid>` | `DatasetAuthorizer(N:dataset:<uuid>)` gives the caller at least viewer |
  | `EVENT_SUBSCRIBE` `/runs/org-<orgUuid>/<runUuid>` or `/runs/org-<orgUuid>/*` | `WorkspaceAuthorizer(N:organization:<orgUuid>)`: the caller is a member |
  | `EVENT_SUBSCRIBE` `/runs/user-<userUuid>/<runUuid>` or `/runs/user-<userUuid>/*` | The caller *is* that user (runs with no workspace) |
  | `EVENT_SUBSCRIBE` `/applications/<appUuid>` | app-deploy-service's `CanAccessApp` allows it (public, owner, or a grant to the user, one of their workspaces or one of their teams); see *Applications* below |
  | `EVENT_PUBLISH` (any) | Never. Publish is IAM-only, so the Lambda denies it if ever asked |
  | Any other channel containing `*` | Deny. AppSync passes wildcard subscriptions through literally. The only wildcard allowed is a whole run scope, `/runs/<scope>/*`, which the scope rule above covers |

- **Caching:** `authorizer_result_ttl_in_seconds = 300`. Losing access therefore takes up to 5 minutes to stop *new* subscriptions. An existing subscription lives until the connection drops. Tokens expire within an hour, and clients re-subscribe on reconnect with a fresh token. Document this; don't try to revoke live subscriptions.
- The client token goes in the WebSocket subprotocol header, **not** the URL, unlike the chat WebSocket's `?token=`. So it never lands in access logs.

**Applications.** Some apps are available across workspaces. App access lives in app-deploy-service (`CanAccessApp`: `visibility == "public"`, the owner, or a grant in the app-access table for `user#`, `workspace#` or `team#`). The authorizer doesn't duplicate those rules. It invokes a small **check-app-access** Lambda that app-deploy-service owns, the same way the websocket authorizer already invokes account-service's check-access Lambda for compute nodes.

- It passes the user's node ID and **all** of the user's workspace and team node IDs, resolved from Postgres, so a shared app is visible from whichever workspace the user is in.
- The app-deploy Lambda returns `hasAccess`, and the authorizer caches it with the usual 300 s TTL.
- This is the only cross-service call the events authorizer makes.

**Runs.** Run status comes from each compute node's `workflow-status-updater`, which calls workflow-service's `PUT …/runs/{id}/status` and processor-status endpoints with the run's callback token. workflow-service updates DynamoDB and then publishes. Today that publish goes to one channel per workspace, and every run page filters for its run. Instead:

- **Publish once, per run.** workflow-service publishes to `/runs/org-<orgUuid>/<runUuid>`, or `/runs/user-<createdByUuid>/<runUuid>` when there's no workspace. It already has `run.OrganizationId` and `run.CreatedBy` in the status handlers.
- **Subscribe narrowly or widely.** A run detail page, RunMonitor or the agent watching a run subscribes to that one run. The runs and notebooks overviews subscribe to `/runs/org-<orgUuid>/*` and get every run in the workspace.
- **No cross-service call.** The workspace or user is in the path, so the authorizer checks membership (or identity) without asking workflow-service about the run.
- **Payloads are unchanged:** `workflow-run-status` `{runId, organizationId, status, timestamp}` and `workflow-processor-status` (plus `nodeId`). IDs and status only.

### 2. Channels and payloads

Namespaces are `datasets`, `runs` and `applications`. Channels use bare UUIDs, plus an `org-`/`user-` prefix on a run's scope segment. A segment is at most 50 characters (`org-` + a UUID is 40) and a channel at most 5 segments.

| Today (Pusher) | Publisher | Events | AppSync channel |
|---|---|---|---|
| `dataset-<uuid>` | upload-service-v2 | `upload-event` | `/datasets/<uuid>` |
| `organization-<uuid>-analytics` | workflow-service | `workflow-run-status`, `workflow-processor-status` | `/runs/org-<orgUuid>/<runUuid>` |
| `user-<uuid>-analytics` | workflow-service (runs with no workspace) | same | `/runs/user-<userUuid>/<runUuid>` |
| `application-<uuid>` | app-deploy-service | `application_status_event` | `/applications/<appUuid>` |

Every event is one JSON envelope, so subscribers keep dispatching by name as they do with Pusher's `bind`:

```json
{ "event": "upload-event", "data": [ { "name": "…", "node_id": "N:package:…", "upload_id": "…" } ] }
```

- **Payloads may keep names.** They stay in AWS under the BAA and only reach authorized subscribers. Moving to IDs only is a later, optional hardening.
- **Event API logging at `ERROR`,** never `ALL`/`DEBUG`: those levels log payloads, which would copy file names into CloudWatch. Logs are KMS-encrypted (clin: accelerator key, 365 days).

### 3. Publishing: `pkg/realtime` in go-core

A small package replaces direct Pusher use:

```go
type Channel struct{ Namespace string; Segments []string }

func Dataset(datasetUuid string) Channel
func Run(orgUuid, createdByUuid, runUuid string) Channel // org scope when orgUuid is set, else user scope
func RunScope(orgUuid, createdByUuid string) Channel      // the /runs/<scope>/* wildcard, for subscribers
func Application(appUuid string) Channel

type Publisher interface {
    Publish(ctx context.Context, ch Channel, event string, data any) error
    PublishBatch(ctx context.Context, events []Event) error // split into requests of ≤5 events
}
```

- **Implementations:**
  - `appsync`: SigV4-signed `POST https://<http-endpoint>/event` (service `appsync`), with the Lambda's own role (`appsync:EventPublish` on the API's namespaces).
  - `noop`: live updates off.
- **Selection:** `REALTIME_EVENTS_ENDPOINT` set means `appsync`; unset means `noop`. There is no Pusher adapter: each environment cuts over (see Rollout). This replaces clin's interim disabled `/ops/pusher-config`.
- **Failures:** publish failures return an error; callers keep today's behaviour of logging and continuing. Nobody's upload fails because an event didn't go out.
- **`domain.PusherAPI` stays** for services that haven't moved yet, then goes (subtask 8).

### 4. Clients

- **Web app:** one `realtime` module (subscribe, unsubscribe, reconnect, re-subscribe, dispatch by `event`) over the Amplify Events client in Lambda-auth mode with the user's access token. The Amplify library is already used for Cognito. It replaces the ~10 `pusher.subscribe` sites and the global `$pusher`. Run pages move from the workspace channel to `runs`: overviews use `/runs/org-<orgUuid>/*`, and detail pages use the single run. The site config gets a `realtime` block, and each environment's CSP gets the Event API's HTTP and realtime hosts.
- **Agent:** a Go client for the Event API WebSocket protocol: subprotocol `aws-appsync-event-ws` plus the base64url auth header, `connection_init`, `subscribe`, `ka`, reconnect with backoff and re-subscribe. It uses the agent's token-pool access token. To follow a run it started, it subscribes to that run's channel. The endpoint comes from the profile, like `api_host`/`api2_host`.

## Limits and cost

From AWS's published quotas and pricing:

| Limit | Value | Our use |
|---|---|---|
| Events per publish request | 5 (fixed) | `PublishBatch` splits |
| Publish payload | 1.2 MB | fine |
| Delivered message | 240 KB | an upload batch's rows fit |
| Channel segments / segment length | 5 / 50 chars | ≤3 segments, UUIDs |
| Subscriptions per connection | 200 (adjustable) | one connection per tab/agent, a few channels |
| Inbound events / connects | 10,000/s, 2,000/s (adjustable) | far below |

**Pricing** is $1.00 per million operations (publishes, deliveries per 5 KB, connects, subscribes, pings) plus $0.08 per million connection-minutes. At today's volume (a handful of upload and run events per active user per day), that's an estimated few cents to dollars per month per environment. Re-check against CloudWatch metrics after the dev soak.

## Terraform

- The AWS provider supports `aws_appsync_api` (with `event_config`: `auth_provider`, `connection_auth_mode`, default publish/subscribe modes, `log_config`) and `aws_appsync_channel_namespace` (per-namespace auth modes, handlers). Verified against **provider 6.66.0**, the version in clin's plugin cache.
- **Module:** `terraform-modules/appsync-events` (vendored into clin-infrastructure following its conventions: `aws_account` names, KMS accelerator key, vendored `service-alarms`, no module-level provider blocks). Outputs: API id, HTTP and realtime endpoints, and a publisher IAM policy ARN.
- **The authorizer's Lambda** is created by pennsieve-go-api's Terraform, like the websocket authorizer. It needs an `appsync.amazonaws.com` invoke permission scoped to the API.

## Rollout

Clin first (Pusher is off there today), then dev, then prod. Each environment cuts over rather than running both:

1. Event API (Terraform) and the events authorizer.
2. Web app subscribing to AppSync.
3. The three publishers (upload-service-v2, workflow-service, app-deploy-service) on `pkg/realtime`.

Live updates are best-effort, so in dev and prod they are missing between steps 2 and 3; a page refresh still shows current state. If that gap is unacceptable for prod, the web app can listen to both Pusher and AppSync for one release. That temporary code stays in the frontend. See subtasks 2–9 for the per-step plan.

## Open items

- **Penn SRE** confirmed (2026-09-28) that the AppSync Event API is included in Penn's BAA. Still to confirm: whether WAF is required on the Event API in the landing zone.
- **Run events for workspace members vs the run creator:** an org-scoped run is visible to every member of its workspace, as today. Narrowing it to the creator later only changes the authorizer rule, not the channels.
