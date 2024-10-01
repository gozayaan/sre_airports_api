# ArgoCD Configuration

Argo CD is a declarative continuous delivery tool for Kubernetes.

In order to use Argo as a CD tool for this project:

## Prerequisites

- Prepare a Kubernetes cluster
- Install ArgoCD

```
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

- Install argocd CLI

## Configuration

1. Register the cluster on argo using `~/.kube/config`

```
kubectl config get-contexts -o <name>
argocd cluster add <cluster-name>
```

2. Settings > Projects > Create a new project with necessary info
<h1 align="center">
    <img alt="argo" src="static\proj1.jpeg" width="480" />
    <br>
</h1>
<h1 align="center">
    <img alt="argo" src="static\proj2.jpeg" width="550px" />
    <br>
</h1>

3. Settings > Repositories > Create a new repo and connect over HTTPS.

> 💡 Note: This repo has to be the GitOps repo containing kube manifests for release. For this purpose, I'm using [`bijoy26/bd-airports-manifests`](https://github.com/bijoy26/bd-airports-manifests/tree/dev) repository.

<h1 align="center">
    <img alt="argo" src="static\repo1.jpeg" width="480px" />
    <br>
</h1>
<h1 align="center">
    <img alt="argo" src="static\repo2.jpeg" width="550px" />
    <br>
</h1>

4. Applications > Create new app with necessary info.

<h1 align="center">
    <img alt="argo" src="static\app1.jpeg" width="400px" />
    <br>
</h1>

Make sure to specify the `synchronise` subdirectory under repo root, where the manifests to be synched reside.

<h1 align="center">
    <img alt="argo" src="static\app2.jpeg" width="480px" />
    <br>
</h1>

Specify the destination kube cluster's API server endpoint.

<h1 align="center">
    <img alt="argo" src="static\app3.jpeg" width="400px" />
    <br>
</h1>

5. Inspect the identified resources before **Synchronizing** the app for application deployment using **Continuous Delivery**

### Before Sync

<h1 align="center">
    <img alt="argo" src="static\sync0.jpeg" width="400px" />
    <br>
</h1>
<h1 align="center">
    <img alt="argo" src="static\sync1.jpeg" width="500px" />
    <br>
</h1>

### After Sync

<h1 align="center">
    <img alt="argo" src="static\sync2.jpeg" width="580px" />
    <br>
</h1>

6. Inspect kube resources from the cluster

Log onto cluster master and execute `kubectl get all -n <namespace>`. All the synced kube resources from argo will be visible.

<h1 align="center">
    <img alt="argo" src="static\mw1.jpeg" width="600px" />
    <br>
</h1>

> 💡 Clarification: Initially, the bd-airports app was deployed with the name `go-app` during dev-test and exposed as load balancer.
