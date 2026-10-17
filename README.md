# 🎮 Daily Word Game

A daily word-guessing game built with **Go**, backed by PostgreSQL and Redis, and containerized with Docker. The project includes automated testing, CircleCI configuration, and Kubernetes deployment manifests for running the application on a local cluster.

## ✨ Features

- **Go backend** — HTTP API for the word game.
- **Daily gameplay** — game state and guess-related functionality.
- **PostgreSQL** — stores word data and application records.
- **Redis** — supports game-state storage and caching.
- **Docker** — packages the application into a container image.
- **Docker Compose** — runs the application and its dependencies locally.
- **Kubernetes** — manifests for the application, database, Redis, and migrations.
- **Automated testing** — unit and integration test files.
- **CircleCI** — CI/CD pipeline configuration.

## 🛠️ Tech Stack

| Technology | Purpose |
|---|---|
| Go | Backend application |
| PostgreSQL 16 | Relational database |
| Redis 7 | Game-state storage and caching |
| Docker | Containerization |
| Docker Compose | Local multi-container orchestration |
| Kubernetes | Application deployment and orchestration |
| Minikube | Local Kubernetes cluster |
| kubectl | Kubernetes resource management |
| CircleCI | Continuous integration configuration |

## 📁 Project Structure

```text
Daily Word Game/
├── .circleci/
│   └── config.yml
├── game/
│   ├── errors.go
│   ├── score_test.go
│   └── score.go
├── k8s/
│   ├── 00-namespace.yaml
│   ├── 01-secrets.yaml
│   ├── 02-postgres.yaml
│   ├── 03-redis.yaml
│   ├── 04-migrate-job.yaml
│   ├── 05-app.yaml
│   └── 06-ingress.yaml
├── migrations/
├── tests/
│   └── integration_test.go
├── .dockerignore
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## ⚙️ Prerequisites

Install the following tools:

- [Go](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Git](https://git-scm.com/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## 🚀 Run Locally

### 1. Clone the repository

```bash
git clone https://github.com/shreeee244/word-game-go.git
cd word-game-go
```

### 2. Start the application

Build and start the services using Docker Compose:

```bash
docker compose up --build -d
```

Check the services:

```bash
docker compose ps
```

View application logs:

```bash
docker compose logs -f app
```

### 3. Test the application

Open these URLs in your browser:

- **Homepage:** http://localhost:8080/
- **Health check:** http://localhost:8080/health

Or use PowerShell:

```powershell
Invoke-RestMethod http://localhost:8080/
Invoke-RestMethod http://localhost:8080/health
```

A healthy response should report that the application is running and that PostgreSQL and Redis are connected.

### 4. Stop the application

```bash
docker compose down
```

To also delete the persisted database volume:

```bash
docker compose down -v
```

**Warning:** Removing the volume deletes the database data stored in it.

## ☸️ Kubernetes Deployment

This project includes Kubernetes manifests for PostgreSQL, Redis, database migrations, and the Go application.

### 1. Start Minikube

```bash
minikube start --driver=docker
```

Verify the cluster:

```bash
kubectl get nodes
```

### 2. Create the namespace

The workload manifests use the `word-game` namespace.

```bash
kubectl apply -f k8s/00-namespace.yaml
```

If that manifest doesn't create the `word-game` namespace, create it with:

```bash
kubectl create namespace word-game
```

### 3. Configure application secrets

```bash
kubectl apply -f k8s/01-secrets.yaml
```

Before publishing the repository, remove hard-coded credentials from the committed Secret manifest and use a secure method to provide credentials. Do not commit production passwords or access tokens.

### 4. Deploy PostgreSQL and Redis

```bash
kubectl apply -f k8s/02-postgres.yaml
kubectl apply -f k8s/03-redis.yaml
```

Wait for PostgreSQL:

```bash
kubectl rollout status deployment/postgres -n word-game --timeout=120s
```

Check the database storage claim:

```bash
kubectl get pvc -n word-game
```

### 5. Run database migrations

```bash
kubectl apply -f k8s/04-migrate-job.yaml
kubectl wait --for=condition=complete job/migrate -n word-game --timeout=180s
```

Inspect the migration logs:

```bash
kubectl logs job/migrate -n word-game
```

### 6. Deploy the Go application

```bash
kubectl apply -f k8s/05-app.yaml
kubectl rollout status deployment/app -n word-game --timeout=180s
```

Check the deployed resources:

```bash
kubectl get pods -n word-game
kubectl get services -n word-game
kubectl get jobs -n word-game
```

### 7. Access the application

Forward the Kubernetes Service to your local machine:

```bash
kubectl port-forward service/app 8080:8080 -n word-game
```

Keep the terminal open while testing the endpoints in another terminal.

**Homepage**

```powershell
Invoke-RestMethod http://localhost:8080/
```

**Health check**

```powershell
Invoke-RestMethod http://localhost:8080/health
```

**Retrieve a player's game state**

```powershell
Invoke-RestMethod "http://localhost:8080/game?player=Shrustie"
```

The `POST /guess` endpoint is also available. Refer to the request handler in `main.go` for its required request body.

## 🔌 API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/` | Returns service information and available endpoints |
| `GET` | `/health` | Reports application and dependency health |
| `GET` | `/game?player=NAME` | Retrieves the game state for a player |
| `POST` | `/guess` | Submits a guess |

