
## Merge semantics:

The `kubesleep suspend` operation can be repeared to handle e.g.:

- Resume a previously cancelled suspend operation.
- Retry a failed suspend operation.
- Redo a suspend operation after adding or modifying the namespace
  with e.g. `kubectl` or `helm`.

New workloads will be added to the suspend state.
Manually changing the replicaCount of a workload won't be integrated into the suspend state.
The `wake` operation will always restore initial state of a workload and not an intermediate state.
A successful `wake` operation will reset the suspend state. (the statefile ConfigMap is deleted)

An example:

You suspend scales a workload down from 2 to 0
You manually  scale it back up to 5 using e.g. `kubectl edit`
You suspend the namespace again.
You wake the namespace back up.
The workload will get scaled to the initial value up to 2, not to the intermediate value of 5


## Limitations

Executing multiple suspend or wake operations on the same namespace
can lead to unexpected errors and is not supported.
In the future this may become explicitly prevented.
