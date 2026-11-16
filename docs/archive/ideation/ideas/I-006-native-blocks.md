+++
id = "I-006"
title = "Native blocks as the escape hatch"
status = "exploring"
summary = "When the language can't express something, write a handwritten file with a declared, typed interface, like inline assembly or extern in C."
created = 2026-11-09
related = ["P-002"]
+++

```
native component Chart(points: [Point]) from "./native/Chart.tsx"
```

- A native block is a handwritten file with a declared, typed interface.
- The compiler never generates or touches it. It only checks that every use matches the declared interface.
- If you want to edit a generated node, promote it to native instead.

The ratio of native to generated nodes is a health metric: if most of an app ends up native, the language is missing something.
