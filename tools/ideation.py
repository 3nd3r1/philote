#!/usr/bin/env python3
"""Validate the ideation items and generate docs/ideation/INDEX.md.

  python3 tools/ideation.py check             validate all items and that INDEX.md is current
  python3 tools/ideation.py index             regenerate INDEX.md
  python3 tools/ideation.py new KIND "Title"  create the next item of KIND

Standard library only (needs Python 3.11+ for tomllib).
"""

import datetime
import json
import re
import sys
import tomllib
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent / "docs" / "ideation"
INDEX = ROOT / "INDEX.md"
TITLE_MAX, SUMMARY_MAX = 80, 200

# Field specs: str, datetime.date, "X" = one id with prefix X,
# ["X"] = list of ids with prefix X, ["*"] = list of any ids.
KINDS = {
    "principle": dict(dir="principles", prefix="P",
                      statuses=["proposed", "accepted", "retired"],
                      fields={}, sections=[]),
    "question": dict(dir="questions", prefix="Q",
                     statuses=["open", "answered", "dropped"],
                     fields={"answered_by": "D"}, sections=[]),
    "decision": dict(dir="decisions", prefix="D",
                     statuses=["proposed", "accepted", "rejected", "superseded"],
                     fields={"answers": ["Q"], "supersedes": ["D"]},
                     sections=["Decision", "Why"]),
    "idea": dict(dir="ideas", prefix="I",
                 statuses=["raw", "exploring", "adopted", "rejected"],
                 fields={}, sections=[]),
    "experiment": dict(dir="experiments", prefix="E",
                       statuses=["planned", "running", "done", "abandoned"],
                       fields={"tests": ["*"]},
                       sections=["Steps", "Kill criteria"]),
    "prior-art": dict(dir="prior-art", prefix="A",
                      statuses=["noted", "studied"],
                      fields={"url": str}, sections=[]),
}
COMMON = {"id": str, "title": str, "status": str, "summary": str,
          "created": datetime.date, "related": ["*"]}
REQUIRED = ["id", "title", "status", "summary", "created"]
REVIEW_STATUSES = {"principle": "proposed", "decision": "proposed"}
TOP_LEVEL = {"thesis.md", "INDEX.md", "archive"} | {k["dir"] for k in KINDS.values()}
PREFIX_KIND = {k["prefix"]: name for name, k in KINDS.items()}


class Item:
    def __init__(self, kind, path, meta, body):
        self.kind, self.path, self.meta, self.body = kind, path, meta, body

    @property
    def id(self):
        return self.meta.get("id", "")

    def sections(self):
        """Map of '## Heading' -> text under it."""
        out, cur = {}, None
        for line in self.body.splitlines():
            m = re.match(r"^## (.+?)\s*$", line)
            if m:
                cur = m.group(1)
                out[cur] = ""
            elif cur is not None:
                out[cur] += line + "\n"
        return out


def rel(path):
    return path.relative_to(ROOT.parent.parent)


def load(errors):
    items = []
    for entry in sorted(ROOT.iterdir()):
        if entry.name not in TOP_LEVEL:
            errors.append(f"{rel(entry)}: unexpected file; items go in one of {sorted(k['dir'] for k in KINDS.values())}")
    if not (ROOT / "thesis.md").is_file():
        errors.append("docs/ideation/thesis.md is missing")
    for kind, spec in KINDS.items():
        d = ROOT / spec["dir"]
        if not d.is_dir():
            continue
        for path in sorted(d.iterdir()):
            if not re.fullmatch(spec["prefix"] + r"-\d{3}-[a-z0-9-]+\.md", path.name):
                errors.append(f"{rel(path)}: name must look like {spec['prefix']}-001-short-slug.md")
                continue
            text = path.read_text()
            if not text.startswith("+++\n") or "\n+++\n" not in text[3:]:
                errors.append(f"{rel(path)}: must start with TOML frontmatter between +++ lines")
                continue
            end = text.index("\n+++\n", 3)
            try:
                meta = tomllib.loads(text[4:end])
            except tomllib.TOMLDecodeError as e:
                errors.append(f"{rel(path)}: bad frontmatter: {e}")
                continue
            items.append(Item(kind, path, meta, text[end + 5:]))
    return items


