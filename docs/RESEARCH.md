# Research basis

This document records the authoritative sources behind the simulation, numerical,
gameplay, and interface plans. It is a design bibliography, not evidence that a
planned model has already been implemented or validated.

Sources were reviewed for this roadmap on 2026-09-01. Primary papers, standards,
government laboratories, and official project documentation are preferred.

## Decisions changed by the research

- The neutral ontology is emitter, interface, exterior, and law profile. Human
  anatomy and Earth physics are presets.
- Ordinary Earth-biological pressure excursions are far below the ideal-gas
  choking threshold. Choked, underexpanded, locally supersonic, and
  shock-containing are separate labels.
- Finite events begin as starting jets or puffs. Steady round-jet scaling is a
  conditional far-field limit.
- Coarse state history constrains procedural audio but does not uniquely contain
  unresolved turbulent waveform detail. The audio closure and seed are part of
  provenance.
- Wetness, breakup, and deposition require more than Weber or Stokes number.
- Similarity requires matching nondimensional equations, closures, dimensions,
  topology, normalized conditions, and active coefficients.
- Conservation, code verification, solution verification, empirical validation,
  and fictional-law consistency are separate certificate claims.
- Alternate dimensions and universes get explicit law packs. Unsupported laws
  fail explicitly rather than reusing Earth fluid equations.
- The story director reacts to immutable event facts and never rerolls a valid
  outcome.
- Physical acoustics, diagnostic sonification, Symphony, and radio are separate
  audio lanes with different truth claims and tests.
