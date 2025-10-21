# System Architecture

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                          USER INTERACTION                            │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    │             │             │
                    ▼             ▼             ▼
        ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
        │   validate   │  │     tag      │  │   extract    │
        │     CLI      │  │     CLI      │  │     CLI      │
        │      ✅       │  │     🚧       │  │     🚧       │
        └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
               │                 │                 │
               │                 │                 │
┌──────────────┼─────────────────┼─────────────────┼──────────────────┐
│              │   APPLICATION LAYER                │                  │
│              ▼                 ▼                 ▼                  │
│   ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │
│   │   Validator     │  │   Tag Writer    │  │    Extractor    │   │
│   │   Orchestrator  │  │   Orchestrator  │  │   Orchestrator  │   │
│   └────────┬────────┘  └────────┬────────┘  └────────┬────────┘   │
│            │                    │                    │             │
└────────────┼────────────────────┼────────────────────┼─────────────┘
             │                    │                    │
             │                    │                    │
┌────────────┼────────────────────┼────────────────────┼─────────────┐
│            │    INFRASTRUCTURE LAYER                 │             │
│            ▼                    ▼                    ▼             │
│   ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │
│   │   validation/   │  │    tagging/     │  │   scraping/     │   │
│   │  AlbumValidator │  │   FLACReader    │  │   Extractors    │   │
│   │   ✅ Complete    │  │  ✅ Read only    │  │    🚧 TODO      │   │
│   └────────┬────────┘  │   FLACWriter    │  └────────┬────────┘   │
│            │           │    🚧 TODO       │           │            │
│   ┌────────▼────────┐  └────────┬────────┘  ┌────────▼────────┐   │
│   │  filesystem/    │           │           │   storage/      │   │
│   │DirValidator     │◄──────────┼───────────┤  Repository     │   │
│   │  ✅ Complete     │           │           │  ✅ Complete     │   │
│   └─────────────────┘           │           └─────────────────┘   │
│                                 │                                 │
└─────────────────────────────────┼─────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         DOMAIN LAYER                                 │
│                                                                      │
│   ┌─────────────────────────────────────────────────────────────┐   │
│   │                    domain/                                   │   │
│   │                                                              │   │
│   │  Value Objects          Entities          Aggregate Root   │   │
│   │  ┌─────────────┐       ┌─────────────┐   ┌─────────────┐   │   │
│   │  │   Level     │       │    Track    │   │    Album    │   │   │
│   │  │   Role      │       │   Edition   │   │             │   │   │
│   │  │   Artist    │       └─────────────┘   └─────────────┘   │   │
│   │  │ ValidationI │                                           │   │
│   │  │   ssue      │                                           │   │
│   │  └─────────────┘                                           │   │
│   │                                                              │   │
│   │                    ✅ Complete & Tested                      │   │
│   └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

## Data Flow Diagrams

### 1. Validation Flow (Currently Working ✅)

```
┌──────────────┐
│ User runs:   │
│ validate     │
│ /path/album  │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────┐
│ DirectoryScanner             │
│ - Walk file tree             │
│ - Find FLAC files            │
│ - Detect multi-disc          │
└──────┬───────────────────────┘
       │
       ├─────────────────┐
       ▼                 ▼
┌─────────────────┐   ┌──────────────────┐
│ FLACReader      │   │ DirectoryValidator│
│ - Read tags     │   │ - Check paths     │
│ - Parse tracks  │   │ - Check structure │
└────────┬────────┘   └────────┬──────────┘
         │                     │
         └──────────┬──────────┘
                    ▼
          ┌──────────────────┐
          │ Build Album      │
          │ from Tracks      │
          └────────┬──────────┘
                   │
                   ▼
          ┌──────────────────┐
          │ AlbumValidator   │
          │ - Check metadata │
          │ - Check rules    │
          └────────┬──────────┘
                   │
                   ▼
          ┌──────────────────┐
          │ ValidationReport │
          │ - Structure      │
          │ - Metadata       │
          │ - Errors         │
          └────────┬──────────┘
                   │
                   ▼
          ┌──────────────────┐
          │ Print Report     │
          │ ❌ Errors        │
          │ ⚠️  Warnings     │
          │ ℹ️  Info         │
          └──────────────────┘
```