def check_field(item, key, spec, value, ids, errors):
    where = f"{rel(item.path)}: '{key}'"
    if spec is str:
        if not isinstance(value, str) or not value.strip():
            errors.append(f"{where} must be a non-empty string")
    elif spec is datetime.date:
        if not isinstance(value, datetime.date) or isinstance(value, datetime.datetime):
            errors.append(f"{where} must be a date like 2026-11-09 (unquoted)")
    elif isinstance(spec, str):
        check_ref(where, spec, value, ids, errors)
    elif isinstance(spec, list):
        if not isinstance(value, list):
            errors.append(f"{where} must be a list of ids")
            return
        for v in value:
            check_ref(where, spec[0], v, ids, errors)


def check_ref(where, prefix, value, ids, errors):
    if not isinstance(value, str) or value not in ids:
        errors.append(f"{where} refers to unknown item {value!r}")
    elif prefix != "*" and not value.startswith(prefix + "-"):
        errors.append(f"{where} must refer to a {PREFIX_KIND[prefix]} ({prefix}-...), not {value}")


def validate(items, errors):
    ids = {}
    for it in items:
        if it.id in ids:
            errors.append(f"{rel(it.path)}: id {it.id} is also used by {rel(ids[it.id].path)}")
        ids[it.id] = it

    for it in items:
        spec = KINDS[it.kind]
        p = rel(it.path)
        allowed = {**COMMON, **spec["fields"]}
        for key in it.meta:
            if key not in allowed:
                errors.append(f"{p}: unknown field '{key}'; {it.kind} fields are {', '.join(allowed)}")
        for key in REQUIRED:
            if key not in it.meta:
                errors.append(f"{p}: missing '{key}'")
        for key, value in it.meta.items():
            if key in allowed:
                check_field(it, key, allowed[key], value, ids, errors)

        if it.id and not it.path.name.startswith(it.id + "-"):
            errors.append(f"{p}: id {it.id} doesn't match the file name")
        if it.meta.get("status") not in spec["statuses"]:
            errors.append(f"{p}: status must be one of {', '.join(spec['statuses'])}")
        if len(it.meta.get("title", "")) > TITLE_MAX:
            errors.append(f"{p}: title is over {TITLE_MAX} characters")
        if len(it.meta.get("summary", "")) > SUMMARY_MAX:
            errors.append(f"{p}: summary is {len(it.meta['summary'])} characters; the cap is {SUMMARY_MAX}")

        secs = it.sections()
        required = list(spec["sections"])
        if it.kind == "experiment" and it.meta.get("status") == "done":
            required.append("Result")
        for s in required:
            if not secs.get(s, "").strip():
                errors.append(f"{p}: needs a non-empty '## {s}' section")
        for key in ("title", "summary"):
            if str(it.meta.get(key, "")).strip() == "TODO":
                errors.append(f"{p}: '{key}' is still TODO")
        for s, text in secs.items():
            if text.strip() == "TODO":
                errors.append(f"{p}: section '## {s}' is still TODO")

    # Questions and decisions must agree with each other.
    for it in items:
        st, p = it.meta.get("status"), rel(it.path)
        if it.kind == "question":
            by = it.meta.get("answered_by")
            if st == "answered" and not by:
                errors.append(f"{p}: answered questions need 'answered_by'")
            if st != "answered" and by:
                errors.append(f"{p}: 'answered_by' is set but status is {st}")
            if st == "answered" and by in ids:
                d = ids[by]
                if d.meta.get("status") != "accepted":
                    errors.append(f"{p}: answered by {by}, which isn't accepted")
                if it.id not in d.meta.get("answers", []):
                    errors.append(f"{p}: answered by {by}, but {by} doesn't list it in 'answers'")
        if it.kind == "decision" and st == "accepted":
            for q in it.meta.get("answers", []):
                if q in ids and ids[q].meta.get("answered_by") != it.id:
                    errors.append(f"{p}: accepted and answers {q}, but {q} isn't marked answered_by {it.id}")
        if it.kind == "decision" and st == "superseded":
            if not any(it.id in o.meta.get("supersedes", []) for o in items):
                errors.append(f"{p}: superseded, but no decision lists it in 'supersedes'")


