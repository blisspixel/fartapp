# Compute architecture

F.A.R.T. Lab is CPU-complete and accelerator-aware. A supported CPU must be able
to run every shipped canonical operation appropriate to its declared
capability, including equation solvers, rule evaluators, constraint engines,
mapping witnesses, and conformance checks. High-fidelity cases may be slow.
GPUs reduce time to solution. They do not change governing laws, grant
accuracy, or turn an unvalidated closure into science.

This is a design contract. The current repository contains only the tiny Go CLI
and no CFD or GPU implementation.

## Language boundaries

The project chooses languages by assurance and hardware fit, not fashion:

| Layer | Primary language | Reason |
| --- | --- | --- |
| Analytical oracle | Go | Small auditable executable, exact fixtures, simple cross-platform builds |
| Product core and scalar CPU reference | Rust | Memory-safe ownership, dimension-safe APIs, deterministic orchestration, strong CLI and native boundaries |
| Production field kernels | C++20 with Kokkos | One HPC-oriented implementation targeting CPU, CUDA, HIP, and SYCL builds |
| NVIDIA acceleration | Kokkos CUDA with measured native CUDA escape hatches | Mature scientific and multi-GPU hardware path |
| AMD acceleration | Kokkos HIP with measured native HIP escape hatches | Native ROCm path and explicit parity testing |
| Intel acceleration | Kokkos SYCL | Optional qualified backend with device capability checks |
| Apple preview kernels | Metal Shading Language plus a thin native host | Native interactive compute, without an FP64 quantitative CFD claim |
| Native presentation | Godot, Rust GDExtension, and native shaders | Native Windows, macOS, and Linux interaction without a browser runtime |
| Scenarios, closures, and law packs | Versioned declarative data | Reviewable inputs with no arbitrary executable extension |

C++ is confined to the production field-kernel library and a narrow versioned C
ABI. Kokkos owns device-local mesh arrays, field stages, parcels, halos, and
reductions inside that library. C++ does not own scenarios, archives, units,
record or case-result identity, validation policy, or the CLI. Rust owns
lifecycle and host allocation across the boundary and treats device failures as typed results. The
Go oracle never depends on the field library or an accelerator.

No Python, JavaScript, JVM, browser, or vendor runtime is required to run the
application. Python may appear in isolated research notebooks or reference-data
preparation, never as the production solver or a release dependency.

## Backend contract

Every field solver targets an internal compute interface for:

- Device discovery and capability reports.
- Typed allocation, transfer, halo, kernel, reduction, and synchronization.
- Precision, rounding, fused-operation, denormal, and fast-math policy.
- Stable error reporting, cancellation, watchdog, and out-of-memory recovery.
- Timing, memory, topology, energy, and hardware provenance.
- Checkpoint import and export through canonical host data.

Initial implementations proceed in this order:

1. Small scalar and SIMD Rust CPU reference kernels.
2. C++20 Kokkos Serial, Threads, or OpenMP production field builds.
3. A Kokkos CUDA build for the benchmarked field solver.
4. Kokkos HIP and SYCL builds after the same backend conformance suite exists.
5. Apple Metal interactive preview kernels where their precision is sufficient.
6. Domain decomposition across several GPUs, then MPI-based multi-node research.

Kokkos permits one device backend per build. F.A.R.T. Lab therefore produces
separate, first-party backend libraries rather than one enormous binary. They
are loaded only from explicit installation locations after ABI, architecture,
manifest, and digest verification. There is no general third-party native plugin
search path.

`auto` selects only a backend that supports the requested equations, precision,
memory, and evidence level. Otherwise it falls back to CPU or explains why the
run cannot proceed. A vendor name is never a validity criterion.

The planned diagnostic surface is:

```console
fart compute list
fart compute inspect cuda:0
fart doctor --compute
fart case run case.toml --backend auto --precision verified
fart case verify run.fart --compare-backends cpu,cuda
fart benchmark run field-core --backend cuda --output evidence/
```

## Precision and reproducibility

The default scientific mode uses double precision where the selected model and
hardware require it, disables unsafe fast-math transformations, uses documented
operation ordering, and computes critical ledgers with compensated or stronger
accumulation where justified. A device without the required precision runs a
qualified reduced model, falls back to CPU, or refuses the claim.

The execution policy has four explicit classes:

- `verified`: FP64 where required, reproducible reduction topology where
  supported, strict compiler policy, full residuals, and backend comparison
  evidence.
- `stable`: run-to-run stable observables on one recorded hardware and toolchain
  profile, with numeric tolerances rather than universal bitwise identity.
- `throughput`: performance-oriented ordering or mixed precision, allowed only
  with an error study and never substituted into a stronger certificate.
- `visual`: renderer-oriented lower precision with no quantitative field claim.

