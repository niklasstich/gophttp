sooner rather than later
- [x] Introduce a handler collection on every path where we can register handlers for different HTTP methods
- [ ] Real logging with loglevels
- [x] Write common headers on every response no matter what handler handles it (write a handler for this)
- [x] Move logic from main into some class and break it up into logical chunks
- [x] correctly write mime on file handler with `file --mime-type` command
- [x] brotli and gzip compression handlers (minimal library support? does stdlib support it?)
  - [ ] cache compressed static content
  - [ ] optional flag to precompress static routes 
  - [ ] replace brotli package with my own brotli implementation
- [ ] support `Connection: keep-alive`
- [ ] write cache headers on file handler responses
- [ ] CORS headers

later:
- [ ] implement delete on radix tree
- [ ] something something custom (dynamic) handlers? perhaps in other languages? look for standards (CGI etc)
- [ ] something something register variable paths? (macro declarations???)