def cell(s):
    return str(s).replace("|", "\\|").replace("\n", " ")


def render_index(items):
    by_kind = {k: sorted((i for i in items if i.kind == k), key=lambda i: i.id) for k in KINDS}
    out = [
        "# Ideation index",
        "",
        "Generated by `python3 tools/ideation.py index`. Don't edit this file; edit the items.",
        "",
        "Start with the [thesis](thesis.md).",
        "",
    ]

    def link(i):
        return f"[{i.id}]({i.path.relative_to(ROOT).as_posix()})"

    review = [i for i in items if REVIEW_STATUSES.get(i.kind) == i.meta.get("status")]
    review.sort(key=lambda i: i.id)
    out += ["## Waiting for review", ""]
    if review:
        out += ["Proposed but not yet accepted by a human.", "",
                "| ID | Title | Summary |", "| --- | --- | --- |"]
        out += [f"| {link(i)} | {cell(i.meta['title'])} | {cell(i.meta['summary'])} |" for i in review]
    else:
        out.append("Nothing.")
    out.append("")

    headings = {"principle": "Principles", "question": "Questions", "decision": "Decisions",
                "idea": "Ideas", "experiment": "Experiments", "prior-art": "Prior art"}
    for kind, heading in headings.items():
        rows = by_kind[kind]
        counts = {}
        for i in rows:
            counts[i.meta["status"]] = counts.get(i.meta["status"], 0) + 1
        tally = ", ".join(f"{n} {s}" for s, n in sorted(counts.items()))
        out += [f"## {heading}", ""]
        if not rows:
            out += ["None yet.", ""]
            continue
        out += [tally, "", "| ID | Status | Title | Summary |", "| --- | --- | --- | --- |"]
        out += [f"| {link(i)} | {i.meta['status']} | {cell(i.meta['title'])} | {cell(i.meta['summary'])} |"
                for i in rows]
        out.append("")
    return "\n".join(out)


def run_checks():
    errors = []
    items = load(errors)
    validate(items, errors)
    return items, errors


def cmd_check():
    items, errors = run_checks()
    if not errors and (not INDEX.exists() or INDEX.read_text() != render_index(items)):
        errors.append("docs/ideation/INDEX.md is stale; run: python3 tools/ideation.py index")
    for e in errors:
        print(e, file=sys.stderr)
    if errors:
        sys.exit(1)
    print(f"ok: {len(items)} items")


def cmd_index():
    items, errors = run_checks()
    if errors:
        for e in errors:
            print(e, file=sys.stderr)
        sys.exit(1)
    INDEX.write_text(render_index(items))
    print(f"wrote {rel(INDEX)} ({len(items)} items)")


def cmd_new(kind, title):
    if kind not in KINDS:
        sys.exit(f"unknown kind {kind!r}; kinds: {', '.join(KINDS)}")
    spec = KINDS[kind]
    d = ROOT / spec["dir"]
    d.mkdir(exist_ok=True)
    nums = [int(m.group(1)) for f in d.iterdir() if (m := re.match(spec["prefix"] + r"-(\d{3})-", f.name))]
    ident = f"{spec['prefix']}-{max(nums, default=0) + 1:03d}"
    slug = re.sub(r"[^a-z0-9]+", "-", title.lower()).strip("-")[:40].rstrip("-")
    path = d / f"{ident}-{slug}.md"
    body = "".join(f"\n## {s}\n\nTODO\n" for s in spec["sections"])
    path.write_text(
        f'+++\nid = "{ident}"\ntitle = {json.dumps(title)}\nstatus = "{spec["statuses"][0]}"\n'
        f'summary = "TODO"\ncreated = {datetime.date.today().isoformat()}\n+++\n{body}')
    print(rel(path))


def main(argv):
    if argv[:1] == ["check"] and len(argv) == 1:
        cmd_check()
    elif argv[:1] == ["index"] and len(argv) == 1:
        cmd_index()
    elif argv[:1] == ["new"] and len(argv) == 3:
        cmd_new(argv[1], argv[2])
    else:
        sys.exit(__doc__)


if __name__ == "__main__":
    main(sys.argv[1:])
