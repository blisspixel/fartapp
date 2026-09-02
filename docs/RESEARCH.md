# Research basis

This document records the authoritative sources behind the simulation, numerical,
gameplay, and interface plans. It is a design bibliography, not evidence that a
planned model has already been implemented or validated.

Sources were reviewed for this roadmap on 2026-09-01. Primary papers, standards,
government laboratories, and official project documentation are preferred.

## Decisions changed by the research

- The profile-neutral bounded core is Lab-local record identity, law-context set,
  scope, and provenance. Outside-formalism cases fail explicitly.
  Emitter, interface, exterior, dimensional equations, human anatomy, and Earth
  physics are capability-selected profile structures.
- Ordinary Earth-biological pressure excursions are far below the ideal-gas
  choking threshold. Choked, underexpanded, locally supersonic, and
  shock-containing are separate labels.
- Finite events begin as starting jets or puffs. Steady round-jet scaling is a
  conditional far-field limit.
- Coarse state history constrains procedural audio but does not uniquely contain
  unresolved turbulent waveform detail. The audio closure and seed are part of
  provenance.
- Wetness, breakup, and deposition require more than Weber or Stokes number.
- Dimensional similarity requires matching nondimensional equations, closures,
  dimensions, topology, normalized conditions, and active coefficients. Other
  law profiles define their own comparison signature.
- Conservation, code verification, solution verification, empirical validation,
  and fictional-law consistency are separate certificate claims.
- Alternate dimensions and universes get explicit law packs. Unsupported laws
  fail explicitly rather than reusing Earth fluid equations.
- The story director reacts to context-scoped occurrence claims and never rerolls a valid
  outcome.
- Physical acoustics, diagnostic sonification, Symphony, and radio are separate
  audio lanes with different truth claims and tests.
- Social meaning is optional. When supported, it is represented through situated
  perspectives, disagreement, and power, never one generated `culture.norms`
  fact or a mandatory institution.
- CLI, MCP, A2A, TUI, accessibility, and native automation lower into one
  canonical play reducer with explicit knowledge and budget policies.
- There is no universal agent-difficulty meaning for Level 4. F.A.R.T. Lab uses
  measurable challenge grades and separate semantic, visual, omnimodal,
  multi-agent, and human tracks.
- The CLI is the first complete product, the TUI is the second, and the native
  Godot app begins only after both are excellent.

## Evidence matrix

The bibliography is not a pile of prestigious names. Each important source has
a bounded design consequence and an open question.

