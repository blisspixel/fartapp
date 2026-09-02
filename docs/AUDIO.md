# Audio, Symphony, and in-world radio

This document defines the planned audio system. It keeps one promise simple:

> Every emission is generated from a versioned scenario, physical history,
> closure, observer, and seed. A deliberate replay recreates the same snowflake.

Two identical certified events may reproduce identical output. That is correct.
Uniqueness comes from the event, not from decorative randomness.

## Four separate audio lanes

| Lane | Purpose | Scientific status |
| --- | --- | --- |
| Physical acoustics | Sound and vibration derived from the event and its medium | A physical projection within stated model limits |
| Diagnostic sonification | Audible inspection of pressure, flow, uncertainty, and proof | A declared data mapping |
| Symphony Mode | A musical composition derived from event features | An artistic interpretation |
| In-world radio | Music for play, Broadcast, and watching episodes | An independent soundtrack |

Narration and dialogue use a fifth speech bus. Players can mix or mute every bus
independently. A radio track is never used as an emission, and Symphony Mode
never claims harmony is a law of fluid dynamics. The physical event must render
correctly when every music asset is absent.

## Symphony Mode

Symphony Mode turns the snowflake into an inspectable score. It reads normalized
features from the event graph rather than analyzing the final fart waveform.

Three listening modes keep the claim honest:

- **Evidence:** direct, repeatable sonification with fixed calibration.
- **Concert:** harmony, rhythm, timbre, and form use declared artistic rules.
- **Split:** physical acoustics, diagnostic cues, and orchestration remain
  separately audible and inspectable.

Useful mappings include:

| Event feature | Possible musical control |
| --- | --- |
| Interface oscillation | Continuous pitch contour, with confidence |
| Mass flow | Voice density or dynamics through a monotone mapping |
| Event duration | Phrase length |
| Strouhal behavior | Rhythmic subdivision |
| Turbulent fluctuation energy | Articulation or noise texture |
| Momentum direction | Spatial position |
| Droplet parcels | Pointillistic attacks |
| Deposition | Note decay or termination |
| Room modes | Resonant voices |
| Regime transitions | Formal or orchestral transitions |

Pitch is optional. An unvoiced or low-confidence physical signal reports no
pitch rather than inventing a note. Twelve-tone equal temperament and A440 are
Earth presets, not universal laws. Other profiles may use just intonation,
microtonal ratios, continuous pitch, pitchless percussion, or a compatible
observer-specific mapping. An incompatible world can return `UNTRANSLATABLE`.

A mapping declares its input, units, reference range, transform, clipping,
quantization, missing-value behavior, and information loss. Comparisons share a
fixed calibration. Event-local normalization is allowed only as a clearly
artistic effect because it can make physically different events sound similar.

Initial experiences include:

- **Snowflake Etude:** one event becomes a short score.
- **Comparative Canon:** a source event and its branch become separate voices.
- **Similarity Fugue:** similar nondimensional behavior preserves a motif across
  scales.
- **Conservation Concerto:** subsystem voices reconcile only when their ledgers
  close.
- **Universal Orchestra:** incompatible channels produce rests or an explicit
  translation failure.

The canonical Symphony artifact is a small semantic event list. WAV and text
are initial renders. MIDI can follow when arbitrary tuning support is mature.

## In-world radio

Radio exists so Quick Play and Broadcast are genuinely pleasant to watch. It is
music first and worldbuilding second. The concept can remain unnoticed for a
whole track.

The first station families are:

- **Drift 93.7:** chill electronica, downtempo, and late-night orbital music.
- **Night Side 106.1:** deep house, melodic EDM, and city-after-dark energy.
- **The Local Medium 88.4:** alternative hip-hop and neo-soul with original,
  restrained lyrics about pressure, distance, infrastructure, awkward silence,
  and life in the source world.
- **Scale Free FM:** dub techno and minimal electronics whose motifs reappear at
  different scales.
- **Public Resonance:** ambient orchestral and chamber minimalism from a very
  serious interdimensional broadcasting office.

Lyrics must work as lyrics without the premise. No constant fart references,
fake dialect, novelty-song delivery, or exposition dump. Lore appears through a
place name, civic rule, weather report, transit line, scientific metaphor, or
one understated line. Hip-hop is treated as music with history and craft, not a
bundle of stereotypes.

The normal now-playing display remains deliberately small:

```text
Drift 93.7
Pressure Weather
```

Lyrics and captions are available on request. Production metadata does not
invade the listening interface.

Optional station IDs, transitions, weather, traffic, and very short host breaks
build the source-world illusion between tracks. Music remains the majority of
airtime. A host can misinterpret an event only as a clearly diegetic opinion;
measured claims still come from immutable facts. Host-free and music-only modes
are complete experiences.

