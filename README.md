# Airport API

## 1. Provision a Cloud Storage Bucket Using Infrastructure as Code (IaC)

**Decision**: Use Terraform for provisioning a Google Cloud Storage (GCS) bucket.

- **Reasoning**: Terraform is a widely-used Infrastructure as Code (IaC) tool that allows for repeatable and version-controlled infrastructure setup. Using Terraform ensures that our infrastructure can be easily recreated or modified.
- **Outcome**: A GCS bucket named `airport-images-bucket` was provisioned with versioning enabled to manage image updates effectively.

---

## 2. Make an Endpoint `/update_airport_image` to Update an Airport’s Image

**Decision**: Implement a RESTful endpoint in a Go application to allow uploading images for airports.

- **Reasoning**: Go is a performant language for building web applications, and it integrates well with Google Cloud services. Creating a RESTful endpoint enables standard HTTP methods for interacting with the application, making it easy to manage airport images.
- **Outcome**: The `/update_airport_image` endpoint allows users to upload images associated with specific airports, providing feedback on the success or failure of the operation.

---

## 3. Containerize the Go Application

**Decision**: Use Docker to create a containerized version of the Go application.

- **Reasoning**: Containerization ensures consistency across different environments (development, staging, production) by packaging the application and its dependencies together. It also simplifies deployment and scaling processes.
- **Outcome**: A multi-stage Dockerfile was created to build the application efficiently, resulting in a minimal image suitable for production deployment.

---

## 4. Prepare a Deployment and Service Resource to Deploy in Kubernetes

**Decision**: Deploy the application using Kubernetes with a Deployment and a Service resource.

- **Reasoning**: Kubernetes provides powerful orchestration capabilities, allowing for automated deployment, scaling, and management of containerized applications. The Deployment resource allows us to define how many replicas of our application should run, ensuring high availability.
- **Outcome**: The Kubernetes configuration includes a Deployment with three replicas and a LoadBalancer Service to expose the application externally.

---

## 5. Use API Gateway to Create Routing Rules to Send 20% of Traffic to the `/airports_v2` Endpoint

**Decision**: Utilize Google Cloud API Gateway for managing traffic routing between different endpoints.

- **Reasoning**: The API Gateway allows for centralized management of APIs, enabling features such as traffic splitting, authentication, and monitoring. Traffic splitting is used to gradually roll out new features (in this case, the `/airports_v2` endpoint) while ensuring existing functionality remains available.
- **Outcome**: Configurations were created to split the traffic such that 80% of requests go to the existing `/airports` endpoint, and 20% go to the new `/airports_v2` endpoint, allowing for controlled testing of the new version.

---

## Testing and Validation

**Decision**: Use `curl` commands to test endpoint functionality and traffic distribution.

- **Reasoning**: Automated testing using command-line tools like `curl` helps validate that the endpoints are functioning as expected and that the API Gateway is correctly distributing traffic.
- **Outcome**: A series of requests were sent to ensure the expected response rates and to monitor the behavior of both endpoints under load.

---

## Conclusion

These decisions were made to ensure a robust, scalable, and maintainable cloud infrastructure and application. Each step was aimed at leveraging modern technologies to enhance the application's performance and reliability.


---
_For tasks, checkout [tasks.md](tasks.md)_
---
_For solutions, checkout [solutions.md](solutions.md)_