| Claim | Evidence and supported scope | Implementation consequence | Open question |
| --- | --- | --- | --- |
| Derived artifacts need traceable provenance | [W3C PROV-DM](https://www.w3.org/TR/prov-dm/) defines entities, activities, derivations, and constraints | Use a typed occurrence provenance graph with versioned derivation edges | What minimal subset stays pleasant in a CLI? |
| Units can be part of program correctness | [Kennedy on units of measure](https://www.microsoft.com/en-us/research/publication/relational-parametricity-and-units-of-measure/) proves dimensional invariance properties for a typed language | Make invalid dimensional combinations difficult to represent | Which quantities need runtime profiles rather than compile-time dimensions? |
| Prediction before explanation can improve transfer | [Kapur's controlled studies](https://doi.org/10.1111/cogs.12107) concern mathematics learning | Let players predict, fail productively, then inspect and transfer | Does this effect hold for short fluid-dynamics play sessions? |
| Complex model conclusions need sensitivity analysis | [Saltelli et al.](https://doi.org/10.1002/9780470725184) covers global sensitivity methods | Report influential inputs, interactions, and nonidentifiability | Which methods fit interactive latency and mixed variables? |
| Sonification must be perceptually evaluated | The [NSF sonification report](https://digitalcommons.unl.edu/psychfacpub/444/) calls for control, interchange, and perceptual testing | Treat sound as both consequence and auditory display, with separate tests | Which mappings teach without ruining the joke or accessibility? |
| Generator robustness does not establish entertainment | The [PCG literature](https://www.pcgbook.com/) provides generation and evaluation methods, not a universal fun metric | Pair seed sweeps with a curated corpus and human playtests | Which repetition and pacing measures predict shareability? |
| Protocol parity requires an application contract | [MCP 2026-07-28](https://modelcontextprotocol.io/specification/2026-07-28) is stateless at its core and [A2A 1.0](https://a2a-protocol.org/v1.0.0/specification/) models longer agent tasks | Keep play handles, state, actions, knowledge, and replay in `PlayService`, not transport sessions | What smallest adapter set remains pleasant for unrelated agents? |
| More modalities can reduce agent performance | [OmniPlay](https://arxiv.org/abs/2508.04361) evaluates reinforcing and conflicting image, video, audio, and text | Identify physical audio, radio, captions, and semantic observations as separate benchmark channels | Which combinations improve reasoning rather than add noise? |
| Social generation needs situated review | [UNESCO's ICH ethical principles](https://ich.unesco.org/en/ethics-and-ich-00866) emphasize community participation, diversity, and context | Model plural positions and trigger focused review for recognizable living affinity | When is generated resemblance substantial enough to require review? |

## Gas dynamics, reservoirs, and jets

- [NASA Glenn: Mass Flow Choking](https://www.grc.nasa.gov/www/k-12/BGP/mflchk.html)
  derives the compressible mass-flow function and its Mach 1 maximum.
- [NASA Glenn: Isentropic Flow Equations](https://www.grc.nasa.gov/www/k-12/airplane/isentrop.html)
  gives stagnation ratios and area-Mach relations under their assumptions.
- [NASA GFSSP validation report](https://ntrs.nasa.gov/api/citations/19980231078/downloads/19980231078.pdf)
  provides analytical choking and pressurized nitrogen-tank blowdown cases.
- [Dutton and Coverdill: gaseous vessel discharge experiments](https://experts.illinois.edu/en/publications/experiments-to-study-the-gaseous-discharge-and-filling-of-vessels/)
  compares measured discharge and filling against limiting theories.
- [NASA report on underexpanded jet shock structure](https://ntrs.nasa.gov/api/citations/19820022412/downloads/19820022412.pdf)
  supports the distinction between choking and external shock-cell structure.
- [Hussein, Capp, and George: high-Reynolds-number round jet](https://turbulence-online.com/Publications/Papers/HCG94.pdf)
  reports self-similar axisymmetric jet measurements and momentum behavior.
- [Morton, Taylor, and Turner: turbulent plumes](https://doi.org/10.1098/rspa.1956.0011)
  establishes integral entrainment theory for buoyant sources.
- [Gharib, Rambod, and Shariff: vortex-ring formation](https://doi.org/10.1017/S0022112097008410)
  supports source-stroke analysis with important geometry-specific limits.
- [Peña Fernández and Sesterhenn: compressible starting jets](https://doi.org/10.1017/jfm.2017.128)
  supports high-fidelity transient cases with shock, shear-layer, and vortex
  interaction. It is not an ordinary low-pressure puff model.
- [Zhu et al.: stopping vortex rings](https://doi.org/10.1017/jfm.2024.883)
  shows why source deceleration and closure remain part of impulse and vortex
  evolution after the main supply phase.
- [Flexible-nozzle pulsed-jet experiments](https://doi.org/10.1017/jfm.2024.720)
  demonstrate that stored elastic energy and closure timing can change impulse
  and entrainment. Their engineered liquid nozzle supplies a mechanism and
  benchmark shape, not biological calibration values.
- [Patel et al.: particle-laden underexpanded jets](https://doi.org/10.1017/jfm.2024.1014)
  couples an Eulerian gas with Lagrangian particles and finds particle loading
  can shift shock structure. It bounds later dense high-pressure models.
- [Ambulatory anorectal pressure study](https://pubmed.ncbi.nlm.nih.gov/3219524/)
  provides scale evidence for ordinary biological pressure excursions.
- [Measured flatus composition study](https://pubmed.ncbi.nlm.nih.gov/9176210/)
  shows substantial event and subject variability in major and trace gases.
- [Quantitative GC-TCD measurements](https://pubmed.ncbi.nlm.nih.gov/35161583/)
  supply a recent measurement-method reference but only a five-subject dietary
  demonstration, not a universal composition.
- [Flatus-related motor events](https://pubmed.ncbi.nlm.nih.gov/8601379/) provide
  small-sample pressure and coordination evidence, not a synchronized source,
  aperture, mass-flow, and audio waveform.
- [Twenty-four-hour flatus production](https://pubmed.ncbi.nlm.nih.gov/1648028/)
  measures daily collected volume, which cannot be assigned to one event without
  an additional model.

## Acoustics, interfaces, and rooms

- [Lighthill: On Sound Generated Aerodynamically](https://doi.org/10.1098/rspa.1952.0060)
  is the foundation for deriving aerodynamic sound sources from flow.
- [Titze: compliant biological-interface oscillation](https://pubmed.ncbi.nlm.nih.gov/3372869/)
  provides a mathematical analogue for pressure-flow-structure coupling without
  implying identical anatomy.
- [Single-mass vocal-fold fluid-structure review](https://pmc.ncbi.nlm.nih.gov/articles/PMC2857605/)
  explains why self-sustained oscillation needs aerodynamic feedback.
- [Collapsible-tube experiments](https://doi.org/10.1063/5.0211227) measure open,
  collapsed, and self-excited compliant-flow regimes and support an explicit
  fluid-structure energy budget.
- [Flexible-wall instability study](https://doi.org/10.1017/jfm.2025.10860)
  shows that some modes survive reduced modeling while other regimes need a
  higher-dimensional coupled system. Fidelity must escalate by observable.
- [Ffowcs Williams and Hawkings](https://doi.org/10.1098/rsta.1969.0031)
  provides a moving-surface acoustic analogy for later field tiers.
- [Wave interactions in a screeching jet](https://arxiv.org/abs/2603.04786)
  analyzes shock-cell, instability-wave, guided-mode, and nonlinear feedback
  interactions. It supports making screech an earned regime, not an audio flag.
- [Allen and Berkley: image-source rooms](https://doi.org/10.1121/1.382599)
  provides deterministic rectangular-room impulse-response modeling.
- [NASA: sound and the vacuum limit](https://science.nasa.gov/ems/02_anatomy/)
  supports the requirement for a material medium for external sound.

## Droplets, particles, bubbles, and vacuum

- [NASA droplet-breakup memorandum](https://ntrs.nasa.gov/api/citations/20070034950/downloads/20070034950.pdf)
  documents Weber, Ohnesorge, breakup-mode, and breakup-time dependencies.
- [Hinze: breakup in dispersion processes](https://doi.org/10.1002/aic.690010303)
  is a foundational source for turbulent breakup criteria and viscosity effects.
- [Role of viscosity in turbulent drop breakup](https://doi.org/10.1017/jfm.2023.345)
  treats the transition as probabilistic and shows why Weber number alone is not
  a universal Boolean boundary.
- [The Atomizing Pulsed Jet](https://arxiv.org/abs/2405.01959) reports complex
  sheet, ligament, droplet, and bubble topology plus a nonconvergent numerical
  droplet peak near three grid cells. It directly motivates cutoff reporting.
- [Evaporating turbulent jet-spray DNS](https://arxiv.org/abs/2010.07689)
  reports clustering and heterogeneous histories that violate naive local
  isolated-droplet assumptions.
- [Statistical turbulent aerosol dispersal](https://doi.org/10.1103/PhysRevFluids.10.054302)
  supports reporting realization variability and ensembles instead of only a
  mean indoor concentration.
- [Sandia aerosol transport and deposition report](https://www.osti.gov/servlets/purl/1675151)
  connects response time, Stokes behavior, settling, inertia, diffusion, and
  deposition mechanisms.
- [Brennen: Cavitation and Bubble Dynamics](https://media.library.caltech.edu/CaltechBOOK%3A1995.001/chap2.htm)
  derives Rayleigh-Plesset dynamics and states their assumptions.
- [Minnaert: musical air bubbles](https://doi.org/10.1080/14786443309462277)
  provides the classic isolated-bubble resonance model.
- [Detaching-bubble acoustics](https://doi.org/10.1103/xshn-mnb8) and its
  [open experimental dataset](https://doi.org/10.57745/Y1XCZM) support deriving
  underwater amplitude from evolving radius, neck, and detachment dynamics.
- [NASA rarefied CFD and DSMC comparison](https://ntrs.nasa.gov/citations/20230013741)
  discusses continuum-breakdown criteria and rarefied-model applicability.
- [NASA rocket thrust equation](https://www1.grc.nasa.gov/beginners-guide-to-aeronautics/rocket-thrust-equation/)
  supports momentum and pressure thrust, including vacuum operation.
- [NASA ideal rocket equation](https://www1.grc.nasa.gov/beginners-guide-to-aeronautics/ideal-rocket-equation/)
  provides the mass-ratio and delta-v reference.

## Thermophysical, plasma, stellar, and relativistic profiles

- [NIST REFPROP](https://doi.org/10.18434/T4/1502528) provides reference
  thermophysical properties for many fluids and mixtures. Integration requires
  a separate license review.
- [NIST humid-air thermodynamics review](https://tsapps.nist.gov/publication/get_pdf.cfm?pub_id=935265)
  documents humid-air mixture behavior and model limits.
- [NASA Chemical Equilibrium with Applications](https://www.nasa.gov/glenn/research/chemical-equilibrium-with-applications/)
  supports equilibrium reacting and ionized-mixture reference cases.
- [NRL Plasma Formulary 2023](https://www.nrl.navy.mil/Portals/38/PDF%20Files/NRL_Plasma_Formulary_2023.pdf)
  provides plasma scale and regime diagnostics.
- [NASA CCMC AWSoM model](https://ccmc.gsfc.nasa.gov/models/SWMF~AWSoM_R~1.0/)
  demonstrates the MHD, thermal, wave, and kinetic ingredients of a serious
  solar-wind profile.
- [Relativistic fluid dynamics review](https://link.springer.com/article/10.1007/s41114-021-00031-6)
  covers conserved relativistic currents, compact objects, and equation-of-state
  dependence.

## Dimensional analysis and numerical proof

- [Buckingham: On Physically Similar Systems](https://doi.org/10.1103/PhysRev.4.345)
  is the original Pi-theorem formulation.
- [BIPM SI Brochure](https://www.bipm.org/en/publications/si-brochure) defines
  the International System of Units and exact defining constants inherited by
  the first Earth law profile. Those constants define coordinates, not a
  physiology or one scalar fart unit.
- The [International Vocabulary of Metrology](https://www.bipm.org/en/committees/jc/jcgm/publications)
  distinguishes measurands, calibration, metrological traceability, and
  uncertainty concepts used by the Reference Pfft procedure.
- [Barth and Ohlberger: finite-volume methods](https://ntrs.nasa.gov/citations/20030020790)
  provides a foundation for conservative discretization.
- [Zhang and Shu: positivity-preserving Euler schemes](https://doi.org/10.1016/j.jcp.2010.08.016)
  supports positive density and pressure under explicit numerical conditions.
- [Sayyari and Yamaleev: implicit positivity-preserving entropy-stable BDF2](https://arxiv.org/abs/2608.20103)
  is a current research candidate for unsteady compressible viscous flow. Its
  stated stability and positivity properties do not eliminate application-level
  convergence and validation work.
- [Yang and Fu: positivity and entropy behavior on curvilinear AMR](https://arxiv.org/abs/2604.21600)
  documents conservation, admissibility, and a real tradeoff between one
  high-order mortar construction and provable entropy stability.
- [Roache: Method of Manufactured Solutions](https://doi.org/10.1115/1.1436090)
  provides a code-verification method independent of physical validation.
- [NASA verification and validation tutorial](https://www.grc.nasa.gov/www/wind/valid/tutorial/tutorial.html)
  separates numerical verification from physical validation.
- [NASA grid-convergence guidance](https://www.grc.nasa.gov/www/wind/valid/tutorial/spatconv.html)
  gives observed-order and Grid Convergence Index practices.
- [JCGM Guide to the Expression of Uncertainty in Measurement](https://doi.org/10.59161/JCGM100-2008E)
  supports explicit uncertainty categories and reporting.
- [NIST numerical reproducibility program](https://www.nist.gov/programs-projects/numerical-reproducibility)
  motivates careful cross-platform reproducibility contracts.
- [Salmon et al.: counter-based random numbers](https://doi.org/10.1145/2063384.2063405)
  supports independently addressable deterministic streams.
- [RFC 8785: JSON Canonicalization Scheme](https://www.rfc-editor.org/rfc/rfc8785.html)
  defines canonical JSON suitable for stable metadata hashing.
- [NISTIR 8298](https://doi.org/10.6028/NIST.IR.8298) frames CFD verification,
  validation, and uncertainty as distinct evidence activities.

## CPU, GPU, and numerical language choices

- The current [CUDA Programming Guide](https://docs.nvidia.com/cuda/cuda-programming-guide/)
  documents scientific GPU execution, floating-point behavior, and multi-GPU
  facilities. It also warns that bitwise determinism depends on specified
  hardware and execution conditions.
- The [CUDA multi-GPU guide](https://docs.nvidia.com/cuda/cuda-programming-guide/03-advanced/multi-gpu-systems.html)
  covers contexts, decomposition, peer transfers, synchronization, and MPI or
  communication-library integration. Multi-GPU support is therefore a distinct
  algorithm and verification milestone, not an automatic switch.
- The [HIP porting guide](https://rocm.docs.amd.com/projects/HIP/en/latest/how-to/hip_porting_guide.html)
  defines HIP as a C++ runtime and kernel language for AMD GPUs with close CUDA
  correspondence. Portability still requires independent AMD evidence.
- Apple's [Metal compute documentation](https://developer.apple.com/documentation/metal/compute-passes)
  supports native parallel compute on Apple GPUs. Feature and precision
  requirements must be queried per target rather than assumed from the API name.
- [Kokkos backend configuration](https://kokkos.org/kokkos-core-wiki/get-started/configuration-guide.html)
  supports CPU, CUDA, HIP, and SYCL execution spaces, but currently selects one
  device backend per build. It is the planned C++20 field-kernel layer, packaged
  as separate first-party backends. Its proof surface and lack of Metal and MPI
  support remain explicit costs.
- [SYCL](https://www.khronos.org/sycl/) is a standards-based C++ portability
  candidate. It is not selected until project kernels demonstrate target
  coverage, accuracy, tooling, and competitive performance.
- [Mojo 1.0](https://mojolang.org/releases/v1.0.0/) shipped on 2026-08-11 and
  begins a stability policy while deliberately stabilizing only a small initial
  standard-library surface. Its
  [system requirements](https://mojolang.org/docs/requirements/) still report no
  native Windows toolchain. Apple GPU precision, cross-compilation, packaging,
  profiling, MAX runtime dependencies, and production CFD evidence remain gates.
  Mojo therefore stays an isolated research backend.

The production choice is a safe Rust control and scalar CPU-reference core, an
optional C++20 Kokkos field library built separately for CPU, CUDA, HIP, and
SYCL behind a narrow C ABI, and optional non-certifying Metal preview kernels.
The ordinary executable remains CPU-complete and free of accelerator
dependencies. See [COMPUTE.md](COMPUTE.md).

## Scientific workflow, causality, and software correctness

- [W3C PROV-DM](https://www.w3.org/TR/prov-dm/) supplies a domain-neutral model
  for entities, activities, derivations, and provenance constraints. The event
  graph borrows those distinctions without requiring a web runtime.
- [Kennedy: Relational Parametricity and Units of Measure](https://www.microsoft.com/en-us/research/publication/relational-parametricity-and-units-of-measure/)
  connects a type system for units with invariance under unit changes. It
  supports dimension-safe core APIs, not a claim that static types replace
  runtime law-profile validation.
- [Global Sensitivity Analysis: The Primer](https://doi.org/10.1002/9780470725184)
  supports reporting how uncertainty in inputs and assumptions affects model
  outputs. Interactive local sensitivity and slower global studies need separate
  cost and interpretation rules.
- [Causal testing for scientific modeling software](https://arxiv.org/abs/2209.00357)
  explores causal inference and metamorphic relations for scientific programs.
  F.A.R.T. Lab will use independently justified metamorphic properties, such as
  unit invariance and similarity scaling, rather than treating any relation as
  automatic proof.
- The [National Academies reproducibility report](https://doi.org/10.17226/25303)
  distinguishes computational reproducibility from broader scientific
  replication. That distinction informs event identities and certificate claims.

## Learning, explanation, and auditory display

- [Kapur: Productive Failure in Learning Math](https://doi.org/10.1111/cogs.12107)
  reports improved conceptual understanding and transfer when problem solving
  precedes instruction in its studied setting. The project's predict, emit,
  explain, vary, transfer loop is a hypothesis to test, not a guaranteed effect.
- [Improving Comprehension of Numbers in the News](https://www.microsoft.com/en-us/research/publication/improving-comprehension-of-numbers-in-the-news/)
  reports controlled experiments in which contextual perspectives improved
  numerical recall and estimation. Explanation cards should connect absurd
  outputs to familiar scale without substituting analogy for units.
- The [NSF Sonification Report](https://digitalcommons.unl.edu/psychfacpub/444/)
  defines sonification as nonspeech audio used to convey information and calls
  for perceptual testing. Procedural audio and data sonification share event
  provenance, but their entertainment and information goals are evaluated
  separately.
- [ITU-R BS.1770-5](https://www.itu.int/rec/R-REC-BS.1770-5-202311-I/en)
  defines program-loudness and true-peak measurement algorithms. It supports
  reproducible measurement, not one mandatory artistic target or a substitute
  for safe monitor gain.
- [ITU-T H.872](https://www.itu.int/epublications/publication/itu-t-h-872-2024-10-safe-listening-for-video-gameplay-and-esports)
  defines safe-listening provisions for video gameplay and esports, including
  information, controls, loudness, and true-peak targets. The Lab applies its
  software profile to idents, stems, mixes, transitions, and representative long
  Chill sessions while retaining independent dynamic-range controls.
- [ITU-R BS.1534-3](https://www.itu.int/rec/R-REC-BS.1534-3-201510-I/en)
  and [ITU-R BS.1116-3](https://www.itu.int/rec/R-REC-BS.1116-3-201502-I/en)
  provide controlled methods for intermediate and small audio impairments.
  Listening tests for realism, mapping quality, transitions, and fatigue need
  different stimuli and questions.

## Music generation and radio production

- The [Eleven Music v2 release](https://elevenlabs.io/docs/changelog/2026/6/15)
  documents `music_v2`, structured plans, compose, streaming, and upload API
  support.
- The official [compose endpoint](https://elevenlabs.io/docs/api-reference/music/compose)
  documents prompt or composition-plan input, explicit `model_id`, output
  formats, duration limits, and a seed that improves consistency without
  guaranteeing exact reproduction after service changes.
- The [composition-plan guide](https://elevenlabs.io/docs/eleven-api/guides/how-to/music/composition-plans)
  supports ordered v2 chunks with styles, lyrics, and durations. The pipeline
  freezes accepted audio bytes because a provider seed is not an archive.
- The [music quickstart](https://elevenlabs.io/docs/eleven-api/guides/cookbooks/music)
  documents detailed responses containing audio and the resolved plan and
  rejects artist, song, and copied-lyric references.
- [Eleven Music model-specific terms](https://elevenlabs.io/eleven-music-model-specific-terms),
  [Music Terms](https://elevenlabs.io/music-terms), and plan rights can change
  independently from this repository's Apache License 2.0. Generated assets
  therefore require a project rights decision at release time. Provider access
  never runs during ordinary CI or becomes a player dependency.
- The [USPTO sound-mark examples](https://www.uspto.gov/trademarks/soundmarks/trademark-sound-mark-examples)
  confirm that audio signatures can function as registered source identifiers,
  including the familiar cinema reference that motivated this feature. The
  project creates an independently specified gesture and performs similarity
  review rather than treating genre or technical differences as clearance.
- The USPTO's [sound-mark filing guidance](https://www.uspto.gov/trademarks/basics/mark-drawings-trademarks)
  requires an audio reproduction and detailed written description for a sound
  mark. F.A.R.T. Lab preserves a canonical master and description even if no
  registration is pursued.

## Agent environments and interoperability

- The [MCP 2026-07-28 release](https://blog.modelcontextprotocol.io/posts/2026-07-28/)
  removes the initialize handshake and protocol session, adds discovery and
  header routing, moves long-running Tasks into an extension, and deprecates
  Roots, Sampling, Logging, and legacy HTTP+SSE. Stateful games should return
  explicit application handles. Its Tier 1 TypeScript, Python, Go, and C# SDKs
  support the revision, while the Rust SDK is described as beta. The project
  requires a maturity and conformance gate rather than selecting by language
  preference.
- The official [MCP conformance framework](https://github.com/modelcontextprotocol/conformance)
  validates messages and scenarios for declared protocol revisions. Its
  2026-07-28 coverage and open harness issues must be pinned and reported rather
  than hidden in a permanent expected-failure baseline.
- The [A2A 1.0.0 specification](https://a2a-protocol.org/v1.0.0/specification/)
  defines Agent Cards, skills, messages, tasks, artifacts, streaming, bindings,
  authorization, and capability equivalence. The latest reviewed patch is
  [v1.0.1](https://github.com/a2aproject/A2A/releases/tag/v1.0.1), while wire
  negotiation remains `1.0`.
- The official [A2A and MCP comparison](https://github.com/a2aproject/A2A/blob/main/docs/topics/a2a-and-mcp.md)
  treats MCP as agent-to-tool interoperability and A2A as coordination between
  autonomous agents. F.A.R.T. Lab keeps them as sibling adapters.
- The [A2A TCK](https://github.com/a2aproject/a2a-tck) tests declared protocol
  bindings, while the [A2A ITK](https://github.com/a2aproject/a2a-itk) exercises
  cross-SDK and multi-hop behavior. Using an SDK is not itself conformance.
- [Gymnasium's environment API](https://gymnasium.farama.org/api/env/) separates
  natural termination from budget truncation. [PettingZoo](https://pettingzoo.farama.org/)
  provides sequential and simultaneous multi-agent environment contracts.
- [TALES](https://arxiv.org/abs/2504.14128) and its
  [official suite](https://github.com/microsoft/tale-suite) show long-horizon
  failure from lost evidence, incorrect inferences, and undiscovered mechanics
  across deterministic text adventures.
- [BALROG](https://arxiv.org/abs/2411.13543) and its
  [official environment](https://github.com/balrog-ai/BALROG) motivate
  procedural worlds, partial progress, held-out instances, and separate visual
  and language tracks over action horizons reaching far beyond short tool use.
- [ORAK](https://arxiv.org/abs/2506.03610) and its
  [official repository](https://github.com/krafton-ai/ORAK) provide direct
  precedent for text and visual game agents connected through MCP.
- [GameWorld](https://arxiv.org/abs/2604.07429) normalizes semantic and GUI
  actions into a budgeted event space and compares paused and real-time play.
  Its own level labels and limited reported human baseline are not adopted as a
  universal standard.
- [OmniGameArena](https://arxiv.org/abs/2606.09826) motivates measuring
  improvement over repeated play and transfer to held-out variants rather than
  reporting only cold-start completion.
- The name ClawBench is currently ambiguous. The
  [claw-bench repository](https://github.com/claw-bench/claw-bench) is a generic
  pytest task suite whose public counts are inconsistent, while
  [ShellBench](https://github.com/openclaw/shellbench) formerly used the same
  name and emphasizes trace and reliability analysis. Neither establishes a
  scientific meaning for `Level 4`; both are implementation references only.

## Culture, religion, and public interest

- [Harvard's Religious Literacy Project](https://rpl.hds.harvard.edu/what-we-do/our-approach/core-principles)
  treats religions as internally diverse, changing, embedded in culture, and
  shaped by power. A world therefore cannot receive one deterministic religious
  or moral rule from its environment.
- [UNESCO's ethical principles for intangible cultural heritage](https://ich.unesco.org/en/ethics-and-ich-00866)
  emphasize community participation, diversity, consent, and customary access.
  Sacred or private contexts are boundaries, not random comic props.
- The [UNESCO Recommendation on the Ethics of Artificial Intelligence](https://www.unesco.org/en/artificial-intelligence/recommendation-ethics)
  supports accountable human oversight, diversity, bias assessment, and digital
  access in generated-content and agent systems.
- The [United Nations Declaration on the Rights of Indigenous Peoples](https://www.un.org/development/desa/indigenouspoples/wp-content/uploads/sites/19/2018/11/UNDRIP_E_web.pdf)
  protects Indigenous cultural, intellectual, religious, and spiritual material
  and makes real provenance, authority, consent, correction, and withdrawal
  necessary for any high-affinity authored pack.
- The [Design Justice Network Principles](https://designjustice.org/read-the-principles)
  center people affected by design decisions and support focused, compensated
  review when content closely resembles a living community.
- The [WHO and UNICEF Joint Monitoring Programme](https://washdata.org/reports)
  documents sanitation and privacy as material inequalities. Poverty,
  inadequate sanitation, incontinence, and involuntary anatomy are not spectacle
  or shorthand for lesser personhood.
- [Bringing Dark Patterns to Light](https://www.ftc.gov/reports/bringing-dark-patterns-light)
  documents manipulative interface practices. The complete core stays offline,
  account-free, telemetry-free, and free of streaks, fear of missing out, loot
  boxes, and pay-to-win systems.
- [WCAG 2.2](https://www.w3.org/TR/WCAG22/) and the Xbox Accessibility
  Guidelines inform equivalent meaning, controllable timing, input alternatives,
  captions, and clear state across terminal and native surfaces.
- The [Stanford Encyclopedia of Philosophy overview of Japanese aesthetics](https://plato.stanford.edu/entries/japanese-aesthetics/)
  supplies historical and philosophical context for impermanence, `wabi`, and
  `sabi`. It does not turn them into interchangeable branding phrases.
- The Government of Japan's discussion of
  [`ichigo ichie`](https://www.gov-online.go.jp/eng/publicity/book/hlj/html/202112/202112_05_en.html)
  supports the one-encounter framing. The implementation adopts impermanence and
  attention, not decorative Japanese motifs or claims of cultural ownership.

These sources do not certify generated humor as respectful or funny. Structural
checks catch predictable failures; situated people judge meaning, dignity, and
comic quality.

## Procedural play, narrative, and accessibility

- [Procedural Content Generation in Games](https://www.pcgbook.com/) covers
  grammars, search, quests, stories, and generator evaluation.
- [Storylets design space](https://mkremins.github.io/publications/Storylets_SketchingAMap.pdf)
  provides a formal basis for modular, state-conditioned narrative beats.
- [MDA framework](https://www.cs.northwestern.edu/~hunicke/MDA.pdf) supports
  designing mechanics from desired player experiences.
- [Motivational model of game engagement](https://journals.sagepub.com/doi/pdf/10.1037/a0019440?download=true)
  supports competence, autonomy, and relatedness over a pure magnitude grind.
- [Random123](https://random123.com/) provides practical counter-based random
  generators for stable named streams.
- [Xbox Accessibility Guidelines](https://learn.microsoft.com/en-us/xbox/accessibility/guidelines)
  covers text, contrast, narration, inputs, motion, photosensitivity, difficulty,
  time limits, and documentation.
- [Microsoft Gaming Accessibility Testing guidance](https://learn.microsoft.com/en-us/xbox/accessibility/mgats)
  supports involving players with disabilities once a representative slice is
  stable, followed by remediation and retesting rather than a final-release-only
  review.
- [Accessible Player Experiences](https://accessible.games/accessible-player-experiences/)
  frames accessibility through concrete player experience patterns.

## CLI, archives, terminal UI, and native delivery

- [Command Line Interface Guidelines](https://clig.dev/) supports a human-first,
  composable, predictable command experience.
- [GNU argument syntax conventions](https://sourceware.org/glibc/manual/latest/html_node/Argument-Syntax.html)
  support `--` as the option terminator and `-` as a conventional standard
  stream operand.
- [Go signal documentation](https://go.dev/pkg/os/signal/#hdr-SIGPIPE)
  defines the native fd 1 and fd 2 `SIGPIPE` behavior preserved by the CLI.
- [The Update Framework](https://theupdateframework.io/) defines signed metadata
  roles and defenses against rollback, freeze, and mix-and-match attacks used by
  the standalone update design.
- [Sigstore verification documentation](https://docs.sigstore.dev/cosign/verifying/verify/)
  supports identity and transparency-bundle verification for release artifacts.
- [clap derive](https://docs.rs/clap/latest/clap/_derive/),
  [completion generation](https://docs.rs/clap_complete/latest/clap_complete/),
  and [manual generation](https://docs.rs/clap_mangen/latest/clap_mangen/struct.Man.html)
  support one typed Rust command model.
- [Rust terminal detection](https://doc.rust-lang.org/std/io/trait.IsTerminal.html)
  supports correct interactive and piped behavior.
- [Ratatui backends](https://ratatui.rs/concepts/backends/) and
  [snapshot testing](https://ratatui.rs/recipes/testing/snapshots/) support the
  cross-platform terminal dashboard and cell-buffer tests.
- [NO_COLOR](https://no-color.org/) defines a widely used opt-out convention.
- [JSON Schema 2020-12](https://json-schema.org/draft/2020-12) supports versioned
  machine-readable contracts.
- [Apache Arrow format](https://arrow.apache.org/docs/format/Columnar.html)
  supports large typed time histories behind an isolated archive layer.
- [PKWARE ZIP APPNOTE](https://support.pkware.com/pkzip/appnote) defines the
  proposed package container and reader hazards that must be handled.
- [Rust random reproducibility](https://rust-random.github.io/book/crate-reprod.html)
  explains why generator choice and version must be pinned.
- [Microsoft RIFF/WAVE structure](https://learn.microsoft.com/en-us/windows/win32/xaudio2/resource-interchange-file-format--riff-)
  defines the initial offline audio container.
- [Godot AudioStreamGenerator](https://docs.godotengine.org/en/latest/classes/class_audiostreamgenerator.html)
  supports native procedural playback.
- [godot-rust compatibility](https://godot-rust.github.io/book/toolchain/compatibility.html)
  and [Godot GDExtension packaging](https://docs.godotengine.org/en/stable/engine_details/engine_api/gdextension/gdextension_file.html)
  define the native extension boundary.
- [Godot export documentation](https://docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html)
  covers native build exports.
- [Godot AccessibilityServer](https://docs.godotengine.org/en/stable/classes/class_accessibilityserver.html)
  and [GUI navigation](https://docs.godotengine.org/en/stable/tutorials/ui/gui_navigation.html)
  support native accessibility and input design.
- [Apple notarization requirements](https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution)
  and [Microsoft SignTool](https://learn.microsoft.com/en-us/windows/win32/seccrypto/signtool)
  define native signing requirements.
- [cargo-dist configuration](https://axodotdev.github.io/cargo-dist/book/reference/config.html)
  is a possible release scaffold that must be pinned and reviewed before use.

## Language, scripts, and observer communication

- [Unicode Locale Data Markup Language](https://unicode.org/reports/tr35/)
  defines the current CLDR model for locale identifiers, numbers, units, dates,
  messages, and related data. The Lab stores semantic values independently and
  applies these conventions only at presentation and input boundaries.
- [Unicode Bidirectional Algorithm](https://www.unicode.org/reports/tr9/)
  defines bidirectional ordering behavior. Terminal and native surfaces must be
  tested with real right-to-left text, isolates, hostile control characters,
  cursor movement, selection, and mixed scientific notation.
- [Project Fluent](https://projectfluent.org/) is a candidate localization model
  because it lets translators control grammar and expression without assembling
  sentences from fragments. Adoption still requires dependency and terminal
  rendering review.
- Human language is one observer communication channel. Pressure, spectra,
  chemistry, fields, images, gestures, and unknown channels use the same
  capability and loss model without being forced through English prose.

## Artifacts, identity, and public remix

- The [OpenAI Terms of Use](https://openai.com/policies/terms-of-use/) state that,
  as between the user and OpenAI and to the extent permitted by law, the user
  owns output, while warning that output may not be unique and that the user
  remains responsible for it. Generated concepts still require prompt, input,
  similarity, rights, and public-release review; provider terms do not establish
  copyrightability or eliminate third-party risk.
- [QR-Bloom](https://github.com/physical-computation/qr-bloom) demonstrates a
  three-dimensional voxel tree whose top projection remains scannable. Its
  repository also credits source shapes with restrictions, so F.A.R.T. Lab uses
  only the high-level constraint as inspiration and copies no code, weights,
  meshes, or assets.
- [IBM Plex](https://github.com/IBM/plex) supplies the candidate Sans, Mono, and
  Math family under the SIL Open Font License. Fonts enter releases only through
  an explicit asset and license manifest.
- [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0) does not grant
  permission to use licensor names or marks except for reasonable descriptive
  use. Source licensing and trademark permission therefore remain separate.
- [USPTO trademark search guidance](https://www.uspto.gov/trademarks/search)
  supports a real clearance process before locking names, marks, merchandise,
  store listings, or implied exclusivity. A repository search is not legal
  clearance.
