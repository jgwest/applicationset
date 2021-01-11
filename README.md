# Argo CD ApplicationSet CRD

The Argo CD ApplicationSet CRD and controller provides a formal solution to replace the
[app-of-apps](https://argoproj.github.io/argo-cd/operator-manual/cluster-bootstrapping/) pattern
with the ultimate goal of introducing ApplicationSet as a first class supported object in 
Argo CD Core.

This project will initially be maintained separately from Argo CD, in order to allow quick iteration
of the spec and implementation, without tying it to Argo CD releases. No promises of backwards
compatibility are made, at least until merging into Argo CD proper.

## Proposal:
* https://docs.google.com/document/d/1juWGr20FQaJmuuTIS8mBFmWWDU422M_FQMuhp5c1jt4/edit?usp=sharing

## Example Spec:

```yaml
# This is an example of a typical ApplicationSet which uses the cluster generator.
# An ApplicationSet is comprised with two stanzas:
#  - spec.generator - producer of a list of values supplied as arguments to an app template
#  - spec.template - an application template, which has been parameterized
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: guestbook
spec:
  generators:
  - clusters: {}
  template:
    metadata:
      name: '{{name}}-guestbook'
    spec:
      source:
        repoURL: https://github.com/infra-team/cluster-deployments.git
        targetRevision: HEAD
        chart: guestbook
      destination:
        server: '{{server}}'
        namespace: guestbook
```

Additional examples are available in the [examples](./examples) directory.


## Development

### Running the ApplicationSet controller as an image within Kubernetes

The following assumes you have: 
1. Installed a recent version of [kustomize](https://github.com/kubernetes-sigs/kustomize) (3.x+). 
2. Created a container repository for your development image.
  - For example, by creating a repository "(your username)/argocd-applicationset" using [Docker Hub](https://hub.docker.com/) or [Red Hat Quay.io](https://quay.io/).
3. Ran `docker login` from the CLI, and provided your registry credentials.
4. Deployed ArgoCD into the `argocd` namespace.
  - To install Argo CD, follow the [Argo CD Getting Started](https://argoproj.github.io/argo-cd/getting_started/) guide.

To build and push a container with your current code, and deploy Kubernetes manifests for the controller Deployment:

```bash
# Build and push the image to container registry
IMAGE="(username)/argocd-applicationset:v0.0.1" make image

# Deploy the ApplicationSet controller manifests
IMAGE="(username)/argocd-applicationset:v0.0.1" make deploy
```

The ApplicationSet controller should now be running in the `argocd` namespace.


### Running the ApplicationSet Controller as a standalone process from the CLI

When developing a Kubernetes controller, it is often easier to run the controller process from your local CLI, rather than requiring a container rebuild and push for new code changes.

1. First, setup a local Argo CD development environment:
  - Clone the Argo CD source, and setup an Argo CD dev environment:
    - [Setting up your development environment](https://argoproj.github.io/argo-cd/developer-guide/contributing/#setting-up-your-development-environment)
    - [Install the must-have requirements](https://argoproj.github.io/argo-cd/developer-guide/contributing/#install-the-must-have-requirements)
    - [Build your code and run unit tests](https://argoproj.github.io/argo-cd/developer-guide/contributing/#build-your-code-and-run-unit-tests)
 
2. Ensure that port 8081 is exposed in the Argo CD test server container:
- In the `Makefile` file at the root of the Argo CD repo:
    - Add the following to [this location in the Makefile](https://github.com/argoproj/argo-cd/blob/27912a08f151fab038ddb804a618ca8cde01d68e/Makefile#L75)
    - Replace: `-p 4000:4000 \`
    - With: `-p 4000:4000 -p 8081:8081 \`
    - This exposes port 8081 (the repo-server listen port), which is required for ApplicationSet Git generator functionality.

3. Ensure that Argo CD is running
- Ensure your active namespace is set to `argocd` (for example, `kubectl config view --minify | grep namespace:`).
- Run `make start` under the Argo CD dev environment.
- Wait for the Argo CD processes to start within the container.
- This process should remaining running, alongside the local ApplicationSet controller, during the following steps.
- Verify that:
    - You have exposed port 8081 in the Makefile (as described in prerequisites). `docker ps` should show port 8081 as mapped to an accessible IP.

4. Apply the ApplicationSet CRDs into the `argocd` namespace, and build the controller:
```
kubectl apply -f manifests/crds/argoproj.io_applicationsets.yaml
make build
```

5. Run the Application Set Controller from the CLI:
```
NAMESPACE=argocd ./dist/argocd-applicationset --metrics-addr=":18081" --probe-addr=":18082"
```

On success, you should see (amongst other text):
```
INFO	controller-runtime.controller	Starting Controller	{"controller": "applicationset"}
INFO	controller-runtime.controller	Starting workers	{"controller": "applicationset", "worker count": 1}
```

