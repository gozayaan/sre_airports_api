# Airports API

> [!IMPORTANT]
> This repository can be treated as a solution to a set of take-away problems, where the problem definitions and requirements are described in the [`tasks`](tasks.md) section.

**Airports API** is a lightweight API that provides information about airports of Bangladesh.

## System Architecture

```
          ┌──────────────┐   ┌──────────────┐
     ┌───►│  Jenkins CI  ├──►│ Execute      │     xxxxxxxxxxxxxxxxxx
     │    │    Trigger   │   │ Build Job    │     x                x
     │    └──────────────┘   │              ├────►x  Docker Hub    x
     │                       │ - Checkout   │Push x   Registry     x
     │                       │ - Img Build  │     x                x
┌────┴─────────────────────┐ │ - Img tag    │     xxxxxxxxxxxxxxxxxx
│                          │ │ - Img push   │              │
│ bijoy26/sre_airports_api │ │ - GitOps tag │              │
│                          │ │ - Alerting   │              │
└──────────────────────────┘ └───────┬──────┘              │
                                   ▲ │                     │ Pull
┌────────────────────────────────┐ │ │                     │
│                                ├─┘ │                     │
│ bijoy26/bd-airports-manifests  ◄───┘                     │
│                                │                         ▼
└───────────┬────────────────────┘                  xx x x xx xxxxxxxxxx
            │                                       x                  xxx
          Watch        ┌──────Synchronise───────► xx                     xxx
           for  ┌──────┴──────────┐              xx                        xx
          commit│                 │            xxx        K8s Cluster        xx
            │   │  ArgoCD Sync    │            x                              x
            └──►│                 │       ┌─────────┐     ┌───────────────┐   x
                └───────┬─────────┘       │         ├────►│bd-airports-app│   x
                                          │  Kong   │     │   /airports   │   x
┌─────────────┐    xxxxxxxxx xxxxxx       │   API   │     │               │   x
│             │    xx            x        │ Gateway │     │               │   x
│  Terraform  │     xx   GCS    xx        │         ├────►│bd-airports-app│   x
└─────┬───────┘     xx  Bucket  x         └─────────┘     │  /airports_v2 │   x
      │             xxx        xx              x          └──────┬────────┘   x
      └─Provision──► xxx      xxx     Upload   xx                │            x
                     xxxx    xxx  ◄──────────────────────────────┘           xx
                      xxxxxxxxxx                xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## 📜 Table of Contents

- [Instructions](#-instructions): Guide to make the system up and running end-to-end.
- [Environment Variables](#-environment-variables) : Reference sheet for environment variables used in the application
- [Activity Journal](#-activity-journal): Raw movement of my solution approach

- [References](#-references): Some resources that I picked up along the way

- [Courtesy](#-courtesy): Ad-hoc tools / softwares that helped me reach this far

## 📃 Instructions

### 1. 👨‍🍳 Setup CI System

**Jenkins** is an open source automation server for reliable continuous integration.

Setup a jenkins pipeline by creating a project with following [Jenkinsfile](/pipeline/ci/Jenkinsfile).

<details>
	
<summary> Click to Expand </summary>

```groovy
pipeline{
    agent any
    environment{
        DOCKERHUB_CREDENTIALS_USER = 'bijoy26'
        DOCKERHUB_CREDENTIALS = credentials('dockerhub-creds-user')
        CODEBASE_REPO_BRANCH = 'main'
        GITOPS_REPO_BRANCH = 'dev'
        PREVIOUS_TAG = 'v1.6'
        LATEST_TAG = 'v1.7'
        CODEBASE_REPO = 'https://github.com/bijoy26/sre_airports_api.git'
        GITOPS_REPO = 'https://github.com/bijoy26/bd-airports-manifests.git'
        GITOPS_SVC_ACCOUNT = 'enter-gitops-email@domain.tld'
        GITOPS_SVC_ACCOUNT_PAT = 'ENTER-CREDENTIAL-HERE'
    }

    stages{
        stage('Checkout Latest Release from SCM'){
            steps{
                git branch: "${CODEBASE_REPO_BRANCH}", credentialsId: 'read-pat', url: "${CODEBASE_REPO}"
            }
        }

        stage('Build Container Artifacts for BD Airports API'){
            steps{
                script {
                    env.GIT_BRANCH = sh([script: "git branch", returnStdout: true]).trim()
                }
                sh 'DOCKER_BUILDKIT=1 docker build -f ../Dockerfile -t bd-airports ../src'
            }
        }

        stage('Authenticate to Docker Hub Registry') {
            steps {
                sh 'echo $DOCKERHUB_CREDENTIALS | docker login -u $DOCKERHUB_CREDENTIALS_USER --password-stdin'
            }
        }

        stage('Retag & Push Image to Registry'){
            steps{
                sh '''
                docker tag bd-airports bijoy26/bd-airports:$LATEST_TAG'
                docker push bijoy26/bd-airports:$LATEST_TAG
                '''
            }
        }

        stage('Push Changes to GitOps Repository for Latest Release'){
            steps{
                deleteDir()
                git branch: "${GITOPS_REPO_BRANCH}", credentialsId: 'write-pat', url: "${GITOPS_REPO}"
                sh "/usr/bin/sed -i 's|bijoy26/bd-airports:$PREVIOUS_TAG|bijoy26/bd-airports:$LATEST_TAG|g' synchronise/deployment.yaml"
                sh "git remote rm origin"
                sh "git remote add origin https://'${GITOPS_SVC_ACCOUNT}':'${GITOPS_SVC_ACCOUNT_PAT}'@github.com/bijoy26/sre_airports_api"
                sh "git add -A"
                sh "git config user.email '${GITOPS_SVC_ACCOUNT}'"
                sh "git config user.name '${GITOPS_SVC_ACCOUNT}'"
                sh "git commit -m 'BUMP image tag $LATEST_TAG'"
                sh "git push -u origin ${GITOPS_REPO_BRANCH}"
            }
        }
        }

    post{
        success{
          // Send Slack notification alert
           slackSend (
                    color: "good",
                    channel: '#ADD-SLACK-CHANNEL-NOTIFICATION-HERE',
                    message: "Build success for ${env.JOB_NAME} pipeline with #${env.BUILD_NUMBER} build from ${env.GIT_BRANCH} branch",
                    tokenCredentialId: 'Slack-notifier',
                )
        }
          // Send Slack notification alert
        failure{
            slackSend (
                    color: "danger",
                    channel: '#jenkins-slack-notification',
                    message: "Build failure for ${env.JOB_NAME} pipeline with #${env.BUILD_NUMBER} build from ${env.GIT_BRANCH} branch",
                    tokenCredentialId: 'Slack-notifier',
                )
        }
    }
}
```

</details>

### 2. 🧺 Provision Object Storage

**Google Cloud Storage** is an _object storage service_ for storing data in Google Cloud Platform.

Here are two approaches to test buckets in GCS:

1. [Cloud: Provision GCS bucket using IaC](#21-provision-a-gcs-bucket-using-terraform)
2. [Self-host: Spin-up Mock GCS instance using Docker](#22-setting-up-mock-gcs-object-storage)

> [!TIP]
> 💡 For local development, it is useful to have mock / emulation servers at your disposal.

#### 2.1 Provision a GCS bucket using Terraform

**Terraform** is an _IaC (Infrastructure as Code_) tool that can automate the provisioning and management of GCS buckets in a replicable manner.

> [!IMPORTANT]
> Install terraform and use the [`terraform.md`](iac/terraform.md) guide for bucket setup and [`/iac/provision.sh`](iac/provision.sh) script for reference purpose.

#### 2.2 Setting up Mock GCS Object Storage

To dev/test GCS functions such as bucket file upload in local machine, we can use [fake-gcs-server](https://github.com/fsouza/fake-gcs-server), which is a GCS emulator & testing library.

> [!IMPORTANT]
> Execute the [`/scripts/infra-spin-up.ps1`](scripts/infra-spin-up.ps1) PowerShell script if you are in Windows environment to have things ready.

Or you can export `GCS_STORAGE_MOUNT_PATH` and spin-up the mock GCS instance using docker.

Following command deploys an instance with HTTPS endpoint on `4443` and HTTP on `8000`.

```bash
docker run -d --name fake-gcs-server --network airport-net -p 4443:4443 -p 8000:8000 -v ${GCS_STORAGE_MOUNT_PATH}:/data fsouza/fake-gcs-server -scheme both -public-host localhost
```

> [!TIP]
> 💡 Make sure to always use _container / service name_ (docker network / kube CNI) while initializing GCS client connection.

To quickstart, you can list bucket contents with `curl --insecure https://127.0.0.1:4443/storage/v1/b`

