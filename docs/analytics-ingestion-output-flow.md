# Analytics Ingestion To Output Flow

This document shows the end-to-end analytics flow from event ingestion to API output.

## Sequence Diagram

```mermaid
sequenceDiagram
    autonumber
    participant W as Write API Handler
    participant E as Analytics Emitter
    participant S as Analytics SQLite Store
    participant O as analytics_outbox
    participant P as Processor (ticker loop)
    participant B as ProjectionBuilder
    participant PR as Projection Tables
    participant C as Checkpoint Table
    participant R as Read API Handler
    participant M as Analytics Manager
    participant D as Session + ProgramItem Stores
    participant U as Dashboard Client

    W->>E: Emit(event)
    E->>S: AppendEventAndEnqueue(event, now)
    S->>S: INSERT analytics_events
    S->>O: INSERT outbox row (state=pending)

    loop every poll interval
        P->>S: ClaimPendingForProcessing(worker, leaseUntil, batch)
        S->>O: UPDATE row -> processing (lease_owner, leased_until, attempt+1)
        P->>S: GetEvent(event_id)

        alt projection rebuild succeeds
            P->>D: Get session snapshot (event.session_id)
            P->>D: List program items (event.session_id)
            P->>B: RebuildSessionProjection(...)
            B->>PR: UpsertSessionProjection

            P->>D: List all sessions
            loop for each session
                P->>D: List program items (session_id)
            end
            P->>B: RebuildPlatformProjection(...)
            B->>PR: UpsertPlatformProjection

            P->>S: SaveCheckpoint(worker, last_event_id)
            S->>C: UPSERT checkpoint row
            P->>S: MarkProcessed(outbox_id)
            S->>O: UPDATE row -> processed
        else rebuild/checkpoint fails
            P->>S: MarkFailed(outbox_id, error, deadLetter?)
            S->>O: UPDATE row -> pending or dead_letter
        end
    end

    U->>R: GET analytics endpoint
    Note over R,PR: Step 5 target path: read projection first, fallback to manager

    alt projection-first read (target)
        R->>PR: GetSessionProjection / GetPlatformProjection
        PR-->>R: projection found
        R-->>U: precomputed analytics + freshness
    else fallback (current behavior)
        R->>D: load snapshots
        R->>M: BuildSessionSummary / BuildPlatformOverview
        M-->>R: computed summary
        R->>S: GetFreshness(worker)
        R-->>U: computed analytics + freshness
    end
```

## Activity Diagram

```mermaid
flowchart TD
    A[Domain mutation handled] --> B[Emit analytics event]
    B --> C[Append event and enqueue outbox row]
    C --> D{Processor poll tick}
    D -->|No pending rows| D
    D -->|Pending rows claimed| E[Load event payload]
    E --> F[Rebuild session projection]
    F --> G[Rebuild platform projection]
    G --> H{Projection rebuild ok?}
    H -->|No| I[Mark outbox failed]
    I --> J{Attempts exceeded?}
    J -->|No| K[Return to pending]
    J -->|Yes| L[Move to dead letter]
    K --> D
    L --> D

    H -->|Yes| M[Save checkpoint]
    M --> N{Checkpoint save ok?}
    N -->|No| I
    N -->|Yes| O[Mark outbox processed]
    O --> D

    P[Client requests analytics] --> Q{Projection read enabled in handler?}
    Q -->|Yes and hit| R[Return projection summary + freshness]
    Q -->|No or miss| S[Compute summary on demand from snapshots]
    S --> T[Return computed summary + freshness]
```

## Notes

- Outbox leasing prevents duplicate simultaneous processing in a single-node deployment model.
- Checkpoint tracks worker progress by last processed event id.
- Dead-letter rows represent events that exceeded retry attempts.
- Projection rebuild currently happens before checkpoint save and outbox processed state transition.