- Social meaning is represented through situated perspectives, institutions,
  disagreement, and power, never one generated `culture.norms` fact.
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
| Derived artifacts need traceable provenance | [W3C PROV-DM](https://www.w3.org/TR/prov-dm/) defines entities, activities, derivations, and constraints | Use a typed event graph with versioned derivation edges | What minimal subset stays pleasant in a CLI? |
| Units can be part of program correctness | [Kennedy on units of measure](https://www.microsoft.com/en-us/research/publication/relational-parametricity-and-units-of-measure/) proves dimensional invariance properties for a typed language | Make invalid dimensional combinations difficult to represent | Which quantities need runtime profiles rather than compile-time dimensions? |
| Prediction before explanation can improve transfer | [Kapur's controlled studies](https://doi.org/10.1111/cogs.12107) concern mathematics learning | Let players predict, fail productively, then inspect and transfer | Does this effect hold for short fluid-dynamics play sessions? |
| Complex model conclusions need sensitivity analysis | [Saltelli et al.](https://doi.org/10.1002/9780470725184) covers global sensitivity methods | Report influential inputs, interactions, and nonidentifiability | Which methods fit interactive latency and mixed variables? |
| Sonification must be perceptually evaluated | The [NSF sonification report](https://digitalcommons.unl.edu/psychfacpub/444/) calls for control, interchange, and perceptual testing | Treat sound as both consequence and auditory display, with separate tests | Which mappings teach without ruining the joke or accessibility? |
| Generator robustness does not establish entertainment | The [PCG literature](https://www.pcgbook.com/) provides generation and evaluation methods, not a universal fun metric | Pair seed sweeps with a curated corpus and human playtests | Which repetition and pacing measures predict shareability? |
| Protocol parity requires an application contract | [MCP 2026-07-28](https://modelcontextprotocol.io/specification/2026-07-28) is stateless at its core and [A2A 1.0](https://a2a-protocol.org/latest/specification/) models longer agent tasks | Keep play handles, state, actions, knowledge, and replay in `PlayService`, not transport sessions | What smallest adapter set remains pleasant for unrelated agents? |
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
- [Ambulatory anorectal pressure study](https://pubmed.ncbi.nlm.nih.gov/3219524/)
  provides scale evidence for ordinary biological pressure excursions.
- [Measured flatus composition study](https://pubmed.ncbi.nlm.nih.gov/9176210/)
  shows substantial event and subject variability in major and trace gases.

## Acoustics, interfaces, and rooms

- [Lighthill: On Sound Generated Aerodynamically](https://doi.org/10.1098/rspa.1952.0060)
  is the foundation for deriving aerodynamic sound sources from flow.
- [Titze: compliant biological-interface oscillation](https://pubmed.ncbi.nlm.nih.gov/3372869/)
  provides a mathematical analogue for pressure-flow-structure coupling without
  implying identical anatomy.
- [Single-mass vocal-fold fluid-structure review](https://pmc.ncbi.nlm.nih.gov/articles/PMC2857605/)
  explains why self-sustained oscillation needs aerodynamic feedback.
- [Allen and Berkley: image-source rooms](https://doi.org/10.1121/1.382599)
  provides deterministic rectangular-room impulse-response modeling.
- [NASA: sound and the vacuum limit](https://science.nasa.gov/ems/02_anatomy/)
  supports the requirement for a material medium for external sound.

## Droplets, particles, bubbles, and vacuum

- [NASA droplet-breakup memorandum](https://ntrs.nasa.gov/api/citations/20070034950/downloads/20070034950.pdf)
  documents Weber, Ohnesorge, breakup-mode, and breakup-time dependencies.
- [Hinze: breakup in dispersion processes](https://doi.org/10.1002/aic.690010303)
  is a foundational source for turbulent breakup criteria and viscosity effects.
- [Sandia aerosol transport and deposition report](https://www.osti.gov/servlets/purl/1675151)
  connects response time, Stokes behavior, settling, inertia, diffusion, and
  deposition mechanisms.
- [Brennen: Cavitation and Bubble Dynamics](https://media.library.caltech.edu/CaltechBOOK%3A1995.001/chap2.htm)
  derives Rayleigh-Plesset dynamics and states their assumptions.
- [Minnaert: musical air bubbles](https://doi.org/10.1080/14786443309462277)
  provides the classic isolated-bubble resonance model.
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
  the International System of Units used by the first Earth law profile.
- [Barth and Ohlberger: finite-volume methods](https://ntrs.nasa.gov/citations/20030020790)
  provides a foundation for conservative discretization.
- [Zhang and Shu: positivity-preserving Euler schemes](https://doi.org/10.1016/j.jcp.2010.08.016)
  supports positive density and pressure under explicit numerical conditions.
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

## Agent environments and interoperability

- The [MCP 2026-07-28 release](https://blog.modelcontextprotocol.io/posts/2026-07-28/)
  removes the initialize handshake and protocol session, adds discovery and
  header routing, moves long-running Tasks into an extension, and deprecates
  Roots, Sampling, Logging, and legacy HTTP+SSE. Stateful games should return
  explicit application handles.
- The official [MCP conformance framework](https://github.com/modelcontextprotocol/conformance)
  validates messages and scenarios for declared protocol revisions. Its
  2026-07-28 coverage and open harness issues must be pinned and reported rather
  than hidden in a permanent expected-failure baseline.
- The [A2A 1.0 specification](https://a2a-protocol.org/latest/specification/)
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

## Culture, religion, and public interest

- [Harvard's Religious Literacy Project](https://rpl.hds.harvard.edu/what-we-do/our-approach/core-principles)
  treats religions as internally diverse, changing, embedded in culture, and
  shaped by power. A world therefore cannot receive one deterministic religious
  or moral rule from its environment.
- [UNESCO's ethical principles for intangible cultural heritage](https://ich.unesco.org/en/ethics-and-ich-00866)
  emphasize community participation, diversity, consent, and customary access.
  Sacred or private contexts are boundaries, not random comic props.
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
- [Accessible Player Experiences](https://accessible.games/accessible-player-experiences/)
  frames accessibility through concrete player experience patterns.

## CLI, archives, terminal UI, and native delivery

- [Command Line Interface Guidelines](https://clig.dev/) supports a human-first,
  composable, predictable command experience.
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
