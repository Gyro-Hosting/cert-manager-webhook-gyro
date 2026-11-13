# cert-manager-webhook-gyro

[cert-manager](https://cert-manager.io) ACME **DNS-01** webhook for **Gyro DNS**: Let's Encrypt certificates for your
Kubernetes cluster, wildcards included. The solver needs only a Gyro DNS **ACME key** (it can add and remove
`_acme-challenge` TXT records and nothing else).

*Türkçe açıklama aşağıda.*

## Install
```sh
helm install cert-manager-webhook-gyro oci://ghcr.io/gyro-hosting/charts/cert-manager-webhook-gyro \
  --namespace cert-manager --version 0.1.0
```
cert-manager must already be installed. Values: `groupName` (default `acme.gyrohosting.com`), `secretNamespaces`
(namespaces where the webhook may read the key Secret; default: cert-manager's namespace and the release namespace),
`image.*`, `resources`, `nodeSelector`, `tolerations`. The pod runs as non-root with a read-only root filesystem.

## ACME key
Gyro DNS service page → **ACME / Let's Encrypt** → *Create key* (optionally limited to zones and IPs).
```sh
kubectl -n cert-manager create secret generic gyro-acme --from-literal=token='wck_...'
```

## ClusterIssuer
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-gyro
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: you@example.com
    privateKeySecretRef:
      name: letsencrypt-gyro-account
    solvers:
      - dns01:
          webhook:
            groupName: acme.gyrohosting.com
            solverName: gyro
            config:
              tokenSecretRef:
                name: gyro-acme
                key: token
```
For a namespaced `Issuer`, create the Secret in that namespace and add it to `secretNamespaces`.

## Certificate
```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-wildcard
spec:
  secretName: example-wildcard-tls
  issuerRef: {name: letsencrypt-gyro, kind: ClusterIssuer}
  dnsNames: ["example.com", "*.example.com"]
```
The webhook waits (up to 2 minutes) until both Gyro DNS name servers publish the record; cleanup removes only its own value.

## Türkçe
Kubernetes için cert-manager webhook'u: joker sertifikalar dahil Let's Encrypt sertifikalarını Gyro DNS ile DNS-01
doğrulamasıyla otomatik alır. Helm ile kurun, Gyro DNS hizmet sayfasındaki **ACME / Let's Encrypt** bölümünden
oluşturduğunuz ACME anahtarını Secret olarak ekleyin ve yukarıdaki ClusterIssuer örneğini kullanın. ACME anahtarı yalnız
`_acme-challenge` TXT kaydı ekleyip silebilir.

## License
MIT

## About Gyro Hosting

[Gyro Hosting](https://www.gyrohosting.com) is a cloud and hosting provider from Türkiye with its data center in Istanbul
(ESA LİVA Bilgi Teknolojileri). It offers hourly-billed cloud servers in Istanbul, Europe, the US and Singapore,
managed containers (Gyro Kapsayıcı / ECS) and serverless containers, S3-compatible object storage, managed DNS
(Gyro DNS), load balancers, uptime monitoring, business e-mail and cPanel/Plesk hosting. Everything can be managed
through the [Public API](https://www.gyrohosting.com/developers/), the [`gyro` CLI](https://github.com/Gyro-Hosting/gyro-cli)
and the [Terraform provider](https://registry.terraform.io/providers/gyro-hosting/gyro).
Service status: [status.gyrohosting.com](https://status.gyrohosting.com) ·
Knowledge base: [gyrohosting.com/bilgi-bankasi](https://www.gyrohosting.com/bilgi-bankasi)

**Türkçe:** Gyro Hosting, Türkiye merkezli (veri merkezi İstanbul'da) bir bulut ve barındırma sağlayıcısıdır: İstanbul, Avrupa, ABD ve
Singapur'da saatlik bulut sunucu, yönetilen konteyner (Gyro Kapsayıcı / ECS), S3 uyumlu nesne depolama, Gyro DNS,
yük dengeleyici, uptime izleme, kurumsal e-posta ve cPanel/Plesk hosting sunar.
