# Ideas

Collection of ideas and random thoughts.

## Terminal output

Maybe something like this:

```bash
┌+ bootstrap ─────────────────────── base.Multi{6} ─ 21:01:01.000 ─
│  check did not pass
│ ┌+ tree ────────────────────── apt.Package{tree} ─ 21:01:01.000 ─
│ │ Check passed
│ └─ tree ─ OK ───────────────────────────────────── 21:01:01.000 ─
│ System is Ubuntu
│ ┌+ something ───────────────── apt.Package{tree} ─ 21:01:01.000 ─
│ │ ┌+ something ─────────────── apt.Package{tree} ─ 21:01:01.000 ─
│ │ │
│ │ └─ something ─ Changed ───────────────────────── 21:01:01.000 ─
│ └─ something ─ Changed ─────────────────────────── 21:01:01.000 ─
└─ Bootstrap ─ Changed ───────────────────────────── 21:01:01.000 ─
```
