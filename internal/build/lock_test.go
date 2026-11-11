package build

import (
	"strings"
	"testing"

	"github.com/3nd3r1/philote/internal/diag"
)

func TestLockRoundTrip(t *testing.T) {
	pins := map[string]pin{
		"api/b": {Kind: "step", Node: "api/b", Key: "k2", Model: "m", Body: "if (x) {\n  return;\n}\n\nreturn y;"},
		"api/a": {Kind: "step", Node: "api/a", Key: "k1", Model: "m", Body: "return 1;"},
		"web/C": {Kind: "component", Node: "web/C", Key: "k3", Model: "m", Body: "return <p>{x.name}</p>;"},
	}
	text := string(printLock(pins))
	if strings.Index(text, "step api/a") > strings.Index(text, "step api/b") {
		t.Errorf("pins are not sorted by node:\n%s", text)
	}
	if strings.Contains(text, " \n") {
		t.Errorf("the lock has trailing spaces:\n%s", text)
	}
	d := &diag.List{}
	got := parseLock("philote.lock", text, d)
	if d.HasErrors() || len(got) != 3 || got["api/a"] != pins["api/a"] || got["api/b"] != pins["api/b"] || got["web/C"] != pins["web/C"] {
		t.Errorf("round trip changed the pins:\n%s\ngot %+v", text, got)
	}
}

func TestLockBadLine(t *testing.T) {
	d := &diag.List{}
	parseLock("philote.lock", "step api/a\nkey k\nmodel m\nreturn 1;\n", d)
	if len(d.Items) != 1 || d.Items[0].Pos.Line != 4 {
		t.Errorf("want one error on line 4, got %+v", d.Items)
	}
}

func TestBodyOf(t *testing.T) {
	reply := "Here it is:\n```ts\n    if (x) {\n      return;\n    }  \n\n    return y;\n```\n"
	want := "if (x) {\n  return;\n}\n\nreturn y;"
	if got := bodyOf(reply); got != want {
		t.Errorf("bodyOf = %q, want %q", got, want)
	}
}
