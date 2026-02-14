# josh.bot

A personal API-first platform for Josh, accessible at [josh.bot](https://josh.bot). This project serves as a dynamic, cloud-native backend that provides real-time status and information.

## Architecture

This project is built using a **Hexagonal Architecture** (also known as Ports and Adapters) in Go. This design pattern helps to isolate the core application logic from outside concerns, making the system more modular, testable, and maintainable.

- **`internal/domain`**: Contains the core business logic and entities.
- **`internal/adapters`**: Contains the adapters that connect the domain to external technologies (e.g., HTTP, AWS services).
- **`cmd/api`**: The main entrypoint for the application.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.22 or later)
- [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) (version 1.0 or later)
- An [AWS Account](https://aws.amazon.com/premiumsupport/knowledge-center/create-and-activate-aws-account/)

### Local Development

To run the server locally, which uses a mock service for immediate feedback, follow these steps:

1. **Clone the repository:**

    ```bash
    git clone https://github.com/vaporeyes/josh.bot.git
    cd josh.bot
    ```

2. **Run the server:**

    ```bash
    uv run go run cmd/api/main.go
    ```

    The server will start on `http://localhost:8080`.

3. **Test the endpoint:**
    In a separate terminal, you can curl the `/v1/status` endpoint:

    ```bash
    curl http://localhost:8080/v1/status
    ```

    You should see a JSON response similar to this:

    ```json
    {"current_activity":"Architecting josh.bot","location":"Clarksville, TN","status":"ok"}
    ```

## Deployment

Deployment is automated via a GitHub Actions CI/CD pipeline.

### Infrastructure

The infrastructure is managed with Terraform, located in the `terraform/` directory. It consists of:

- **AWS Lambda**: The Go application is deployed as a Lambda function.
- **API Gateway**: An HTTP API Gateway exposes the Lambda function to the internet.
- **IAM Roles**: Roles for Lambda execution and for GitHub Actions to deploy resources.
- **S3 Backend**: Terraform state is stored in an S3 bucket with native locking.

### CI/CD Pipeline

The CI/CD pipeline is defined in `.github/workflows/cicd.yaml` and triggers on every push to the `main` branch. The pipeline performs the following steps:

1. **Builds the Go binary** for a Linux ARM64 environment, compatible with AWS Lambda's `provided.al2023` runtime.
2. **Zips the binary** into a `function.zip` deployment package.
3. **Authenticates with AWS** using an OIDC role.
4. **Initializes and applies** the Terraform configuration to deploy the infrastructure.

To enable the pipeline, you need to set the `AWS_ACCOUNT_ID` secret in your GitHub repository settings.

## API Reference

### GET /v1/status

Returns the current status of the bot.

- **Success Response (200 OK):**

    ```json
    {
      "current_activity": "string",
      "location": "string",
      "status": "string"
    }
    ```