## 🧪 Testing

The repository contains test files for game scoring and integration testing.

Run the Go test suite:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

The integration tests may require the application's dependencies to be available. Check `tests/integration_test.go` for any required configuration before running them.

## 🔄 Continuous Integration

The repository includes `.circleci/config.yml` for CircleCI pipeline configuration.

To enable the pipeline:

1. Push the repository to GitHub.
2. Connect the GitHub repository to CircleCI.
3. Review the jobs and workflows defined in `.circleci/config.yml`.
4. Configure any required environment variables or credentials in CircleCI.
5. Push a commit and review the pipeline result.

The exact checks and deployment actions depend on the workflow configured in `config.yml`.

## 🔍 Troubleshooting

**Check Kubernetes pods**

```bash
kubectl get pods -n word-game
```

**View application logs**

```bash
kubectl logs deployment/app -n word-game
```

**View migration logs**

```bash
kubectl logs job/migrate -n word-game
```

**Inspect a failing pod**

```bash
kubectl describe pods -n word-game
```

**Check Services**

```bash
kubectl get services -n word-game
```

If the application is unhealthy, verify that PostgreSQL and Redis are ready, the Services are available, and the environment variables match the configured service names.

## 🔐 Security

- Never commit real passwords, access tokens, or private keys.
- Use Kubernetes Secrets or an external secret manager for credentials.
- Configure GitHub and CircleCI credentials through their secure settings.
- Use strong credentials outside local development.
- Review `.gitignore` and `.dockerignore` before publishing.

## 🎯 Learning Outcomes

This project provides practical experience with:

- Building a backend application using Go.
- Designing and testing game-scoring logic.
- Integrating PostgreSQL and Redis.
- Containerizing a Go application with Docker.
- Running multi-container applications with Docker Compose.
- Deploying workloads with Kubernetes.
- Managing database migrations and persistent storage.
- Configuring health probes and Kubernetes Services.
- Writing Go unit and integration tests.
- Configuring a CircleCI pipeline.

author 
shrustie patil

## Git Branch Promotion Workflow

This project uses a pull-request-based workflow:

1. **Feature → main:** Submit feature changes through pull requests.
2. **main → staging:** Promote reviewed changes for validation.
3. **staging → prod:** Promote validated changes for release.

Each promotion is reviewed through a pull request before merging.

## Git Branch Promotion Workflow

1. Feature branches are merged into main through pull requests.
2. Main is promoted to staging through a pull request.
3. Staging is promoted to prod through a pull request.

Each promotion provides a review point before changes move forward.