### 2. Tagging Flow (Coming Soon 🚧)

```
┌──────────────┐
│ User runs:   │
│ tag --apply  │
│ metadata.json│
└──────┬───────┘
       │
       ▼
┌──────────────────────────────┐
│ Repository                   │
│ LoadFromJSON(metadata.json)  │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ Validate Album               │
│ (ensure valid before write)  │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ DirectoryScanner             │
│ Find all FLAC files          │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ For each file:               │
│ 1. Backup original           │
│ 2. FLACWriter.WriteTag()     │
│ 3. Verify write success      │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ Report                       │
│ - Files updated: N           │
│ - Errors: M                  │
│ - Backups: /path             │
└──────────────────────────────┘
```

### 3. Extraction Flow (Coming Soon 🚧)

```
┌──────────────┐
│ User runs:   │
│ extract --url│
│ https://...  │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────┐
│ Detect Site                  │
│ (harmoniamund, naxos, etc.)  │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ Site-specific Extractor      │
│ - Fetch HTML                 │
│ - Parse structure            │
│ - Extract metadata           │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ Build Album                  │
│ from extracted data          │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ Validate Album               │
│ (ensure completeness)        │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ Repository                   │
│ SaveToJSON(metadata.json)    │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ Output metadata.json         │
│ Ready for tag --apply        │
└──────────────────────────────┘
```

## Layer Dependencies

```
┌─────────────────────────────────────────┐
│           cmd/ (CLI Layer)              │
│         Can import anything             │
├─────────────────────────────────────────┤
│                   ▼                     │
├─────────────────────────────────────────┤
│    infrastructure/ (Implementation)     │
│  Can import: domain only                │
│  - validation/                          │
│  - tagging/                             │
│  - filesystem/                          │
│  - storage/                             │
│  - scraping/                            │
├─────────────────────────────────────────┤
│                   ▼                     │
├─────────────────────────────────────────┤
│       domain/ (Core Business)           │
│  Cannot import: anything!               │
│  Pure Go, no external deps              │
│  - Album, Track, Artist                 │
│  - ValidationIssue, Level, Role         │
└─────────────────────────────────────────┘

Rule: Dependencies point INWARD only
```

## Package Structure

```
internal/
│
├── domain/                    ✅ Complete
│   ├── level.go              # Enum: ERROR, WARNING, INFO
│   ├── role.go               # Enum: Composer, Soloist, etc.
│   ├── artist.go             # Value object
│   ├── validation_issue.go   # Value object
│   ├── edition.go            # Value object
│   ├── track.go              # Entity
│   ├── album.go              # Aggregate root
│   └── *_test.go             # Tests (>95% coverage)
│
├── validation/                ✅ Complete
│   ├── rules.go              # Rule database
│   ├── validator.go          # AlbumValidator
│   └── validator_test.go     # Tests
│
├── tagging/                   ✅ Read, 🚧 Write
│   ├── reader.go             # ✅ FLACReader
│   ├── reader_test.go        # ✅ Tests
│   ├── writer.go             # 🚧 FLACWriter (TODO)
│   └── writer_test.go        # 🚧 Tests (TODO)
│
├── filesystem/                ✅ Complete
│   ├── directory_validator.go
│   └── directory_validator_test.go
│
├── storage/                   ✅ Complete
│   ├── repository.go         # JSON serialization
│   └── repository_test.go    # Round-trip tests
│
└── scraping/                  🚧 TODO
    ├── scraper.go            # Common interface
    ├── harmoniamund/         # Site-specific
    ├── naxos/
    ├── classicalarchives/
    ├── presto/
    └── arkivmusic/
```