A station schedule is deterministic from a station-pack revision and a
presentation seed. Tuning, skipping, volume, lyrics, vocals, or host controls
change presentation only. They cannot change physics, story facts, challenge
scores, or agent observations outside the audio presentation.

Radio defaults to opt-in in terminal interfaces. The native app may remember an
explicit player preference. Music never carries the only copy of gameplay
information.

## Developer music pipeline

Eleven Music is a development-time authoring tool, never a runtime dependency.
The game ships approved audio bytes and works offline. No player, agent, build,
or replay needs an ElevenLabs account.

The planned developer surface is deliberately separate from `fart`:

```console
cargo xtask radio brief new drift-93-7
cargo xtask radio plan drift-93-7/pressure-weather.toml --model music_v2
cargo xtask radio generate drift-93-7/pressure-weather.toml \
  --max-minutes 5 --max-cost-usd 1.00 --candidate-dir .radio-work
cargo xtask radio review .radio-work/CANDIDATE_ID
cargo xtask radio approve .radio-work/CANDIDATE_ID --catalog assets/radio
```

These names are a contract proposal, not implemented commands. Generation
requires an interactive confirmation showing model, requested minutes, and
maximum spend. Approval is a different command so an API response cannot publish
itself. Candidate work remains ignored until editorial, rights, similarity,
lyrics, loudness, and packaging review pass.

The current production path targets `music_v2` and a reviewed composition plan:

1. Write a short station brief and the musical goal.
2. Author a composition plan without real artist, song, album, label, or copied
   lyric references.
3. Validate duration, spend limit, and rights status locally.
4. Generate a candidate manually with `ELEVENLABS_API_KEY` read only from the
   environment.
5. Listen without the prompt or joke visible and reject music that is not good
   on its own.
6. Review lyrics, similarity, cultural framing, and long-session fatigue.
7. Master an approved track locally and freeze its bytes and hash.
8. Package only the approved derivative used by the game.

The first production tranche is small: enough instrumental material for Drift
93.7, then Night Side 106.1, then a carefully reviewed Local Medium 88.4 vocal
set. A track passes only if blind listening says it works as music, its station
identity is recognizable without exposition, it survives repeated low-volume
play, transitions cleanly, and its lore remains optional. More tracks are not a
substitute for better tracks.

The detailed generation endpoint is preferred because it returns audio and the
resolved plan together. Provider seeds improve consistency but do not guarantee
exact regeneration after service updates, so accepted audio bytes are the source
of truth. Initial tracks are capped at five minutes even though the composition
plan reference currently permits longer work, because current product pages and
request schemas disagree about the maximum.

The public track record stays small:

```text
station_id
track_id
title
audio_hash
duration
rights_status
```

A private generation receipt may retain the sanitized request, model, plan,
date, response identifier, cost, and source hash. It never contains credentials,
account details, or local paths.

The repository owner has chosen Apache License 2.0 for repository-owned source,
documentation, and approved project media, including approved tracks.
Music-service terms are separate and can change. Release tooling therefore
requires the release owner to attest distribution and sublicensing rights before
a generated track enters a public repository or game package. This is a quiet
release check, not player-facing attribution.

If GitHub generation is added later, it uses an encrypted
`ELEVENLABS_API_KEY`, a protected environment, manual dispatch, explicit maximum
minutes and dollars, and no execution on pull requests or forks. Ordinary CI
never spends money or calls the service.

## Agent listening

Agents receive the same timed audio state through the play service without
having to pretend they heard PCM. A compact observation can include physical
propagation status, optional pitch and confidence, active sonification cues,
Symphony motif, station name, track title, playback position, and synchronized
lyrics or speech.

Raw audio is an optional bounded resource for audio-capable agents. Semantic
values come directly from the event and score graph, not from a separate model
guessing at the waveform. Challenges can deliberately compare structured-only,
audio-only, visual-only, and multimodal divisions.

## Safety, accessibility, and quality gates

- Extreme simulated pressure is never mapped directly to headphone amplitude.
- Scientific pressure is preserved before monitor gain, compression, and safety
  limiting.
- Physical acoustics, Symphony, radio, and speech have independent controls.
- Lyrics and meaningful non-speech audio have synchronized text alternatives.
- Music-only, host-free, vocals-off, reduced-dynamic-range, and audio-disabled
  play remain complete.
- Physical audio works with the radio catalog missing.
- Radio and Symphony changes cannot alter physical-result identity.
- A no-medium path produces no exterior physical sound while clearly labeled
  diagnostic sonification may remain available.
- Mappings advertised as monotone are property-tested as monotone.
- Musical time stretching preserves event order.
- Listening tests evaluate quality, fatigue, transitions, intelligibility, and
  whether the lore stays subtle.

The research sources and current provider constraints are collected in
[RESEARCH.md](RESEARCH.md).
