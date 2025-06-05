sooner rather than later
- [x] Write common headers on every response no matter what handler handles it (write a handler for this)
- [ ] write cache headers on file handler responses
- [ ] Move logic from main into some class and break it up into logical chunks
- [ ] brotli and gzip compression handlers (minimal library support? does stdlib support it?)
- [ ] correctly write mime on file handler with `file --mime-type` command
- [ ] support `Connection: keep-alive`

later:
- [ ] implement delete on radix tree
- [ ] something something custom handlers? perhaps in other languages? look for standards
- [ ] something something register variable paths? (macro declarations???)