package nextjs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckBodies(t *testing.T) {
	bodies := map[string]string{
		"card.tsx":    "    const n = 1;\n    return (\n      <Card className=\"h-full\">{n}</Card>\n    );",
		"bare.tsx":    "    return <Card>\n      x\n    </Card>;",
		"wrapped.tsx": "    return (\n      <div className=\"px-4\">\n        <Card />\n      </div>\n    );",
		"nested.tsx":  "    return (\n      <div>{[1].map((n) => {\n        return <Card key={n} />;\n      })}</div>\n    );",
		"early.tsx":   "    if (!x) {\n        return null;\n    }\n    return <CardHeader />;",
	}
	dir := t.TempDir()
	common := filepath.Join(dir, "src/components/common")
	if err := os.MkdirAll(common, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range bodies {
		src := "function XBody({ x }: XProps): React.ReactNode {\n" + body + "\n}\n"
		if err := os.WriteFile(filepath.Join(common, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	actions := "export default function XActions({ id }: { id: string }) {\n    return <div />;\n}\n"
	if err := os.WriteFile(filepath.Join(common, "x-actions.tsx"), []byte(actions), 0o644); err != nil {
		t.Fatal(err)
	}
	problems := map[string][]string{}
	if err := checkBodies(dir, problems); err != nil {
		t.Fatal(err)
	}
	for name := range bodies {
		file := "src/components/common/" + name
		bad := name == "wrapped.tsx" || name == "nested.tsx" || name == "early.tsx"
		if got := len(problems[file]) > 0; got != bad {
			t.Errorf("%s: problem reported = %v, want %v (%v)", name, got, bad, problems[file])
		}
	}
	if len(problems) != 3 {
		t.Errorf("want problems for exactly three files, got %v", problems)
	}
}

func TestCheckLayouts(t *testing.T) {
	layouts := map[string]string{
		"horses":     "    return <div>{content(\"grid gap-4\")}</div>;",
		"matches":    "    return <h1>{title}</h1>;",
		"horses/new": "    return <div>{content(\"a\")}{content(\"b\")}</div>;",
	}
	dir := t.TempDir()
	for route, body := range layouts {
		page := filepath.Join(dir, "src/app", route, "page.tsx")
		if err := os.MkdirAll(filepath.Dir(page), 0o755); err != nil {
			t.Fatal(err)
		}
		src := "function PageBody({ title, content }: PageBodyProps): React.ReactNode {\n" + body + "\n}\n"
		if err := os.WriteFile(page, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	problems := map[string][]string{}
	if err := checkLayouts(dir, problems); err != nil {
		t.Fatal(err)
	}
	if len(problems) != 2 || problems["src/app/matches/page.tsx"] == nil || problems["src/app/horses/new/page.tsx"] == nil {
		t.Errorf("want problems for the layouts that call content zero or two times, got %v", problems)
	}
}
