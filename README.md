# Key/Value Server

This project implements a single-machine key/value server with a
distributed lock that works over an unreliable network. This is
Lab 2 from MIT's 6.5840 distributed systems course.

## Project Overview

The lab consists of four tasks that build upon each other, each introducing
fundamental distributed systems concepts:

1. Basic Put/Get with versioning (reliable network)
2. Distributed lock implementation on top of KV store
3. Handle unreliable networks with retry logic
4. Make locks work with unreliable networks

## Task 1: Basic Put/Get with Versioning (Reliable Network)

### What Was Built

- A KV server with conditional updates using version numbers
- Each key has a value and a version number
- `Get(key)` returns value, version, and error
- `Put(key, value, version)` only succeeds if current version matches
- If version mismatches → `ErrVersion`

### Key Concepts

- **Optimistic concurrency control**: Version numbers enable conditional updates
- **Compare-and-swap semantics**: Put only succeeds if version hasn't changed
- This prevents lost updates in concurrent scenarios

## Task 2: Distributed Lock on KV Store

### What Was Built

- A `Lock` type with `Acquire()` and `Release()` methods
- Lock state stored in the KV store using:
  - Empty string `""` = unlocked
  - `clientID` = locked by that client
- Used version numbers to detect races when multiple clients try to acquire

### Key Concepts

- **Building abstractions**: Higher-level primitives (locks) built on
lower-level ones (KV store)
- **Lock ownership**: Each client has a unique ID to claim locks
- **Race handling**: When two clients try to acquire simultaneously, version
numbers ensure only one succeeds

### Implementation Details

- `Acquire()`: Loop until you successfully claim the lock (Put your clientID
when value is "")
- On `ErrVersion`: someone else got there first, sleep and retry
- `Release()`: Put empty string if you own the lock

## Task 3: Handle Unreliable Networks with Retry Logic

### The Challenge

Networks can re-order, delay, or discard RPC requests/replies. Clients must
retry when they don't receive replies, but this creates ambiguity.

### Key Concepts Learned

#### 1. At-Most-Once Semantics

Operations execute at most one time, even with retries.

#### 2. Idempotency for Get

- Get doesn't modify state -> safe to retry indefinitely
- Implementation: Simple loop with `!ok` check and sleep

#### 3. Conditional Execution for Put

- Put with version number is naturally idempotent
- If Put succeeded on first attempt, retry will get `ErrVersion`
- Server won't execute the same Put twice

#### 4. Critical Distinction: ErrVersion vs ErrMaybe

When a client retries a Put:

**Scenario 1: First attempt succeeds, reply dropped**

```
Client → Put(key, "foo", v=5)
Server: Executes, version now 6, sends OK
Network: Drops reply
Client: No reply, retries Put(key, "foo", v=5)
Server: Version mismatch, sends ErrVersion
Client: Sees ErrVersion on retry → ???
```

**Scenario 2: First attempt never arrives**

```
Network: Drops request
Client: No reply, retries Put(key, "foo", v=5)
Server: Executes, sends OK
Client: Sees OK on retry → success
```

**The ambiguity:** When you get `ErrVersion` on a retry, you don't know which
scenario occurred!

- Maybe your first attempt succeeded (scenario 1)
- Maybe someone else updated the key before your request arrived

**Solution:** Return `ErrMaybe` to the application when:

- You retried the RPC (retry flag = true)
- AND you got `ErrVersion` back

#### 5. Why OK on Retry is Unambiguous

If a retry succeeds with OK:

- The version matched
- Which means the first attempt **didn't** execute
- Otherwise version would have been incremented
- So OK on retry = definitive success, exactly once

### Implementation (kvsrv/client.go)

```go
retry := false
for {
    ok := ck.clnt.Call(...)
    if !ok {
        retry = true  // Track that we're retrying
        time.Sleep(100 * time.Millisecond)
        continue
    }
    if retry && reply.Err == rpc.ErrVersion {
        return rpc.ErrMaybe  // Ambiguous!
    }
    return reply.Err  // Definitive answer
}
```

## Task 4: Make Locks Work with Unreliable Networks

### The Challenge

Now that `Clerk.Put()` can return `ErrMaybe`, the lock implementation must
handle this ambiguity.

**The problem:**

```go
err := lk.ck.Put(lk.name, lk.clientID, ver)
if err == rpc.OK { return }        // Got it
if err == rpc.ErrVersion { retry } // Didn't get it
if err == rpc.ErrMaybe { ??? }     // Don't know
```

### The Key Insight

When you don't know if your operation succeeded, **query the current state** to
find out!

### Implementation (lock/lock.go)

```go
if err == rpc.ErrMaybe {
    val, _, _ := lk.ck.Get(lk.name)
    if val == lk.clientID {
        return  // We got the lock!
    } else {
        time.Sleep(10 * time.Millisecond)
        continue  // didn't get it, retry
    }
}
```

**Why this works:**

- Get queries the **current truth** at the server
- If lock value is your clientID -> you acquired it (maybe on first attempt,
maybe on retry, doesn't matter)
- If lock value is something else -> you didn't get it, safe to retry

**Two places needing ErrMaybe handling:**

1. Creating the lock when `ErrNoKey`
2. Claiming an unlocked lock when `val == ""`

Both use the same pattern: ErrMaybe -> Get -> check result -> return or retry

## Concepts Learned

1. **Unreliable networks are the norm**: Messages can be lost, delayed, or reordered
2. **At-most-once vs exactly-once**: Building exactly-once is hard;
at-most-once with clear failure modes is more practical
3. **Idempotency**: Operations that can be safely retried
4. **Conditional operations**: Version numbers enable safe concurrent updates
5. **Handling ambiguity**: When you don't know the outcome, query the state
6. **Building abstractions**: Locks built on KV store, showing how distributed
systems compose

### Engineering Patterns

1. **Retry loops**: Keep trying until you get a definitive answer
2. **State tracking**: Using flags (like `retry`) to track request history
3. **Error semantics**: Different errors convey different information (OK, ErrVersion, ErrMaybe)
4. **Testing concurrent systems**: Using race detector and multiple clients

### The Core Trade-off

Perfect clarity is impossible in distributed systems with unreliable networks.
This system provides:

- **Definitive success**: When Put returns OK
- **Definitive failure**: When Put returns ErrVersion on first attempt
- **Admitted uncertainty**: When Put returns ErrMaybe on retry

Applications must handle `ErrMaybe`, but this is better than silently doing the wrong thing!

## Running Tests

```bash
# Test basic KV server
cd kvsrv/
go test -v

# Test locks (with race detector)
cd kvsrv/
go test -v -race
```

## Reference

Lab specification: <https://pdos.csail.mit.edu/6.824/labs/lab-kvsrv1.html>
