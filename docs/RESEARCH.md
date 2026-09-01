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
- The CLI is the first complete product, the TUI is the second, and the native
  Godot app begins only after both are excellent.

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
