# 🚀 Amplified.AI — AI-Driven Adaptive Learning Platform

> **Jury & Evaluator Setup Guide**  
> Complete local setup and evaluation instructions for running the Amplified.AI platform using pre-built Docker Hub containers or source code.

---

## 📌 Prerequisites

Before running the application, ensure your system has:
- **Docker Desktop** (or Docker Engine + Docker Compose) installed and running.
- Port availability: `5173` (Frontend), `8080` (Go Backend), `5432` (PostgreSQL), `6379` (Redis), `8001` (ChromaDB).

---

## ⚡ Method 1: Instant Jury Setup via GitHub Repository (Recommended)

### Step 1: Clone the Repository
```bash
git clone https://github.com/darshanar190607/Ai_Amplified_Challenge.git
cd Ai_Amplified_Challenge
```

### Step 2: (Optional) Set API Key for Live LLM Features
Copy the example environment file:
```bash
cp .env.example .env
```
*(Optional: Add your `GEMINI_API_KEY` or `GROQ_API_KEY` to `.env` for live LLM generation).*

### Step 3: Launch with Docker Compose
```bash
docker compose up -d
```

### Step 4: Open in Browser
Open your browser and navigate to:
👉 **[http://localhost:5173](http://localhost:5173)**

---

## 🐳 Method 2: Instant Setup via Docker Hub Images

If you do not want to clone the full repository code, create a `docker-compose.yml` file in any folder with the following content:

```yaml
services:
  db:
    image: ankane/pgvector:latest
    container_name: amplified_db
    restart: always
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-postgres}
      POSTGRES_DB: ${POSTGRES_DB:-platform}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-postgres} -d ${POSTGRES_DB:-platform}"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: amplified_redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  chroma:
    image: chromadb/chroma:latest
    container_name: amplified_chroma
    restart: always
    ports:
      - "8001:8000"
    volumes:
      - chroma_data:/chroma/chroma_data
    environment:
      - IS_PERSISTENT=TRUE
      - PERSIST_DIRECTORY=/chroma/chroma_data

  api-go:
    image: wtfwizz30/amplified-api-go:latest
    container_name: amplified_api_go
    restart: always
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=${POSTGRES_USER:-postgres}
      - DB_PASSWORD=${POSTGRES_PASSWORD:-postgres}
      - DB_NAME=${POSTGRES_DB:-platform}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - CHROMA_HOST=chroma
      - CHROMA_PORT=8000
      - GEMINI_API_KEY=${GEMINI_API_KEY}
      - GROQ_API_KEY=${GROQ_API_KEY}
      - NVIDIA_API_KEY=${NVIDIA_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
      chroma:
        condition: service_started

  web:
    image: wtfwizz30/amplified-web:latest
    container_name: amplified_web
    restart: always
    ports:
      - "5173:80"
      - "80:80"
    depends_on:
      - api-go

volumes:
  postgres_data:
  redis_data:
  chroma_data:
```

### Launch Commands:
```bash
# Pull images from Docker Hub
docker compose pull

# Start containers
docker compose up -d
```
Access the application at: 👉 **[http://localhost:5173](http://localhost:5173)**

---

## 🎯 Jury Evaluation Walkthrough & Feature Verification

### 1️⃣ Authentication & Onboarding
1. Go to `http://localhost:5173`.
2. Click **Sign Up** or **Get Started** to create a test user.
3. Upon registration, you will land on `/onboarding/goal`.

### 2️⃣ 12 Career Role Selection
- Verify that **all 12 Career Roles** are exposed as interactive domain cards:
  - Software Architect
  - Machine Learning Engineer
  - Data Scientist
  - AI Engineer
  - Full Stack Developer
  - Backend Engineer (Go/Node)
  - Frontend Engineer
  - Data Analyst
  - Data Engineer
  - DevOps / SRE
  - Mobile Engineer
  - Product Manager

### 3️⃣ Adaptive Diagnostic Quiz (RAG + LLM Powered)
- Choose your target role and prior experience level (*Beginner*, *Intermediate*, or *Advanced*).
- System dynamically fetches authoritative RAG context for the selected domain and generates 5 custom prerequisite questions.

### 4️⃣ Diagnostic Results & Personalised Roadmap
- Submit answers to view calculated competency scoring, accuracy, and identified weak concepts.
- Continue to `/roadmap` to see your AI-generated milestone learning graph tailored strictly to your target career role.

---

## 🛠 Tech Stack Architecture

- **Frontend**: React 18, Vite, TailwindCSS, Framer Motion, Nginx
- **Backend API**: Go (1.24+), Clean Architecture, PGX, JWT Auth
- **AI / RAG Pipeline**: In-Process Go AI Engine, Gemini Embeddings, RAG Retriever
- **Database & Storage**: PostgreSQL (with `pgvector`), Redis 7, ChromaDB
