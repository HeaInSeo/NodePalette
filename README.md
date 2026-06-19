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
make test
make coverage-check
```

`coverage-check`는 전체 statement coverage가 70% 미만이면 실패한다.

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
| `lint` | `.golangci.yml` 기반 golangci-lint (gosec/bodyclose/noctx 포함) |
| `build` | go build + go vet |
| `test-unit` | `go test -race -covermode=atomic`, 전체 coverage 70% 미만 실패 |
| `k8s-contract` | bori-facing K8s 데이터 플레인 매니페스트 계약 테스트 |
| `vuln-scan` | govulncheck (continue-on-error) |

---

## K8s 데이터 플레인 계약

NodePalette는 향후 `bori`가 배포를 오케스트레이션할 read-only 데이터 플레인 앱이다.
이 repo의 `deploy/` 매니페스트는 직접 배포 절차가 아니라 **bori가 소비해야 할 기준 계약**이다.

| 파일 | 계약 |
|------|------|
| `deploy/00-namespace.yaml` | `nodepalette-system` 네임스페이스 |
| `deploy/01-nodepalette.yaml` | HTTP 8083, `/healthz` readiness/liveness, NodeVault catalog REST 주소 |

운영 안전 조건:

- NodePalette는 `PromotionStatus=active` 항목만 외부에 노출한다. 단건 조회도 inactive 항목은 `404`로 숨긴다.
- 컨테이너는 `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true`, `capabilities.drop: [ALL]`을 유지한다.
- Replica 기본값은 2다. 서비스는 상태를 갖지 않고 NodeVault REST를 source of truth로 사용한다.
- 이미지 태그는 `latest`를 쓰지 않는다. bori 릴리스 파이프라인에서 명시 태그/다이제스트로 치환한다.
- CI의 `k8s-contract` job은 bori-facing 매니페스트가 위 계약을 깨지 않는지 검사한다.

로컬 품질 게이트:

```bash
make lint
make test
make coverage-check
go test -v ./test/k8s/...
```

테스트 정책:

- fail path: NodeVault REST 4xx/5xx, invalid JSON, broken body, context cancellation, upstream outage를 테스트한다.
- regression: base URL 트레일링 슬래시, 기본 in-cluster NodeVault 주소, inactive tool 비노출, `/healthz` method contract를 고정한다.
- 실제 클러스터 배포/rollout은 bori 트랙에서 담당하고, 이 repo는 bori가 소비할 앱 계약을 검증한다.

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
