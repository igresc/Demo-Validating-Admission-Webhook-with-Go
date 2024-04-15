# Demo Validating Admission Webhook created with Go

Kubernetes documentation about admission webhooks:
https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers

### Generate CA and sign a cert:
```
openssl genrsa -out certs/tls.key 2048
openssl req -new -key certs/tls.key -out certs/tls.csr -subj "/CN=demo-webhook-svc.default.svc"
openssl x509 -req -extfile <(printf "subjectAltName=DNS:demo-webhook-svc.default.svc") -in certs/tls.csr -signkey certs/tls.key -out certs/tls.crt
```

```
kubectl create secret tls demo-webhook-tls \
    --cert "certs/tls.crt" \
    --key "certs/tls.key" -n default 
```

```
kubectl get secret demo-webhook-tls -n default -oyaml > k8s/tls-secret.yaml
```

```
ENCODED_CA=$(cat certs/tls.crt | base64 | tr -d '\n')
```


#### ValidatingWebhookConfiguration:
The service default path is `/`, our server listens to `/validate` on port `443`
https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#service-reference

### How to run this demo

1. Docker and Go installed (optionally tmux and k9s for `task demo`)
2. Install [taskfile](https://taskfile.dev/installation/)
3. Generate the CA and certs as explained in the first section or use CertManager
4. Run `task demo` 
5. Run `task k8s-up`
6. Try to deploy the example pods `allowed-pod.yaml` and `unallowed-pod.yaml`
7. See the messages you recieve when trying to apply the unallowed-pod yaml
8. In the `admission-go` pod you can see the logs that show what as sent as a request and the reason for the aproval/rejection


### Possible Improvements
- [ ] verify the resource is a pod
- [ ] check all containers in the pod not only the first (main.go#L72)
- [ ] Check for a minum resource value
- [ ] [CertManager CA injector](https://cert-manager.io/docs/concepts/ca-injector/) as an automated alternative of generating the certs
