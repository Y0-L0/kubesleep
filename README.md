# kubesleep

Save your development cluster one pod at a time.

*kubesleep* is a lightweight CLI utility that can **suspend** (scale to zero) and **wake** Kubernetes namespaces, saving cluster capacity by shutting down idle development and testing environments and restoring them on demand.

---

## 📦 Installation

**Homebrew (macOS/Linux):**
```bash
brew install y0-l0/tap/kubesleep
```

**APT (Debian/Ubuntu):**
```bash
echo "deb [trusted=yes] https://apt.fury.io/stadimeter/ /" \
  | sudo tee /etc/apt/sources.list.d/fury.list
sudo apt update && sudo apt install kubesleep
```

**Binary (Linux amd64):**
```bash
VERSION=0.5.3
curl -fsSL "https://github.com/Y0-L0/kubesleep/releases/download/v${VERSION}/kubesleep_${VERSION}_linux_amd64.tar.gz" | tar -xz
sudo mv kubesleep /usr/local/bin/
```

Binaries for other platforms and architectures are available on the [latest release page](https://github.com/Y0-L0/kubesleep/releases/latest).

**Build from source:**
```bash
go install github.com/Y0-L0/kubesleep/cmd/kubesleep@latest
```

> **Prerequisites**
>
> * A kubeconfig whose user or service‑account has RBAC rights to interact with namespaces, ConfigMaps, Deployments, StatefulSets and CronJobs.

---

## Quick Start

```bash
# Suspend the dev namespace
kubesleep suspend --namespace dev

# Wake the namespace back up
kubesleep wake --namespace dev
```

Add `-v` or `-vv` to any command for info/debug logs.

## Common use‑cases

### Wake

Wake up one or more previously suspended namespaces by restoring the replica counts recorded during the last successful `suspend` operation.

```bash
kubesleep wake -n dev
kubesleep wake -n dev -n staging -vv
```

You can also wake a namespace by redeploying your workloads to it (e.g., with `helm upgrade --install`). If you choose this option, delete the `kubesleep‑suspend‑state` ConfigMap manually:

```bash
kubectl -n <your-namespace> delete configmap kubesleep-suspend-state
```

### Suspend

Suspend one or more namespaces by scaling workloads to zero and persisting their original replica counts in a state ConfigMap.

```bash
kubesleep suspend -n dev
kubesleep suspend -n dev -n staging
```

The operation is mostly idempotent and can be rerun to update the suspend state or to repeat a failed or aborted attempt.

See below for details about the suspend‑state merge behaviour.

#### Protected Namespaces

Certain Kubernetes namespaces are protected from accidental suspension: `default`, `kube-{system,public,node-lease}`, `ingress-nginx`, `istio`, `local-path`.

These namespaces are skipped unless you explicitly override the protection with the `--force` flag:

```bash
kubesleep suspend -n ingress-nginx --force
```

You can protect additional namespaces by annotating the namespace manifest with `kubesleep.xyz/do-not-suspend="true"`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    kubesleep.xyz/do-not-suspend: "true"
```

#### Periodic auto‑suspension

Humans are forgetful. For development and testing clusters you may want to schedule an automatic suspension of all unprotected namespaces:

```bash
kubesleep suspend --all-namespaces
```

The `--all-namespaces` flag cannot be combined with `--force`.

## Merge semantics

The `kubesleep suspend` command can be repeated to:

* Resume a previously cancelled suspend operation.
* Retry a failed suspend operation.
* Redo a suspend operation after adding or modifying workloads with tools like `kubectl` or `helm`.

New workloads are added to the suspend state. Manual changes to `replicaCount` are **not** incorporated. The `wake` command always restores the initial replica counts, not any intermediate values. After a successful `wake`, the suspend state is reset (the ConfigMap is deleted).

**Example**

1. You suspend a workload, scaling it from 2 replicas to 0.
2. You manually scale it to 5 replicas with `kubectl`.
3. You suspend the namespace again.
4. You wake the namespace.

The workload is restored to 2 replicas (the original value), **not** to 5.

## Limitations

Running multiple concurrent suspend or wake operations on the same namespace can lead to undefined behavior and is not supported.

---

## 💻 Development

```bash
# Run unit tests
go test ./... -run 'TestUnit'

# Run unit and integration tests
KUBEBUILDER_ASSETS=$(setup-envtest use -p path) \
  go test ./... --coverprofile=coverage.out -run '' && \
  go tool cover -html=coverage.out -o coverage.html

# Build the CLI binary
go build -o bin/kubesleep ./cmd/kubesleep
```

---

## 🤝 Contributing

Feature requests, bug reports, and pull requests are welcome!

---

## 📄 License

Distributed under the [AGPL‑v3 License](LICENSE).
