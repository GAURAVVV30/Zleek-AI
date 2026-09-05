<div align="center">
  <img src="https://raw.githubusercontent.com/lucide-icons/lucide/main/icons/compass.svg" width="60" alt="Zleek AI Logo">
  <h1>Zleek AI (Amplified.AI)</h1>
  <p><strong>An AI-Driven Adaptive Learning Platform that builds your personalized learning path based on proven competence.</strong></p>
</div>

---

## 🎯 What it has (Key Features)

Zleek AI shifts the focus from passive content consumption to **active, proven competence**. 

- **12 Career Role Tracks:** Pick your domain, ranging from Software Architect, Machine Learning Engineer, and Data Scientist to Product Manager.
- **Adaptive Baseline Diagnostic:** A dynamic RAG + LLM powered quiz that establishes your exact starting point. You never waste time re-learning what you already know.
- **Dynamic Personalized Roadmaps:** Milestone-based learning paths are generated specifically based on your current experience level and diagnostic results.
- **Tia AI Agent:** An embedded AI assistant available in every module to instantly clear your doubts and explain complex concepts.
- **Custom Badges:** Earn unique custom badges upon successfully completing modules to showcase your verified skills.
- **Targeted Remediation:** Struggling with a specific concept? The platform automatically inserts targeted remediation resources before moving you forward.
- **Transparent Explainability:** Every node in your learning path tells you *why* you are learning it in relation to your end goal.
- **Evidence-Based Mastery:** Learning paths are continuously updated, and every piece of progress is meticulously noted by verifying skills, not just marking videos as complete.

---

## 🛠 What it has used (Tech Stack)

Zleek AI is built using a modern, scalable, and AI-native stack:

### Frontend
- **Framework:** React 18 & Vite
- **Styling:** TailwindCSS & Vanilla CSS
- **Animations:** Framer Motion
- **Deployment:** Nginx

### Backend API
- **Language:** Go (1.24+)
- **Architecture:** Clean Architecture principles
- **Authentication:** JWT Auth
- **Database Driver:** PGX

### AI & Data Engine
- **AI Pipeline:** In-Process Go AI Engine utilizing Gemini LLMs and Embeddings
- **Vector Search / RAG:** ChromaDB (for fast vector retrieval)
- **Primary Database:** PostgreSQL (with `pgvector` for semantic querying)
- **Caching & Pub/Sub:** Redis 7

---

## ⚙️ How it works (Architecture Flow)

1. **Goal & Onboarding:** The user selects 1 of the 12 distinct tech career roles and chooses an initial experience level (Beginner/Intermediate/Advanced).
2. **Diagnostic Evaluation (RAG + LLM):** The backend queries ChromaDB for relevant domain knowledge and context, passing this into an LLM to generate 5 highly specific, custom diagnostic questions.
3. **Roadmap Generation:** Based on the user's exact level and diagnostic performance, the system maps out a tailored learning roadmap. Known concepts are bypassed, and weak spots are prioritized.
4. **Learning & Assessment Loop:** 
   - The user engages with curated web resources (articles, docs, tutorials).
   - The **Tia AI Agent** is available at every step to answer questions and clear doubts instantly.
   - To unlock the next milestone, the user must pass an assessment (proving competence).
   - If the user struggles, the **Adaptive Remediation Engine** injects supplementary modules into the path on the fly.
5. **Progress & Rewards:** The user's learning path is dynamically updated as progress is continuously noted. Successful completions are rewarded with **Custom Badges** and logged in PostgreSQL as "Evidence", building a verified portfolio of skills.

---

## 🚀 How to use it (Quick Start Guide)

You can launch the entire ecosystem locally in under 3 minutes using Docker.

### 📌 Prerequisites
- **Docker Desktop** (or Docker Engine + Docker Compose) installed and running.
- Ensure the following ports are free: `5173` (Frontend), `8080` (Go Backend), `5432` (Postgres), `6379` (Redis), `8001` (ChromaDB).

### Step 1: Clone the Repository
```bash
git clone https://github.com/GAURAVVV30/Zleek-AI.git
cd Zleek-AI/
```

### Step 2: Configure Environment Variables
Copy the example environment file to set up your keys.
```bash
cp .env.example .env
```
*(Important: Open `.env` and add your `GEMINI_API_KEY` for live LLM generation to function correctly).*
*(Important: Open `.env` and add your `NVIDIA_API_KEY` for live LLM generation to function correctly).*
*(Important: Open `.env` and add your `GROQ_API_KEY` for live LLM generation to function correctly).*
*(Important: Open `.env` and add your `OPENAI_API_KEY` for live LLM generation to function correctly).*

### Step 3: Launch with Docker Compose
Pull the pre-built images and spin up the containers:
```bash
docker compose pull
docker compose up -d
```
or 
```bash
make dev
```
### Step 4: Access the Platform
Once all containers show as healthy, open your browser and navigate to:
👉 **[http://localhost:5173](http://localhost:5173)**

*Sign up for a new account, pick your career track, and take your first diagnostic to see the engine in action!*