### 3. 🐙 Setup CD System & Deploy Application

**Argo CD** is a declarative continuous delivery tool for Kubernetes.

> [!IMPORTANT]
> Refer to [`pipeline/cd/argocd.md`](pipeline/cd/argocd.md) for guide on setting up the **continuous delivery** pipeline with cloud-native solution ArgoCD.

### 4. 🔨 Sanity Test for Airport Image Update

Attempt to upload an airport image to the go application's `/update_airport_image` endpoint.

> [!NOTE]
> Since _**OSMANI INTERNATIONAL AIRPORT**_ value is passed as form data, it should be accepted by the server.

<h1 align="center">
    <img alt="argo" src="static\1.jpeg" width="700px" />
    <br>
</h1>

**HTTP 200** Response received with image upload completion. Happy path is working. It will be rejected for a mismatch.

To double check, we can inspect container logs.

#### Inspect Container log

<h1 align="center">
    <img alt="argo" src="static\2.jpeg" width="650px" />
    <br>
</h1>

> [!NOTE]
> 💡 The uploaded object has been updated in both the datastores (v1 and v2) and the file upload is successful.

### 5. 🦍 Configure API Gateway

**Kong API Gateway** is a lightweight, fast, and flexible cloud-native gateway suitable for canary release.

> [!IMPORTANT]
> 💡 Use the [`kubernetes/kube-setup.sh`](kubernetes/kube-setup.sh) for bootstrapping k8s associated networking objects, Kong API gateway necessities and object storage secrets.

