# Boombox

<p><img src="https://ivan.vc/boombox/images/logo.png" alt="Boombox" title="Boombox" align="right" width="25%" style="padding-left: 10px"></p>

Boombox is a shell on-demand service that runs on Kubernetes. It listens on SSH
and creates a Pod in the cluster where the terminal will be attached.

* The first time a user logs in, it creates a Persistent Volume Claim, which will
  be mounted on `/home`. This ensures that there's persistence in the user's
  home.
* The user doesn't have sudo access, so the container is stateless.
* It uses [Homebrew](https://brew.sh), so the user can install new applications,
  which are persisted in `/home/linuxbrew`.

## Why?

Why don't you create a Pod and attach to it? I wanted to keep persistence for
the users while limiting access to the cluster. As this runs on SSH, they just
need to start a new terminal without requiring access to the Kubernetes
cluster. Still, they can access internal Kubernetes services.

## Deploying

Use the provided helm chart `ivanvc/boombox`.

Creating a private key for the server is strongly suggested. You can do it with
`ssh-keygen -t ed25519`. Then, provide it to the chart when deploying to your
cluster by setting it with `secrets.hostKey`.

### Configuration options

The following options can be set as environment variables (upper snake case and
with the `BOOMBOX_` prefix), or as arguments to the application. They are also
exposed in the Helm chart.

* `listen`: The `host:port` where to start the SSH daemon (default: `:2828`)
* `host-key-path`: The location for the host key path (default:
  `.ssh/boombox_ed25519`)
* `namespace`: The namespace where Boombox will create the PVCs and Pods
  (default: `default`, with Helm it defaults to the deployment namespace)
* `container-image`: The image for the Pod container (default: `ubuntu`)
* `pvc-size`: The size for the PVC that is mounted at `/home` (default: `10Gi`)
* `log-level`: The log level (default: `INFO`)

#### Setting the user shell

To set the user shell, create a file `~/.boombox_shell` with the content of the
shell to execute (i.e., `/bin/bash` or `/home/linuxbrew/.linuxbrew/bin/zsh`).

### Expose with an ingress

As Kubernetes ingresses don't have support for TCP, you need to follow the
following guide: [Exposing TCP and UDP services].

1. In your TCP services config map, add 2828 (assuming the default port)
   pointing to the Boombox service (assuming it's deployed in the boombox
   namespace).
  ```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tcp-services
  namespace: ingress-nginx
data:
  2828: "boombox/boombox:2828"
  ```

2. Modify the ingress controller service, by adding Boombox's port.
  ```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-nginx
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
spec:
  type: LoadBalancer
  ports:
    - name: http
      port: 80
      targetPort: 80
      protocol: TCP
    - name: https
      port: 443
      targetPort: 443
      protocol: TCP
    - name: proxied-tcp-2828
      port: 2828
      targetPort: 2828
      protocol: TCP
  selector:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
  ```

3. Ensure that the config map is in the ingress controller deployment args.
  ```
...
args:
  - /nginx-ingress-controller
  - --tcp-services-configmap=ingress-nginx/tcp-services
  ```

By following these steps, Boombox can be now reached by SSHing into the ingress
controller's host on port 2828.

## TODO

- [ ] Authentication
- [ ] Allow overriding Pod configuration scripts
- [ ] Allow to expose Pod ports, will need a service with defined ports, and
      an ingress
- [ ] Handle termination state of the Pod (`metadata.DeletionTimestamp`)

## License

See [LICENSE](LICENSE) © [Ivan Valdes](https://github.com/ivanvc/)

[Exposing TCP and UDP services]: https://kubernetes.github.io/ingress-nginx/user-guide/exposing-tcp-udp-services/