Parallel floating-point reductions are generally order-sensitive. CUDA, HIP,
Metal, CPU vectorization, compiler versions, and device generations may produce
different low bits. The archive records hardware, driver, compiler, flags,
kernel revision, work decomposition, reduction policy, and precision. D2
cross-platform claims compare declared observables and conservation residuals;
D3 bitwise claims are exceptional.

Mixed precision is an optimization hypothesis. It must pass conditioning,
conservation, shock, positivity, convergence, and backend-differential studies
for each affected observable. Tensor or matrix accelerators are used only where
the algorithm and error analysis justify them.

## GPU verification

Every accelerator backend runs the same semantic suite as CPU plus device tests:

- Kernel-level analytical and manufactured cases.
- CPU-to-device differential tests across normal, boundary, and adversarial
  states.
- Conservation and positivity after every stage where practical.
- Fixed and randomized mesh, boundary, parcel, and material cases.
- Single-block, multi-block, single-device, and multi-device decomposition
  equivalence within declared tolerances.
- Deterministic-mode repeatability and performance-mode variance.
- Sanitizer, bounds, race, uninitialized-memory, cancellation, reset, device-loss,
  out-of-memory, and corrupted-checkpoint cases.
- Compiler-optimization comparisons, including fast math disabled and enabled as
  a deliberate experiment.
- Published performance only beside accuracy, hardware, driver, power, memory,
  and command provenance.

The CPU backend is a reference implementation, not presumed ground truth. Both
CPU and GPU are checked against analytical solutions, manufactured solutions,
trusted datasets, and independent formulations. Agreement between two copies of
the same bug is not evidence.

## Multi-GPU and institutional-scale runs

High-fidelity cases use explicit domain decomposition, halo exchange, load
balance, checkpointing, and failure recovery. The first multi-GPU target is one
workstation. A later Linux HPC profile may add CUDA-aware or ROCm-aware MPI,
vendor collectives where justified, job-scheduler launch, restartable
checkpoints, and content-addressed evidence bundles.

Multi-GPU results record topology, rank and device mapping, peer-access policy,
communication library, partition, reduction tree, clock policy, and failure
handling. Scaling reports include strong and weak scaling plus parallel
efficiency. A faster result that changes a certified observable outside tolerance
is a regression.

Cloud and cluster execution remain optional adapters over the headless CLI. The
public desktop application does not need credentials, a service, or network
access.

## Portability candidates and dependency budget

Kokkos is the selected production field-kernel portability layer because its
Serial, Threads, OpenMP, CUDA, HIP, and SYCL execution spaces can reduce the
number of independently maintained numerical implementations. It remains an
optional field-solver dependency, not part of the ordinary analytical binary.
Backend-specific layouts, kernels, and libraries still require profiling and
parity evidence. Kokkos does not provide Metal or MPI and does not make code
performance-portable by assertion.

Vulkan or WebGPU compute may be useful for visual effects and portable
lower-precision presentation. They are not the default quantitative CFD backend
until required precision, memory semantics, tooling, and benchmark evidence are
demonstrated on every supported platform.

Each optional backend is a separately built feature or package. The ordinary
CLI has no CUDA, ROCm, Metal, MPI, Kokkos, SYCL, or graphics dependency. Release
packages declare exactly which backends they contain. Unsupported libraries are
not downloaded at runtime.

## Mojo decision

Mojo is an interesting research backend, not a production commitment. Mojo 1.0
shipped on 2026-08-11 and began a language and library stability policy, but its
initial stable standard-library surface is deliberately small. Native Windows
is not supported, Apple GPU kernels do not provide the FP64 path required for
verified CFD, cross-compilation and systems tooling remain incomplete, and the
GPU programming surface crosses into MAX packages that require their own
dependency and redistribution review. Its cross-vendor direction is worth a
contained benchmark after the CPU and CUDA contracts exist.

Mojo can be promoted only after it has:

- A sufficiently broad stable language, standard-library, and package surface.
- Native Windows, macOS, and Linux toolchains needed by this project.
- A fully reviewable redistribution and offline-build path.
- Required FP64, atomics, synchronization, profiling, debugging, and device-loss
  behavior on the target hardware.
- Sanitizer and testing support appropriate to numerical kernels.
- Performance and accuracy at least competitive with the maintained native
  backend on the project benchmark suite.
- Lower total dependency and maintenance cost, not merely shorter source code.

Until then, Mojo prototypes stay outside release artifacts and cannot become the
only implementation of an equation or benchmark.

## Performance acceptance

Performance work starts with profiles and fixed evidence cases. Reports include
time to solution, cell updates per second, memory bandwidth, peak memory,
transfer cost, startup and compilation cost, power or energy where measurable,
and all scientific error metrics. Benchmarks cover a tiny CPU case, an
interactive case, a memory-bound field, a shock case, a parcel-heavy case, and a
multi-device case.

No roadmap item is complete because it "uses the GPU." It is complete when the
accelerated backend is faster on a named workload, preserves the required
scientific claims, fails safely, and can be independently reproduced.
