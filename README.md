# Airport API

## System Architecture

```
          ┌──────────────┐   ┌──────────────┐
     ┌───►│  Jenkins CI  ├──►│ Execute      │     xxxxxxxxxxx xxxxxx
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

## Table of Contents

- [Instructions](#-instructions): Guide to make the system up and running end-to-end.
- [Environment Variables](#-environment-variables) : Reference sheet for environment variables used in the application
- [Activity Journal](#-activity-journal): Raw movement of my solution approach

- [References](#-references): Some resources that I picked up along the way

## 📃 Instructions

### 1. Setup CI System

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

### 2. Provision Object Storage

**Google Cloud Storage** is a service for storing objects in Google Cloud.

[fake-gcs-server](https://github.com/fsouza/fake-gcs-server) is a GCS emulator & testing library for mock purpose.

#### Setting up Mock GCS object storage

Use the [`./scripts/infra-spin-up.ps1`](scripts/infra-spin-up.ps1) PowerShell script if you are in Windows environment.

Or you can manually spin-up the mock GCS instance using docker.

```
docker run -d --name fake-gcs-server --network airport-net -p 4443:4443 -p 8000:8000 -v ${GCS_STORAGE_MOUNT_PATH}:/data fsouza/fake-gcs-server -scheme both -public-host localhost
```

> NOTE: When in make sure to always use _container / service instance name_ (docker network / kube CNI) for initializing GCS client connection. Modify ENV variables according to that if required.

### 4. Setup CD System & Deploy Application

Refer to [`pipeline/cd/argocd.md`](pipeline/cd/argocd.md) for guide on setting up the **continuous delivery** pipeline.

### 6. Sanity Test

Attempt to upload an airport image to the go application's `/update_airport_image` endpoint.

<h1 align="center">
    <img alt="argo" src="static\1.jpeg" width="700px" />
    <br>
</h1>

Response received with image upload completion.

### Check Container log

<h1 align="center">
    <img alt="argo" src="static\2.jpeg" width="650px" />
    <br>
</h1>

> 💡 NOTE: The uploaded object has been updated in both the datastores (v1 and v2). Also file upload is successful.

### 3. Configure API Gateway

**Kong API Gateway** is a lightweight, fast, and flexible cloud-native gateway suitable for canary release.

> Use the [`kubernetes/kube-setup.sh`](kubernetes/kube-setup.sh) for bootstrapping k8s associated networking objects, Kong API gateway necessities and object storage secrets.

Now you can cURL a bunch of requests against kong gateway instance for simulating `80:20` traffic split into `/airports`:`/airports_v2` endpoint.

### 7. TODO: Configure Monitoring

- [] To-Do - Incorporate prometheus SDK for exposing application metrics including _response time_.

- [x] Deploy [`kubernetes/network-setup/services/apigw-transparent/service.yaml`](kubernetes/network-setup/services/apigw-transparent/service.yaml) service to scrape **response_time** metrics from `/metrics` endpoint exposed via port `17701`.

---

## 🏭 Environment Variables

- **GCS_BUCKET_DOMAIN** : DNS name for \_Google Cloud Storage\* service i.e `storage.googleapis.com`

- **GCS_LOCALHOST_URL** : For self-hosted fake-gcs-server running in port 8000, use `http://localhost:8000/storage/v1/`

- **GCS_BUCKET_NAME** : Specify bucket name like `storage.googleapis.com`

- **GCS_STORAGE_MOUNT_PATH** : Specify filesystem source path for bucket to be mounted as a volume to GCS mock container

---

## 📚 Activity Journal

I tend to log my progress when I approach a problem for solution. Here are raw notes of how I navigated and also how circumvented when I encountered challenges.

<details>
	
<summary> Click to expand</summary>

- [x] write working function logic

  - [x] parse the request to collect multiplart form data (filename and image)
  - [x] sanity test successful image write in local FS

    ```
    curl -v  -X POST -H "Content-Type: multipart/form-data" -F "airport_name=Osmani International Airport" -F "airport_img=@/c/Users/ASUS/Desktop/az400.jpg" localhost:8080/update_airport_image
    ```

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

  - [x] test image upload functionality

- [x] write Dockerfile and containerize

  - issue encountered > GCS client's bucket listing doesn't work from container > solved
    - tested different combinations of existing host and bridge networks > no luck
    - tried different values of externalised env vars > no luck
    - tried modifying dockerfile to (easen up security) > no luck
    - take both containers under a new network with GCS referenced via container name > it works

  ```shell
  # docker-utilities
  DOCKER_BUILDKIT=1 docker build -f Dockerfile -t bd-airports .

  docker tag bd-airports hub.docker.com/bijoy26/bd-airports:v1.0

  docker login -u bijoy26

  docker run -d --name bd-airport-api -p 8080:8080 -e STORAGE_EMULATOR_HOST=http://localhost:4443 -e GCS_BUCKET_NAME=dev_bd_airport -e GCS_LOCALHOST_URL=http://localhost:8000/storage/v1/ -e GCS_BUCKET_DOMAIN=storage.googleapis.com bijoy26/bd-airports:v1.0

  docker push hub.docker.com/bijoy26/bd-airports:v1.0
  ```

- [x] Setup Kong API GW / Ingress to handle traffic split
- [x] Setup kube manifests
- [x] Setup Jenkinsfile
- [x] Setup terraform script

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
