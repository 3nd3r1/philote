package nextjs

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCheckBodies(t *testing.T) {
	const says = "const says = [\"Find your herd.\", \"Say \\\"hi\\\"]\"] as const;\n"
	const linked = "export default function X(props: XProps) {\n    return <XBody {...props} link={{ href: \"/a\" }} actions={<XActions id={props.x.id} />} />;\n}\n"
	files := map[string]struct {
		head, body string
		want       []string
	}{
		"says.tsx":    {says, "return <h1>{says[0]}{says[1]}</h1>;", nil},
		"missing.tsx": {says, "return <h1>{says[0]}</h1>;", []string{"the body must render says[1]"}},
		"linked.tsx":  {linked, "return <Link href={link.href}>{actions}</Link>;", nil},
		"unlinked.tsx": {linked, "return <div>{actions}{actions}</div>;", []string{
			"the body must link to link.href", "the body must render {actions} exactly once",
		}},
	}
	dir := t.TempDir()
	common := filepath.Join(dir, "src/components/common")
	if err := os.MkdirAll(common, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, f := range files {
		src := f.head + "function XBody({ x }: XProps): React.ReactNode {\n    " + f.body + "\n}\n"
		if err := os.WriteFile(filepath.Join(common, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	problems := map[string][]string{}
	if err := checkBodies(dir, problems); err != nil {
		t.Fatal(err)
	}
	for name, f := range files {
		if got := problems["src/components/common/"+name]; !slices.Equal(got, f.want) {
			t.Errorf("%s: got %q, want %q", name, got, f.want)
		}
	}
}

// The nav's body has to use everything its wrapper hands it.
func TestCheckNav(t *testing.T) {
	const head = "export default function MainNavbar() {\n    return <NavBody items={nav} active={active} home={home} toggle={<ModeToggle />} />;\n}\n"
	cases := []struct {
		body string
		want []string
	}{
		{"return <nav><Link href={home.href}>{home.label}</Link>{items.map((i) => <Link key={i.href} href={i.href} aria-current={i.href === active ? \"page\" : undefined}>{i.label}</Link>)}{toggle}</nav>;", nil},
		{"return <nav>{items.map((i) => i.label)}</nav>;", []string{
			"the nav must use active", "the nav must use home.href", "the nav must use home.label", "the nav must render {toggle} exactly once",
		}},
	}
	for _, tc := range cases {
		dir := t.TempDir()
		path := filepath.Join(dir, navFile)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		src := head + "function NavBody({ items, active, home, toggle }: Props): React.ReactNode {\n    " + tc.body + "\n}\n"
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		problems := map[string][]string{}
		if err := checkBodies(dir, problems); err != nil {
			t.Fatal(err)
		}
		if got := problems[navFile]; !slices.Equal(got, tc.want) {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}
