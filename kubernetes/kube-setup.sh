# App namespace
kubectl create ns go

# GCS Mock Bucket
kubectl create secret generic gcs-bucket-name-secret --from-literal=gcs-bucket-name="dev_bd_airport" -n go

###### Kong API Gateway Setup ######

# Install k8s Gateway API CRD
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.1/standard-install.yaml

# install kong gateway class & gateway
kubectl apply -f ./network-setup/kong-apigw/gateway.yaml

# install kong helm chart
helm repo add kong https://charts.konghq.com
helm repo update
helm install kong/kong --generate-name --set ingressController.installCRDs=false -n kong --create-namespace

# Enable gateway support in kong
kubectl set env -n kong deployment/ingress-kong CONTROLLER_FEATURE_GATES="GatewayAlpha=true" -c ingress-controller
kubectl rollout restart -n NAMESPACE deployment DEPLOYMENT_NAME

# Deploy HTTPRoute attached to the Kong APIGW for traffic split routing
kubectl apply -f ./network-setup/kong-apigw/httproute.yaml

# Deploy standard service (stable) for 100% /airports routing
kubectl apply -f ./network-setup/services/apigw-transparent/service.yaml

# Deploy traffic-split services (stable and canary) for 80:20 /airports_v2 routing
kubectl apply -f ./network-setup/services/apigw-traffic-split/service_split.yaml
