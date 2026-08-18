# Project: Homelab CI/CD Orchestrator (Go)

A CLI tool, written in Go, that bootstraps/updates an entire homelab infrastructure with a single command. It reads a small central YAML file that declares execution order and dependencies between self-contained "tool" directories (terraform, ansible, docker, etc.), builds a dependency graph, executes tools in the correct order (parallel where possible), passes runtime output from one tool into the next, and rolls back to the pre-run state on failure.

```
homelab/
    pipeline.yml              # central file: order + dependencies ONLY
    terraform/
      main.tf
      ...
    ansible/
      playbook.yml
      ...
    docker/
      compose.yml
      ...
```

Design principle: \*\*pipeline.yml owns ordering. Each tool directory is portable — it can be copied to another pipeline and it still describes itself.

## 2. pipeline.yml schema

```yaml
steps:
  - name: terraform-vpc
    tool: terraform
    dir: tools/terraform-vpc
    depends_on: []
  - name: ansible-base
    tool: ansible
    dir: tools/ansible-base
    depends_on: [terraform-vpc]
  - name: docker-services
    tool: docker-compose
    dir: tools/docker-services
    depends_on: [ansible-base]
```

the name must be in the format tool-name

## 4. Execution model

### Config reading → graph → stages

1. Parse `pipeline.yml` → build graph nodes/edges from `depends_on`.
2. Topologically sort the graph → produces `[][]Runner` (aka stages):
   - **Outer slice** = sequential stages (must run in order).
   - **Inner slice** = runners within a stage that have no dependency on each other → safe to run **in parallel**.
3. For each `dir` referenced by a step, attach the concrete `Runner` implementation for that type. type is set by the dir name.

### CLI / preflight

- Parse flags and inputs.
- Validate that config files passed to the CLI exist and are well-formed (pipeline.yml and every dir is in a valid state)
- Check system binaries: verify `terraform`, `ansible-playbook`, `docker` exist in `$PATH` before running anything.
- Structured logging / terminal feedback mechanism active from the start (stage/step/status level logs).

### Runner interface

```go
type Runner interface {
    Validate() error
    Snapshot() error                          // capture current state before run
    Run(ctx context.Context, state *ExecutionState) error
    Rollback() error                          // restore snapshot on failure
}
```

One concrete implementation per `type` (terraform, ansible, docker, ...).

### Shared execution state

```go
type ExecutionState struct {
    mu      sync.Mutex
    outputs map[string]map[string]string // step name -> output key -> value
}
```

- Thread-safe, passed to every runner.
- Holds runtime metadata one step produces that a later step needs (e.g. terraform outputs, generated temp file paths).

### Parallelism

- Within a stage: runners execute concurrently via goroutines + `errgroup`, all sharing the same `*ExecutionState` (mutex-guarded).
- Across stages: strictly sequential — a stage only starts once the previous stage fully succeeds.

### Failure & rollback

- Before stage 1 begins, every runner's `Snapshot()` is called (e.g. copy terraform state file, save docker compose state) into `.backup/<run-id>/`.
- after each successfull run the current state is stored in a success-backup for rollback to previous states
- Rollback target = **the infra state immediately before this run started** — even if this run was an update to existing infra, not a fresh bootstrap.
- On any runner error:
  1. Cancel context → stops in-flight goroutines in the current stage.
  2. Call `Rollback()` on every runner that already completed successfully, in reverse order.
  3. Ansible steps are generally idempotent-safe and may not have a true "undo" — rollback for ansible-type runners means re-applying last-known-good state rather than a literal revert.

---

## 5. Go project skeleton

```
cmd/homelab/main.go       # CLI entry point, flag parsing
internal/config/          # pipeline.yml + tool.yml parsing & validation
internal/graph/           # dependency graph construction + topo sort
internal/runner/          # Runner interface + terraform/ansible/docker implementations and ExecutionStatedd
internal/orchestrator/    # stage execution, parallel dispatch, rollback logic
```

---

## 6. Build checklist (source of truth, mapped to design above)

- [x] Parse CLI flags and inputs
- [x] Validate config files passed to CLI (pipeline.yml + tool.yml files) are well-formed
- [ ] Check system binaries: terraform, ansible-playbook, docker in `$PATH`
- [ ] Structured logging / terminal feedback mechanism
- [ ] Config reader: pipeline.yml → dependency graph → `[][]Runner`
- [ ] Runner interface definition
- [ ] Runner implementations per tool type (terraform, ansible, docker, ...)
- [ ] Method: config → dependency graph
- [ ] Method: dependency graph → attach runner per step
- [ ] `ExecutionState` struct (thread-safe, shared across runners)
- [ ] Dynamic state injection (tool.yml `inputs[].from` → ExecutionState lookup → runner config)
- [ ] Orchestrator: execute all stages/runners (parallel within stage, sequential across stages)
- [ ] Snapshot mechanism (pre-run state capture per runner)
- [ ] Fail/rollback logic: restore infra to pre-run state on any failure

---

## Open items / not yet decided

- Exact snapshot mechanism per runner type (terraform state copy is straightforward; docker/ansible need more thought).
- Exact structured logging format/library choice.
- CLI flag surface (what exactly is configurable at invocation time vs. in pipeline.yml).
