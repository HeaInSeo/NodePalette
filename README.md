# NodePalette

NodeVault의 `GET /v1/catalog/certified-tools`를 조회해 `PromotionStatus=active`인 인증 tool 목록을 pipeline 빌더(DagEdit 등)에게 노출하는 **read-only REST 서비스**.

NodePalette는 인증 결정을 내리지 않는다. 데이터 소유와 PromotionStatus 관리는 NodeVault `certification.Service` 책임이다.

→ 전체 플랫폼 구성 및 end-to-end 흐름: [NodeVault README](https://github.com/HeaInSeo/NodeVault/blob/main/README.md)

---

## 아키텍처 경계

NodePalette는 **read-only proxy**다.

- 인증 결정: NodeVault `certification.Service` 책임
- 데이터 소유(CertifiedToolImageRecord, ToolFunctionCatalogEntry): NodeVault 책임
- NodePalette는 `PromotionStatus=active` 항목만 필터링하여 반환하고 상태를 변경하지 않는다.
- NodeVault index나 Harbor를 직접 조회하지 않는다.

---

## 전체 구조

```
DagEdit (pipeline 빌더) / 기타 소비자
    │  GET /v1/palette/tools
    ▼
NodePalette (이 프로젝트 — read-only REST 서비스)
    │
    ├── pkg/server         — HTTP 서버: /v1/palette/tools, /healthz
    └── pkg/paletteclient  — NodeVault REST 클라이언트: GET /v1/catalog/certified-tools
    │
    ▼ GET /v1/catalog/certified-tools
NodeVault
    └── CertifiedToolImageRecord + ToolFunctionCatalogEntry (PromotionStatus=active)
```

---

## 엔드포인트

| 엔드포인트 | 설명 |
|------------|------|
| `GET /v1/palette/tools` | `PromotionStatus=active` 인증 tool 전체 목록 |
| `GET /v1/palette/tools/{cas_hash}` | `cas_hash` 기준 단건 조회 |
| `GET /healthz` | 헬스체크 → `200 ok` |

`cas_hash`는 비어 있거나 경로 순회 패턴(`.`, `..`)이 포함된 경우 `400 Bad Request`를 반환한다.

---

## 환경 변수

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `NODEVAULT_API_ADDR` | `http://nodevault.nodevault-system.svc:8082` | NodeVault REST 주소 (certified-tools 조회) |
| `NODEPALETTE_ADDR` | `:8083` | NodePalette HTTP 서버 바인딩 주소 |

---

## 빌드 및 실행

### 사전 조건

| 도구 | 용도 |
|------|------|
| Go 1.25.5 이상 | 빌드 |

### 빌드

```bash
go build ./...
```

### 테스트

```bash
go test -race ./...
```

### 실행 (로컬 디버깅)

```bash
NODEVAULT_API_ADDR=http://localhost:8082 \
NODEPALETTE_ADDR=:8083 \
go run ./cmd/nodepalette
```

헬스체크:

```bash
curl http://localhost:8083/healthz
# → ok

curl http://localhost:8083/v1/palette/tools
# → [...CertifiedToolImageRecord 목록...]
```

---

## CI (GitHub Actions)

`.github/workflows/ci.yml` 구성:

| Job | 내용 |
|-----|------|
| `lint` | golangci-lint (zero-warning) |
| `build` | go build + go vet |
| `test` | -race -cover |
| `vuln-scan` | govulncheck (continue-on-error) |

---

## 패키지 구조

```
cmd/nodepalette/     — HTTP 서버 진입점 (main.go): graceful shutdown 포함
pkg/
  paletteclient/     — NodeVault REST 클라이언트 (GET /v1/catalog/certified-tools)
  server/            — HTTP 핸들러: /v1/palette/tools, /v1/palette/tools/{cas_hash}, /healthz
```

---

## 관련 프로젝트

| 프로젝트 | 역할 |
|----------|------|
| [`NodeVault`](https://github.com/HeaInSeo/NodeVault) | 인증 tool 데이터 소유 — certification.Service, GET /v1/catalog/certified-tools |
| [`NodeSentinel`](https://github.com/HeaInSeo/NodeSentinel) | K8s 데이터플레인 검증 에이전트 — L3/L4/L5-a/L5-b |
| [`NodeKit`](https://github.com/HeaInSeo/NodeKit) | C# 어드민 UI — ToolDefinition 편집 및 BuildRequest gRPC 전송 |
| [`DockGuard`](https://github.com/HeaInSeo/DockGuard) | OPA/Rego Dockerfile 정책 + .wasm 번들 빌드 |
