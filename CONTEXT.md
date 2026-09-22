# Graphify-Go

A high-performance, deterministic codebase knowledge graph engine and query tool implemented in Go.

## Language

**Symbol Node**:
A declared code entity (function, method, struct, interface, or package) with a stable URI identifier and source location.
_Avoid_: Entity, token, code item

**Extracted Edge**:
A relationship directly verified by the AST parser within a file.
_Avoid_: Direct link, static edge

**Inferred Edge**:
A relationship resolved across files or packages using import paths and symbol tables.
_Avoid_: Guessed edge, dynamic link

**God Node**:
A symbol with abnormally high in-degree or out-degree centrality, indicating an architectural hub.
_Avoid_: Super node, hot spot, bottleneck

**Blast Radius**:
The transitive set of upstream symbols that depend on or call a given symbol or file.
_Avoid_: Impact surface, ripple effect, affected set

**Raw Call**:
An unresolved call expression captured during AST parsing containing caller ID, callee expression, and receiver context.
_Avoid_: Call site, unlinked call

**Community**:
A cohesive cluster of symbols identified by graph modularity optimization representing an architectural subsystem.
_Avoid_: Module cluster, group, partition

**Manifest**:
A lightweight metadata store recording file paths, mtimes, and content hashes to support incremental graph synchronization.
_Avoid_: File cache, state file, index
