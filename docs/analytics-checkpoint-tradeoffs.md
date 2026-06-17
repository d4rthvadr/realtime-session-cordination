# Analytics Processor: Checkpoint & Scaling Tradeoffs

## Current Implementation (Single-Node)

The analytics processor uses **checkpoint-based recovery** for idempotent processing:

- **SaveCheckpoint()**: Records the last successfully processed event ID per worker
- **LoadCheckpoint()**: Retrieves the last known position to resume from
- **Per-Worker Tracking**: Each worker (identified by `WorkerName`) maintains its own checkpoint row

### Design Decisions for v1

**Tradeoff 1: No Startup Recovery (Intentional for Single-Node)**

- ✓ Simpler: Processor starts and begins claiming from available pending rows
- ✗ Reprocessing: If processor crashes and restarts, events marked as `processed` but not checkpointed may be reprocessed
- **Rationale**: On single-node deployment, reprocessing is safe (idempotent) and simpler than coordinating recovery. Checkpoints are saved _before_ MarkProcessed, so at-least-once semantics are preserved.

**Tradeoff 2: Single Worker Instance**

- ✓ Simple: No distributed coordination needed; single worker with fixed WorkerName
- ✗ Scaling: Multiple processor instances would need unique worker names to avoid duplicate processing
- **Rationale**: Current deployment is single-node single-instance. If scaling is needed later, workers can be assigned unique names (e.g., via hostname or environment variable) and explicit coordination added.

**Tradeoff 3: Lease-Based Concurrency Without Distributed Lock**

- ✓ Works: Lease timestamps (`leased_until`) prevent concurrent processing of same row
- ✓ Works: Expired leases are auto-claimable after 15 seconds if processor crashes
- ✗ Doesn't handle: Multiple processor instances without name coordination
- **Rationale**: Sufficient for single-instance case. Cross-instance coordination (Redis lock, etcd, etc.) deferred to multi-node deployment.

## Future: Multi-Node Scaling

When extracting the processor to a separate service or scaling horizontally:

### Required Changes

1. **Checkpoint Recovery on Startup**
   - Call `LoadCheckpoint(workerName)` in `Processor.Start()` to resume from last known event
   - Skip claiming rows before checkpoint
   - Avoids reprocessing on restarts

2. **Unique Worker Names Per Instance**
   - Assign via hostname: `WorkerName = os.Hostname()`
   - Or environment variable: `WorkerName = os.Getenv("PROCESSOR_WORKER_ID")`
   - Prevents checkpoint collisions

3. **Distributed Coordination (Optional)**
   - If workers should coordinate (e.g., leader election, balanced claiming):
     - Use Redis/etcd for lock on checkpoint updates
     - Or implement gossip protocol for lease expiry awareness
   - **Not needed** if workers are simply independent instances with unique names

## Checkpoint Table Schema

```sql
CREATE TABLE analytics_checkpoints (
  worker_name TEXT PRIMARY KEY,
  last_event_id TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (last_event_id) REFERENCES analytics_events(id)
);
```

- One row per worker
- PK ensures at most one checkpoint per worker name
- `last_event_id` tracks progress for idempotency
- `updated_at` for operational visibility

## Atomicity Guarantees

- **SaveCheckpoint()**: Uses SQLite `ON CONFLICT` upsert—atomic
- **Lease Expiry**: Timestamp comparison in `ClaimPendingForProcessing()`—atomic
- **State Transitions**: `MarkProcessed()` and `MarkFailed()` are atomic updates
- **At-Least-Once Semantics**: Checkpoint saved before MarkProcessed ensures no event is silently dropped

## When to Upgrade

| Condition                                      | Action                                                |
| ---------------------------------------------- | ----------------------------------------------------- |
| Single processor instance, single node         | Current implementation is sufficient                  |
| Multiple processor instances competing         | Add unique worker names and optional startup recovery |
| Distributed processors with coordination needs | Add Redis/etcd lock or gossip protocol                |
| Exactly-once processing required               | Add idempotency table keying on (worker, event_id)    |

## Monitoring

Current observability via `GetFreshness()`:

- `PendingCount`: Outstanding events in outbox
- `OldestPendingAt`: Lag indicator (now - oldest pending time)
- `LastProcessedAt`: Last checkpoint update time

Alerts to set up:

- Pending count > threshold (backlog building)
- Lag > threshold (processor stalled)
- LastProcessedAt stale (processor hung)
