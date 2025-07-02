# 🚀 Feather CI/CD Auto-Registration
![[feather-logo.png]]
>Git 저장소를 Argo Workflows 및 ArgoCD와 자동으로 연결하여 CI/CD를 손쉽게 구현

![[feather-intro|800]]

---

## 🌐 Environment

| Category          | Tech Stack                                              |
| ----------------- | ------------------------------------------------------- |
| **Language**      | Go                                                      |
| **Framework**     | Gin (Go)                                                |
| **IDE**           | Cursor, GoLand                                          |
| **Dev Env**       | Kubernetes, Docker, Gitea, Argo Workflows, ArgoCD, Dive |
| **Service Mesh**  | Istio                                                   |
| **Observability** | OpenTelemetry, Loki, Prometheus, Tempo, Kiali           |
| **AI**            | ChatGPT                                                 |
| DB                | MySQL                                                   |

---

## 🧩 Summary

Feather는 **Git 템플릿으로 생성된 저장소를 자동 등록**하여 Argo Workflows와 ArgoCD로 연동함으로써, **프로젝트가 생성되는 순간부터 자동 배포가 가능한 CI/CD  파이프라인**을 구축합니다.

---

## 🛠️ Features

- 🔨 **Gitea API를 통해 템플릿 기반 저장소 생성**
- ⚙️ **Argo Workflows에 CI 파이프라인 자동 구성**
- 🚀 **ArgoCD에 GitOps 방식의 CD 자동 연동**
- 🔍 **Prometheus, Tempo, Loki, Kiali를 통한 완전한 관측 가능성 제공**


---
## 🗂️ Gitea란?

**Gitea**는 가볍고 빠르며 자체 호스팅 가능한 Git 서비스입니다. GitHub과 유사한 기능을 제공하며, DevOps 환경에서 **내부 코드 저장소**로 활용하기 좋습니다.

### 💡 주요 특징

- **GitHub과 유사한 웹 UI**
- **RESTful API**를 통한 자동화 지원
- **Webhook** 기능으로 외부 시스템과 연동 가능
- **조직/사용자/팀 권한 관리** 지원
- MySQL, PostgreSQL, SQLite 등 다양한 DB 지원

### 🧩 Feather에서의 Gitea 활용

Feather는 Gitea를 통해 다음 작업을 자동화합니다:

1. **템플릿 기반 프로젝트 생성**  
   - Gitea의 `GenerateRepo API`를 사용해 템플릿으로부터 신규 프로젝트를 생성합니다.
   - 프로젝트 메타데이터(이름, 설명, 조직 등)를 입력으로 받아 자동으로 Git 저장소 생성.

2. **Webhook 등록**  
   - `CreateHook API`를 통해 Gitea 저장소에 Webhook을 설정합니다.
   - `push` 이벤트가 발생하면 Argo Events로 전달되어 CI/CD가 트리거됨.

3. **GitOps 기반 CD 연동**  
   - Gitea 저장소가 ArgoCD `ApplicationSet`에 자동 등록되어 Git 변경 사항이 배포까지 연동됩니다.

> Feather는 Gitea를 단순한 Git 저장소를 넘어 **CI/CD 파이프라인의 시작점**으로 사용합니다.

---
### 🔗 관련 API 문서

- [GenerateRepo API](https://docs.gitea.com/en-us/api-reference/repository/generate-repo) – 템플릿 기반 저장소 생성
- [CreateHook API](https://docs.gitea.com/en-us/api-reference/repository/repo-create-hook) – Webhook 설정

---
## 📘 Feather가 사용하는 Argo 구성요소

Feather는 **Argo 프로젝트의 4가지 핵심 구성요소**를 통해 완전 자동화된 CI/CD 파이프라인을 구축합니다.

### 🔔 Argo Events
**이벤트 수신기 (Event Bus + EventSource + Sensor)**  
외부 이벤트(Gitea webhook 등)를 감지하여 후속 동작을 트리거합니다.

- `EventSource`: Gitea Webhook 등 외부 이벤트를 수신
- `Sensor`: 이벤트를 감지하고, 지정된 Argo Workflow 실행
- `EventBus`: 내부 이벤트 라우팅을 담당 (Kafka, NATS 등 지원)

> Feather는 Gitea의 `push` 이벤트를 감지하여 CI 워크플로우를 트리거합니다.

---

### ⚙️ Argo Workflows
**Kubernetes 네이티브 CI 파이프라인 엔진**

- DAG 기반 작업 정의
- Kaniko, Docker, Test, Deploy 등 다양한 작업 구성 가능
- 개별 Step 간 종속성, 조건부 실행 등 정교한 흐름 제어 지원

> Feather는 `build → test → image push` 단계를 Argo Workflow로 자동 수행합니다.

---

### 🧩 Argo Sensor
**Argo Events와 Workflows를 연결하는 트리거 구성 요소**

- EventSource로부터 이벤트를 수신
- 조건에 따라 Argo Workflow를 실행 (예: 특정 브랜치에 push 시)

> Feather에서는 Gitea Webhook으로부터 받은 이벤트를 기반으로 CI Workflow 실행을 담당합니다.

---

### 🚀 Argo CD (ApplicationSet 포함)
**GitOps 기반 Kubernetes 애플리케이션 배포 도구**

- Git 저장소의 manifest를 Kubernetes에 동기화
- 자동화된 배포, 롤백, 상태 모니터링 제공
- `ApplicationSet`을 사용하면 여러 애플리케이션을 동적으로 관리 가능

> Feather는 `ApplicationSet`을 통해 새로 생성된 프로젝트를 자동으로 ArgoCD에 등록하고 배포합니다.

