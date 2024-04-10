# Demo Validating Admission Webhook created with Go

Kubernetes documentation about admission webhooks:
https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers

Generate CA and sign a cert:
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
ENCODED_CA=$(cat certs/tls.crt | base64 | tr -d '\n')
sed -e 's@${ENCODED_CA}@'"$ENCODED_CA"'@g' <"k8s/validatingWebhook.yaml" | kubectl create -f -
```

CertManager CA injector as an automated alternative:
https://cert-manager.io/docs/concepts/ca-injector/

#### ValidatingWebhookConfiguration:
The service default path is `/`, our server listens to `/validate` on port `443`
https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#service-reference