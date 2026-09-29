# SDK harness contracts

`tests.tsv` is the versioned entry-point contract used by the centralized SDK
quality harness. The provider repository owns the executable health,
conformance, and resilience assertions. The harness supplies a fresh local B2
simulator, fault controls, the provider built from this exact revision,
orchestration, evidence retention, and fleet reporting.

These contracts never require production B2 credentials. Terraform selects
the supplied provider executable through `dev_overrides`, so a passing result
cannot silently come from a previously published registry build.
