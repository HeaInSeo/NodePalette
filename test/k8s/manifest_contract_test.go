package k8s_test

import (
	"os"
	"strings"
	"testing"
)

func readManifest(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest %s: %v", path, err)
	}
	return string(b)
}

func requireContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("manifest missing %q", want)
	}
}

func TestNodePaletteDeploymentContract(t *testing.T) {
	body := readManifest(t, "../../deploy/01-nodepalette.yaml")

	for _, want := range []string{
		"kind: Deployment",
		"name: nodepalette",
		"namespace: nodepalette-system",
		"replicas: 2",
		"name: NODEPALETTE_ADDR",
		"value: \":8083\"",
		"name: NODEVAULT_API_ADDR",
		"http://nodevault-controlplane.nodevault-system.svc:8082",
		"path: /healthz",
		"readinessProbe:",
		"livenessProbe:",
		"allowPrivilegeEscalation: false",
		"readOnlyRootFilesystem: true",
		"drop:",
		"- ALL",
		"kind: Service",
		"targetPort: http",
	} {
		requireContains(t, body, want)
	}

	if strings.Contains(body, ":latest") {
		t.Fatal("bori-facing manifest must not pin an implicit latest image")
	}
}

func TestNodePaletteNamespaceContract(t *testing.T) {
	body := readManifest(t, "../../deploy/00-namespace.yaml")

	for _, want := range []string{
		"kind: Namespace",
		"name: nodepalette-system",
		"app.kubernetes.io/name: nodepalette",
	} {
		requireContains(t, body, want)
	}
}