## Validation Pipeline

```
                  ┌─────────────────┐
                  │  Album + Files  │
                  └────────┬────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Structure   │  │   Metadata   │  │  Filesystem  │
│  Validation  │  │  Validation  │  │  Validation  │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       │                 │                 │
       └────────┬────────┴────────┬────────┘
                │                 │
                ▼                 ▼
        ┌───────────────────────────────┐
        │   Collect All Issues          │
        │   - ERROR (must fix)          │
        │   - WARNING (should fix)      │
        │   - INFO (nice to fix)        │
        └───────────────┬───────────────┘
                        │
                        ▼
        ┌───────────────────────────────┐
        │   Generate Report             │
        │   - Group by level            │
        │   - Show rule references      │
        │   - Colored output            │
        └───────────────────────────────┘
```

## Rule Application

```
Rule Database (validation/rules.go)
         │
         ├── Section 2.3.12: "Path length ≤ 180"
         ├── Section 2.3.16.4: "Required tags"
         ├── Section classical.composer: "Not in title"
         └── ... (50+ rules)
         │
         ▼
    Applied by:
         │
    ├────┴────────┐
    ▼             ▼
domain/       filesystem/
Validate()    DirectoryValidator
    │             │
    └──────┬──────┘
           │
           ▼
    ValidationIssue[]
    (with rule references)
```

## JSON Schema Flow

```
Domain Model                    JSON DTO
    │                              │
    │                              │
┌───▼──────────────┐     ┌────────▼──────────┐
│ Album            │────▶│ AlbumDTO          │
│ - title          │     │ - title           │
│ - originalYear   │     │ - original_year   │
│ - edition        │────▶│ - edition         │
│ - tracks[]       │     │   - label         │
│                  │     │   - catalog_number│
│                  │     │   - edition_year  │
│                  │     │ - tracks[]        │
└──────────────────┘     └───────────────────┘
         ▲                        │
         │                        │
         │    LoadFromJSON()      │
         └────────────────────────┘
              SaveToJSON()
```

## Test Strategy

```
                    ┌──────────────┐
                    │   Testing    │
                    └──────┬───────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Unit       │  │ Integration  │  │ Acceptance   │
│   Tests      │  │   Tests      │  │   Tests      │
├──────────────┤  ├──────────────┤  ├──────────────┤
│ Every func   │  │ Workflows    │  │ Real files   │
│ Edge cases   │  │ Round-trips  │  │ End-to-end   │
│ Error paths  │  │ Cross-pkg    │  │ User stories │
│              │  │              │  │              │
│ >95% cover   │  │ Happy paths  │  │ Full albums  │
└──────────────┘  └──────────────┘  └──────────────┘
```

## Deployment Architecture (Future)

```
┌─────────────────────────────────────────────┐
│            User Workstation                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │ validate │  │   tag    │  │ extract  │  │
│  │   CLI    │  │   CLI    │  │   CLI    │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  │
└───────┼─────────────┼─────────────┼─────────┘
        │             │             │
        │             │             │
        ▼             ▼             ▼
   ┌─────────────────────────────────────┐
   │         Local Files                 │
   │  - FLAC audio files                 │
   │  - JSON metadata                    │
   │  - Directory structures             │
   └─────────────────────────────────────┘
                     │
                     │ (extract only)
                     ▼
   ┌─────────────────────────────────────┐
   │      External Data Sources          │
   │  - Harmonia Mundi                   │
   │  - Classical Archives               │
   │  - Naxos                            │
   │  - Presto Classical                 │
   │  - ArkivMusic                       │
   └─────────────────────────────────────┘
```

## Legend

```
✅ Complete and tested
🚧 In progress / TODO
⏸️ Not started
```

---

**Note:** This is a living document. Update as architecture evolves.
