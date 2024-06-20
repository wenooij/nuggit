# `nuggit`

## Commands

```
:c[n]    op key       ; create node
:ce      key,src,dst  ; create edge
:cg      key          ; create graph
:cx      key,n1,n2    ; create exchange
:d[n][!] glob         ; delete node [force]
:dg[!]   glob         ; delete graph [force]
:dx[!]   glob         ; delete exchange [force]
:e[g][!] key          ; edit graph [:cg]
:f[g]    key,url      ; fetch graph
:h       [topic]      ; help [topic]
:q[!]                 ; quit [force]
:o[g]    key,filename ; open graph
:r[n][a] [glob]       ; read node [all]
:rg[a]   [glob]       ; read graph [all]
:rx      [glob]       ; read exchange
:u[!]    key,path,val ; update node [:c]
:ue[!]   key,src,dst  ; update edge [:ce]
:ug[!]   path,val     ; updage graph [:cg]
:ux[!]   key,n1,n2    ; update exchange [:cx]
:w[q][!] file         ; write [quit] [force]
:x[n]    glob         ; diff node
:xg[a]   glob         ; diff graph [all]
:xx      glob         ; diff exchange
```