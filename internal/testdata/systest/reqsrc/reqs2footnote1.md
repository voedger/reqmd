---
reqmd.package: reqs2footnote1
---

# Two sites one footnote

`~func1~`uncvrd[^1]❓
`~func2~`
> replace `~func2~`uncvrd[^2]❓

Delete the last (empty) line, add a footnote and append an empty line:
> deletelast
> append [^2]: `[~reqs2footnote1/func2~impl]`
> append

[^1]: `[~reqs2footnote1/func1~impl]`