Now, to test the API gateway, you can initiate a bunch of requests against gateway instance for simulating `80:20` traffic split into `/airports`:`/airports_v2` endpoints.

```bash
for x in $(seq 1 15); do curl -s --resolve bd-airports.local:80:<Ingress-Controller-IP> bd-airports.local/airports; done
```

### 6. 🔥 TODO: Configure Monitoring

- [] To-Do - Incorporate prometheus SDK for exposing application metrics including _response time_.

- [x] Deploy [`kubernetes/network-setup/services/apigw-transparent/service.yaml`](kubernetes/network-setup/services/apigw-transparent/service.yaml) service to scrape **response_time** metrics from `/metrics` endpoint exposed via port `17701`.

---

## 🏭 Environment Variables

- **GCS_BUCKET_DOMAIN** : DNS name for _Google Cloud Storage_ service i.e `storage.googleapis.com`

- **GCS_LOCALHOST_URL** : For self-hosted fake-gcs-server running in port 8000 (HTTP), use `http://localhost:8000/storage/v1/` to access from host machine or alternatively, use `http://fake-gcs-server:8000/storage/v1/` to access from docker / k8s network.

- **GCS_BUCKET_NAME** : Specify bucket name like `bd-airport-data`

- **GCS_STORAGE_MOUNT_PATH** : Specify host filesystem path for the storage volume to be mounted on the mock container

- **GOOGLE_CLOUD_PROJECT** : For provisioning bucket using terraform, specify the ID of the project in which the resource belongs.

---

## 📚 Activity Journal

I tend to log my progress when I approach a problem for solution. Here are raw notes of how I navigated and also how circumvented when I encountered challenges.

<details>
	
<summary> Click to expand</summary>

