# secretsyncer

## Quickstart

### Add helm repo:

```bash
helm repo add secretsyncer https://iamsmartad.github.io/secretsyncer
helm repo update
```

### Inspect chart:

```bash
# show all version
helm search repo secretsyncer --versions

# show created resources
helm template mysecretsyncer secretsyncer/secretsyncer
```

### Deploy `local-only` secretsyncer :

```bash
kubectl create namespace truth
helm -n truth upgrade --install secretsyncer secretsyncer/secretsyncer
```

### Create a source secret `my-truth.yaml`:

```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: my-truth
  namespace: truth
stringData:
  name: "John Doe"
  password: "eW91IGFyZSB2ZXJ5IGdvb2QgYXQgZGVjb2RpbmcgYmFzZTY0"
```

```bash
# and apply it
kubectl -n truth apply -f my-truth.yaml
```

### Create a placeholder secret `my-secret.yaml`:

```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: my-secret
  namespace: default
  annotations:
    iamstudent.dev/syncResourceVersion: "0"
    iamstudent.dev/syncSourceName: my-truth
    iamstudent.dev/syncSourceNamespace: truth
  labels:
    iamstudent.dev/sync: "receiver"
data: {}
```

```bash
# and apply it
kubectl apply -f my-secret.yaml

# wait a few seconds
kubectl get secret my-secret -o yaml
```

### Result

The newly created secret `my-secret` in namespace `default` has the same values as the template secret in namespace `truth`:

```yaml
apiVersion: v1
kind: Secret
data:
  name: Sm9obiBEb2U=
  password: ZVc5MUlHRnlaU0IyWlhKNUlHZHZiMlFnWVhRZ1pHVmpiMlJwYm1jZ1ltRnpaVFkw
metadata:
  name: my-secret
  namespace: default
  annotations:
    field.cattle.io/description: synced Sep 30 17:15:34 from truth/my-truth
    iamstudent.dev/syncResourceVersion: "43775"
    iamstudent.dev/syncSourceName: my-truth
    iamstudent.dev/syncSourceNamespace: truth
[...]
```
