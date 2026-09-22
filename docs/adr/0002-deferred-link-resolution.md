# Deferred Link Resolution for Cross-File Calls

We decided to use single-pass AST parsing that emits unresolved `Raw Call` records alongside declarations, followed by a corpus-wide link rewiring phase. This avoids re-parsing ASTs across multiple files, allows file parsing to execute concurrently across worker goroutines, and isolates symbol resolution into a dedicated pipeline.
