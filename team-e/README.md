# 🚀 Feather CI/CD Auto-Registration

> **Seamlessly connect your GitHub repositories with Argo Workflows and ArgoCD for automated deployments**

![build](https://github.com/user-attachments/assets/ce2e1094-c89e-4c49-8a1c-1c16741dedee)

---

## 🌐 Environment

| Category        | Tech Stack                                                                 |
|----------------|------------------------------------------------------------------------------|
| **Language**    | Go, JavaScript                                                              |
| **Framework**   | Gin (Go), React (JS)                                                        |
| **IDE**         | Cursor, GoLand                                                              |
| **Dev Env**     | Kubernetes, Docker, Gitea, Argo Workflows, ArgoCD, Dive                    |
| **Service Mesh**| Istio                                                                       |
| **Observability**| OpenTelemetry, Loki, Prometheus, Tempo, Kiali                             |
| **AI**          | ChatGPT                                                                     |

---

## 🧩 Summary

Feather streamlines the CI/CD pipeline setup by **auto-registering repositories** created via Gitea templates into **Argo Workflows** and **ArgoCD**, enabling **hands-free, automated deployments** from the moment a project is scaffolded.

---

## 🛠️ Features

- 🔨 **Create repositories from templates** via the Gitea API
- ⚙️ **Auto-configure Argo Workflows** for CI pipelines
- 🚀 **Auto-sync with ArgoCD** for GitOps-based CD
- 🔍 **Fully observable** with Prometheus, Tempo, Loki, and Kiali

---

## 📦 Create Repository Using a Template

Feather leverages the [Gitea GenerateRepo API](https://docs.gitea.com/en-us/api-reference/repository/generate-repo) to scaffold new repositories from templates with customizable options.

### 📁 Gitea API: `GenerateRepoOption`
**Create a repository using a template**  
[🔗 API 문서 보기](https://demo.gitea.com/api/swagger#/repository/generateRepo)

|필드|타입|설명|
|---|---|---|
|`avatar`|boolean|템플릿 저장소의 아바타 포함 여부|
|`default_branch`|string|새 저장소의 기본 브랜치 이름|
|`description`|string|새로 만들 저장소의 설명|
|`git_content`|boolean|템플릿 저장소의 기본 브랜치의 Git 콘텐츠 포함 여부|
|`git_hooks`|boolean|Git hook 포함 여부|
|`labels`|boolean|템플릿 저장소의 라벨 포함 여부|
|`name`*|string|새로 만들 저장소의 이름|
|`owner`*|string|새 저장소의 소유자(사용자명 또는 조직명)|
|`private`|boolean|저장소를 비공개로 만들지 여부|
|`protected_branch`|boolean|보호된 브랜치 포함 여부|
|`topics`|boolean|토픽 포함 여부|
|`webhooks`|boolean|웹훅 포함 여부|

---
### 🔔 Gitea API: `CreateHookOption`
**Create a webhook**  
[🔗 API 문서 보기](https://demo.gitea.com/api/swagger#/repository/repoCreateHook)

|필드|타입|설명|
|---|---|---|
|`active`|boolean (기본값: false)|훅이 활성화 상태인지 여부|
|`authorization_header`|string|Authorization 헤더 값|
|`branch_filter`|string|특정 브랜치에만 적용될 필터|
|`config`*|`CreateHookOptionConfig`|웹훅 설정 구성|
|`events`|array of string|트리거할 이벤트 목록|
|`type`*|string (enum)|웹훅 타입 (`gitea`, `slack`, 등)|
#### 🔧 CreateHookOptionConfig
웹훅 설정을 위한 `config` 필드는 웹훅 타입에 따라 다른 키를 포함합니다.
**Gitea 웹훅**:

|키|값 예시|
|---|---|
|`url`|`https://example.com`|
|`content_type`|`json` 또는 `form`|


---

## ⚙️ Argo Workflows

Feather registers newly created repositories to **Argo Workflows**, enabling CI pipelines to be executed automatically. This includes:

- 🧪 Test automation
- 🛠️ Build steps
- ✅ Status reporting

---

## 🔁 ArgoCD

Repositories are also linked with **ArgoCD**, providing:

- 📦 Continuous Deployment through GitOps
- 🔄 Automatic sync of application manifests
- 🔒 Declarative security and rollback support

---

## 📸 Preview

Coming soon...

---

## 🤝 Contributing

Contributions are welcome! Please open issues or submit PRs to improve Feather or its integrations.

---

## 📄 License

[MIT License](./LICENSE)

---

*Made with ❤️ by developers for developers*