- [x] Explore and Setup GCS service

  - [x] Attempt to create a google cloud account - abandon option as credit card being declined because of payment method not being supported. `[OR-CCSEH-34]`

  - [x] Alternative: find an GCS emulator to replicate mock environment

    - [x] setup [fake-gcs-server](https://github.com/fsouza/fake-gcs-server) using docker for quick infra management

      ```
       # create bucket in local FS
       mkdir -p "${PWD}"/mock_gcs_storage
       cd "${PWD}"/mock_gcs_storage && touch file.txt

       docker run -d --name fake-gcs-server -p 4443:4443 fsouza/fake-gcs-server

       # try git bash > got issue git-bash mintty docker mount issue
       docker run -d --name fake-gcs-server -p 4443:4443 -v "${PWD}"/mock_gcs_storage:/data fsouza/fake-gcs-server

       # try in powershell > works
       cd "G:\Career Stuff\GoZayaan - SRE\Stage-1 (Sep 2024)"
       docker run -d --name fake-gcs-server -p 4443:4443 -v ${PWD}/mock_gcs_storage:/data fsouza/fake-gcs-server

       # HTTPS port 4443, HTTP port 8000
       docker run -d --name fake-gcs-server -p 4443:4443 -p 8000:8000 -v ${PWD}/mock_gcs_storage:/data fsouza/fake-gcs-server -scheme both

       # 'localhost' configured as public host
       docker run -d --name fake-gcs-server -p 4443:4443 -p 8000:8000 -v ${PWD}/mock_gcs_storage:/data fsouza/fake-gcs-server -scheme both -public-host localhost

      ```

    - [x] quick test mock GCS APIs with sample bucket

      ```
      # from inside container
      apk add curl
      curl --insecure https://127.0.0.1:4443/storage/v1/b # list buckets
      curl --insecure https://127.0.0.1:4443/storage/v1/b/dev_bd_airport/o    # list objects

      # from host machine
      curl --insecure https://127.0.0.1:4443/storage/v1/b # list buckets
      ```

    - [x] write a Go API upload / delete client to interact with the sample bucket

    ```
    export STORAGE_EMULATOR_HOST=http://localhost:4443

    # set in powershell for to override 'could not find default credentials' error
    $env:STORAGE_EMULATOR_HOST = 'http://localhost:4443'
    # set bucket name
    $env:GCS_BUCKET_NAME = 'dev_bd_airport'

    go build main.go    # fake gcs example api
    ./main.exe
    ```

    - got `failed to list: googleapi: got HTTP response code 400 with body: Client sent an HTTP request to an HTTPS server.` > Reconfigure server http mode with proper flag > it works.

    - Can list bucket contents now but got `storage: object doesn't exist` > try changing "-public-host" to "localhost" > it works.

    - Can DELETE bucket contents now

      - [x] Write and test upload and delete API
      - [] (OPTIONAL) turn upload snippet into standalone function

- [x] Write the actual image upload API as per instructions

  - [x] write working function logic

    - [x] parse the request to collect multiplart form data (filename and image)
    - [x] test image upload functionality

      - [x] sanity test successful image write in local FS

        ```
        curl -v  -X POST -H "Content-Type: multipart/form-data" -F "airport_name=Osmani International Airport" -F "airport_img=@/c/Users/ASUS/Desktop/az400.jpg" localhost:8080/update_airport_image
        ```

- [x] write Dockerfile and containerize

  - issue encountered > GCS client's bucket listing doesn't work from container > solved
    - tested different combinations of existing host and bridge networks > no luck
    - tried different values of externalised env vars > no luck
    - tried modifying dockerfile to (easen up security) > no luck
    - take both containers under a new network with GCS referenced via container name > it works

- summarize the local-build-push commands

  ```shell
  # docker-utilities
  DOCKER_BUILDKIT=1 docker build -f Dockerfile -t bd-airports .

  docker tag bd-airports hub.docker.com/bijoy26/bd-airports:v1.0

  docker login -u bijoy26

  docker run -d --name bd-airport-api -p 8080:8080 -e STORAGE_EMULATOR_HOST=http://localhost:4443 -e GCS_BUCKET_NAME=dev_bd_airport -e GCS_LOCALHOST_URL=http://localhost:8000/storage/v1/ -e GCS_BUCKET_DOMAIN=storage.googleapis.com bijoy26/bd-airports:v1.0

  docker push hub.docker.com/bijoy26/bd-airports:v1.0
  ```

- [x] Setup API GW to handle traffic split

  - [x] Explore available tools that support _canary release with weighted rewrites_
    - [x] Check **Kong API Gateway** capabilties
      - DBless mode with declarative approach
        - [x] Check **canary release** plugin > not enough resource available for weighted canary
        - [x] Check kong-native configurations > not enough resource available
        - [x] Check Kong **HTTPRoute** capabilities for k8s gateway api > **URLRewrite** and **weight** properties available
          - [x] Check HTTPRoute API spec for _weighted URLRewrite_ property > available > solution found

- [x] Setup kube manifests
- [x] Setup Jenkinsfile
- [x] Setup ArgoCD
- [x] Setup terraform configuration

</details>

---

## 📃 References

- HTML form file upload https://dev.to/wassimbj/how-to-upload-files-in-golang--10p5
- Multi-part form upload CURL ref https://reqbin.com/req/c-sma2qrvp/curl-post-form-example
- Golang io.copy snippet- https://medium.com/google-cloud/golang-copy-to-gcs-check-bucket-58721285788e
- Using Kong API gateway to access K8s services- https://medium.com/@martin.hodges/using-kong-to-access-kubernetes-services-using-a-gateway-resource-with-no-cloud-provided-8a1bcd396be9
- Kong LB ref - https://docs.konghq.com/gateway/latest/how-kong-works/load-balancing/
- k8s Gateway API HTTPRoute Ref - https://gateway-api.sigs.k8s.io/api-types/httproute/
- HTTPRoute API Spec - https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter
- HTTRoute RewriteURL - https://gateway-api.sigs.k8s.io/guides/http-redirect-rewrite/

---

## 🙏 Courtesy

- _System architecture_ diagram made with ❤ by [ASCIIFlow](https://asciiflow.com/)
