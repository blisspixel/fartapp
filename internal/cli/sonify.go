package cli

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/blisspixel/fartapp/internal/walkcase"
)

const (
	sonifyMappingID      = "diagnostic.flow-pitch"
	sonifyMappingVersion = "v0alpha1"
	sonifyNonclaim       = "diagnostic sonification, not physical acoustics"
	sonifySchema         = "fart.diagnostic-sonification/v0alpha1"
	defaultDurationScale = 20.0
	defaultSampleRate    = 48000
	defaultGainDBFS      = -18.0
)

const sonifyHelp = `Diagnostic sonification of one walk account.

Usage:
  fartapp walk sonify <case.json|-> --output <out.wav> [options]

Options:
  --output <out.wav>       Required WAV path. An existing file is replaced.
  --duration-scale <n>     Multiply history elapsed time. Default 20.
  --sample-rate <hz>       8000 to 96000. Default 48000.
  --gain-dbfs <db>         Peak ceiling from -60 to 0. Default -18.
  --play                   Play with pw-play, paplay, or afplay when present.
  --format text|json       English summary or the mapping report.
  -h, --help               Show this help without reading a case.

Mapping diagnostic.flow-pitch/v0alpha1:
  mass flow sets amplitude, restriction pressure ratio sets pitch, and
  choking switches in a brighter timbre. Time stretch is the duration scale.
  Output is at most 10 seconds, with a short fade at each end.

diagnostic sonification, not physical acoustics

Example:
  fartapp walk sonify testdata/walk/ordinary-low-pressure.json --output out.wav
`

type sonifySettings struct {
	sampleRate    int
	durationScale float64
	gainDBFS      float64
}

type sonifyReport struct {
	Schema            string   `json:"schema"`
	Status            string   `json:"status"`
	MappingID         string   `json:"mapping_id"`
	MappingVersion    string   `json:"mapping_version"`
	SampleRate        int      `json:"sample_rate"`
	DurationScale     float64  `json:"duration_scale"`
	GainDBFS          float64  `json:"gain_dbfs"`
	PeakDBFS          float64  `json:"peak_dbfs"`
	OutputDurationS   float64  `json:"output_duration_s"`
	HistoryElapsedS   float64  `json:"history_elapsed_s"`
	DurationClamped   bool     `json:"duration_clamped"`
	ChokedStartS      *float64 `json:"choked_start_s,omitempty"`
	ChokedEndS        *float64 `json:"choked_end_s,omitempty"`
	ChokedStartSample *int     `json:"choked_start_sample,omitempty"`
	ChokedEndSample   *int     `json:"choked_end_sample,omitempty"`
	Nonclaims         []string `json:"nonclaims"`
	Output            string   `json:"output,omitempty"`
	Player            string   `json:"player,omitempty"`
}

func defaultSonify() sonifySettings {
	return sonifySettings{
		sampleRate:    defaultSampleRate,
		durationScale: defaultDurationScale,
		gainDBFS:      defaultGainDBFS,
	}
}

func runSonify(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	play, args, err := stripFlag(args, "--play")
	if err != nil {
		writeDiagnostic(stderr, "invalid walk sonify: %v\n", err)
		return 1
	}
	options, values, err := parseValuedOptions(args, "--output", "--duration-scale", "--sample-rate", "--gain-dbfs")
	if err != nil {
		writeDiagnostic(stderr, "invalid walk sonify: %v\n", err)
		return 1
	}
	if options.help {
		return writeText(stdout, stderr, sonifyHelp)
	}
	if len(options.positional) != 1 {
		writeDiagnostic(stderr, "usage: fartapp walk sonify <case.json|-> --output <out.wav>\n")
		return 1
	}
	destination := values["--output"]
	if destination == "" {
		writeDiagnostic(stderr, "invalid walk sonify: --output requires a WAV path\n")
		return 1
	}
	settings := defaultSonify()
	if raw, ok := values["--duration-scale"]; ok {
		settings.durationScale, err = parseFinite("--duration-scale", raw)
		if err != nil || settings.durationScale <= 0 || settings.durationScale > 1_000_000 {
			writeDiagnostic(stderr, "invalid walk sonify: --duration-scale must be a finite number in (0, 1000000]\n")
			return 1
		}
	}
	if raw, ok := values["--sample-rate"]; ok {
		rate, convErr := strconv.Atoi(raw)
		if convErr != nil || strconv.Itoa(rate) != raw || rate < 8000 || rate > 96000 {
			writeDiagnostic(stderr, "invalid walk sonify: --sample-rate must be an integer from 8000 to 96000\n")
			return 1
		}
		settings.sampleRate = rate
	}
	if raw, ok := values["--gain-dbfs"]; ok {
		settings.gainDBFS, err = parseFinite("--gain-dbfs", raw)
		if err != nil || settings.gainDBFS < -60 || settings.gainDBFS > 0 {
			writeDiagnostic(stderr, "invalid walk sonify: --gain-dbfs must be from -60 to 0\n")
			return 1
		}
	}
	report := readWalk(options.positional[0], "simulate", stdin)
	if !report.Predicted() {
		return writeWalkReport(report, "sonify", options.format, stdout, stderr)
	}
	wav, meta, err := synthesize(report, settings)
	if err != nil {
		writeDiagnostic(stderr, "walk sonify failed: %s\n", strconv.QuoteToASCII(boundDiagnostic(err.Error())))
		return 1
	}
	if err := os.WriteFile(destination, wav, 0o644); err != nil {
		writeDiagnostic(stderr, "walk sonify failed: %s\n", strconv.QuoteToASCII(boundDiagnostic(err.Error())))
		return 1
	}
	meta.Output = destination
	if play {
		player, code := playWAV(destination, stderr)
		meta.Player = player
		if writeSonify(meta, options.format, stdout, stderr) != 0 {
			return 1
		}
		return code
	}
	return writeSonify(meta, options.format, stdout, stderr)
}

func writeSonify(meta sonifyReport, format outputFormat, stdout, stderr io.Writer) int {
	return writeValue(stdout, stderr, format, meta, formatSonify)
}

func formatSonify(meta sonifyReport) string {
	var output strings.Builder
	fmt.Fprintf(&output, "DIAGNOSTIC SONIFICATION\n\nMapping: %s/%s\n%s\n", meta.MappingID, meta.MappingVersion, sonifyNonclaim)
	fmt.Fprintf(&output, "Sample rate: %d Hz\nDuration scale: %s\nGain: %s dBFS\nPeak: %s dBFS\n",
		meta.SampleRate, formatScientificValue(meta.DurationScale), formatScientificValue(meta.GainDBFS), formatScientificValue(meta.PeakDBFS))
	fmt.Fprintf(&output, "Output duration: %s s\nHistory elapsed: %s s\nDuration clamped: %t\n",
		formatScientificValue(meta.OutputDurationS), formatScientificValue(meta.HistoryElapsedS), meta.DurationClamped)
	if meta.ChokedStartSample != nil && meta.ChokedEndSample != nil {
		fmt.Fprintf(&output, "Choked samples: %d-%d\n", *meta.ChokedStartSample, *meta.ChokedEndSample)
	}
	if meta.Output != "" {
		fmt.Fprintf(&output, "Output: %s\n", meta.Output)
	}
	if meta.Player == "not-found" {
		output.WriteString("Player: none of pw-play, paplay, or afplay is on PATH\n")
	} else if meta.Player != "" {
		fmt.Fprintf(&output, "Player: %s\n", meta.Player)
	}
	return output.String()
}

func writeIntensityWAV(level intensity, path string, stderr io.Writer) int {
	data, err := presetBytes(presetForIntensity(int(level) + minimumIntensity))
	if err != nil {
		writeDiagnostic(stderr, "intensity wav: %s\n", strconv.QuoteToASCII(boundDiagnostic(err.Error())))
		return 1
	}
	report := walkcase.Run(data, "simulate")
	if !report.Predicted() {
		writeDiagnostic(stderr, "intensity wav: preset did not predict\n")
		return 1
	}
	wav, _, err := synthesize(report, defaultSonify())
	if err != nil {
		writeDiagnostic(stderr, "intensity wav: %s\n", strconv.QuoteToASCII(boundDiagnostic(err.Error())))
		return 1
	}
	if err := os.WriteFile(path, wav, 0o644); err != nil {
		writeDiagnostic(stderr, "intensity wav: %s\n", strconv.QuoteToASCII(boundDiagnostic(err.Error())))
		return 1
	}
	return 0
}

func synthesize(report walkcase.Report, settings sonifySettings) ([]byte, sonifyReport, error) {
	history := report.History
	if len(history) == 0 {
		if report.Initial == nil || report.InitialRestriction == nil {
			return nil, sonifyReport{}, errors.New("walk account has no history to sonify")
		}
		history = []walkcase.HistorySample{{
			TimeSeconds:                0,
			PressurePascals:            report.Initial.PressurePascals,
			MassFlowKilogramsPerSecond: report.InitialRestriction.MassFlowKilogramsPerS,
			Regime:                     report.InitialRestriction.Regime,
		}}
	}
	elapsed := 0.0
	if report.ElapsedSeconds != nil && *report.ElapsedSeconds > 0 {
		elapsed = *report.ElapsedSeconds
	}
	start := history[0].TimeSeconds
	span := history[len(history)-1].TimeSeconds - start
	if elapsed > span {
		span = elapsed
	}
	outputDuration := span * settings.durationScale
	clamped := false
	if outputDuration > 10 {
		outputDuration = 10
		clamped = true
	}
	if outputDuration < 1.0/float64(settings.sampleRate) {
		outputDuration = 0.25
	}
	count := int(math.Round(outputDuration * float64(settings.sampleRate)))
	if count < 2 {
		count = 2
	}
	if count > settings.sampleRate*10 {
		count = settings.sampleRate * 10
		clamped = true
	}
	outputDuration = float64(count) / float64(settings.sampleRate)
	back := 0.0
	if report.Initial != nil && report.InitialRestriction != nil && report.Initial.PressurePascals > 0 {
		back = report.InitialRestriction.BackPressureRatio * report.Initial.PressurePascals
	}
	maxFlow := 0.0
	for _, sample := range history {
		if sample.MassFlowKilogramsPerSecond > maxFlow {
			maxFlow = sample.MassFlowKilogramsPerSecond
		}
	}
	samples := make([]float64, count)
	var chokeStart, chokeEnd *int
	var chokeStartS, chokeEndS *float64
	phase := 0.0
	for index := 0; index < count; index++ {
		histT := start
		if count > 1 && span > 0 {
			histT = start + float64(index)/float64(count-1)*span
		}
		flow, pressure, choked := historyAt(history, histT)
		ratio := 0.0
		if pressure > 0 && back > 0 {
			ratio = back / pressure
		}
		if ratio < 0 {
			ratio = 0
		} else if ratio > 1 {
			ratio = 1
		}
		freq := 110 + 660*(1-ratio)
		amplitude := 0.0
		if maxFlow > 0 {
			amplitude = flow / maxFlow
		}
		tri := 1 - 4*math.Abs(phase-0.5)
		square := 1.0
		if phase >= 0.5 {
			square = -1
		}
		wave := tri
		if choked {
			wave = 0.7*tri + 0.3*square
			if chokeStart == nil {
				started, startedAt := index, histT
				chokeStart = &started
				chokeStartS = &startedAt
			}
			ended, endedAt := index, histT
			chokeEnd = &ended
			chokeEndS = &endedAt
		}
		samples[index] = wave * amplitude
		phase += freq / float64(settings.sampleRate)
		if phase >= 1 {
			phase--
		}
	}
	fade := min(64, count/10)
	if fade > 0 {
		for index := 0; index < fade; index++ {
			window := float64(index) / float64(fade)
			samples[index] *= window
			samples[count-1-index] *= window
		}
	}
	target := int(math.Floor(32767 * math.Pow(10, settings.gainDBFS/20)))
	if target < 0 {
		target = 0
	}
	if target > 32767 {
		target = 32767
	}
	peak := 0.0
	for _, sample := range samples {
		if abs := math.Abs(sample); abs > peak {
			peak = abs
		}
	}
	pcm := make([]int16, count)
	peakDB := -144.0
	peakAbs := 0
	if peak > 0 && target > 0 {
		scale := float64(target) / peak
		for index, sample := range samples {
			quantized := int(math.Round(sample * scale))
			if quantized > target {
				quantized = target
			} else if quantized < -target {
				quantized = -target
			}
			pcm[index] = int16(quantized)
			if abs := quantized; abs < 0 {
				abs = -abs
			} else if abs > peakAbs {
				peakAbs = abs
			}
			if quantized < 0 {
				if -quantized > peakAbs {
					peakAbs = -quantized
				}
			} else if quantized > peakAbs {
				peakAbs = quantized
			}
		}
		if peakAbs > 0 {
			peakDB = 20 * math.Log10(float64(peakAbs)/32767)
		}
	}
	meta := sonifyReport{
		Schema:            sonifySchema,
		Status:            "synthesized",
		MappingID:         sonifyMappingID,
		MappingVersion:    sonifyMappingVersion,
		SampleRate:        settings.sampleRate,
		DurationScale:     settings.durationScale,
		GainDBFS:          settings.gainDBFS,
		PeakDBFS:          peakDB,
		OutputDurationS:   outputDuration,
		HistoryElapsedS:   elapsed,
		DurationClamped:   clamped,
		ChokedStartS:      chokeStartS,
		ChokedEndS:        chokeEndS,
		ChokedStartSample: chokeStart,
		ChokedEndSample:   chokeEnd,
		Nonclaims:         []string{sonifyNonclaim},
	}
	return buildWAV(pcm, settings.sampleRate, sonifyMappingID+"/"+sonifyMappingVersion, sonifyNonclaim), meta, nil
}

func historyAt(history []walkcase.HistorySample, t float64) (flow float64, pressure float64, choked bool) {
	index := 0
	for cursor := 1; cursor < len(history); cursor++ {
		if history[cursor].TimeSeconds <= t {
			index = cursor
			continue
		}
		break
	}
	left := history[index]
	flow = left.MassFlowKilogramsPerSecond
	pressure = left.PressurePascals
	choked = left.Regime == "choked"
	if index+1 < len(history) {
		right := history[index+1]
		span := right.TimeSeconds - left.TimeSeconds
		if span > 0 && t > left.TimeSeconds {
			frac := (t - left.TimeSeconds) / span
			flow += (right.MassFlowKilogramsPerSecond - left.MassFlowKilogramsPerSecond) * frac
			pressure += (right.PressurePascals - left.PressurePascals) * frac
		}
	}
	if flow < 0 {
		flow = 0
	}
	return flow, pressure, choked
}

func buildWAV(pcm []int16, rate int, name, comment string) []byte {
	data := make([]byte, len(pcm)*2)
	for index, sample := range pcm {
		binary.LittleEndian.PutUint16(data[index*2:], uint16(sample))
	}
	info := append([]byte("INFO"), wavInfo("INAM", name)...)
	info = append(info, wavInfo("ICMT", comment)...)
	body := append([]byte("WAVE"), wavChunk("fmt ", wavFormat(rate))...)
	body = append(body, wavChunk("LIST", info)...)
	body = append(body, wavChunk("data", data)...)
	header := make([]byte, 8)
	copy(header, "RIFF")
	binary.LittleEndian.PutUint32(header[4:], uint32(len(body)))
	return append(header, body...)
}

func wavFormat(rate int) []byte {
	payload := make([]byte, 16)
	binary.LittleEndian.PutUint16(payload[0:], 1)
	binary.LittleEndian.PutUint16(payload[2:], 1)
	binary.LittleEndian.PutUint32(payload[4:], uint32(rate))
	binary.LittleEndian.PutUint32(payload[8:], uint32(rate*2))
	binary.LittleEndian.PutUint16(payload[12:], 2)
	binary.LittleEndian.PutUint16(payload[14:], 16)
	return payload
}

func wavInfo(tag, text string) []byte {
	return wavChunk(tag, append([]byte(text), 0))
}

func wavChunk(id string, payload []byte) []byte {
	chunk := make([]byte, 8+len(payload))
	copy(chunk[:4], id)
	binary.LittleEndian.PutUint32(chunk[4:8], uint32(len(payload)))
	copy(chunk[8:], payload)
	if len(payload)%2 == 1 {
		chunk = append(chunk, 0)
	}
	return chunk
}

func playWAV(path string, stderr io.Writer) (string, int) {
	var player string
	for _, name := range []string{"pw-play", "paplay", "afplay"} {
		found, err := exec.LookPath(name)
		if err == nil {
			player = found
			break
		}
	}
	if player == "" {
		return "not-found", 0
	}
	command := exec.Command(player, path)
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		writeDiagnostic(stderr, "play %s: %s\n", path, strconv.QuoteToASCII(boundDiagnostic(err.Error())))
		return player, 1
	}
	return player, 0
}

func stripFlag(args []string, flag string) (bool, []string, error) {
	seen := false
	ended := false
	kept := make([]string, 0, len(args))
	for _, arg := range args {
		if !ended && arg == "--" {
			ended = true
			kept = append(kept, arg)
			continue
		}
		if !ended && arg == flag {
			if seen {
				return false, nil, fmt.Errorf("%s may be specified only once", flag)
			}
			seen = true
			continue
		}
		kept = append(kept, arg)
	}
	return seen, kept, nil
}

func parseFinite(name, raw string) (float64, error) {
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("%s must be a finite number", name)
	}
	return value, nil
}

func intensityWavPath(args []string) (string, error) {
	var path string
	seen := false
	for index := 0; index < len(args); index++ {
		arg := args[index]
		var value string
		switch {
		case arg == "--wav":
			index++
			if index >= len(args) {
				return "", errors.New("missing")
			}
			value = args[index]
		case strings.HasPrefix(arg, "--wav="):
			value = strings.TrimPrefix(arg, "--wav=")
		default:
			return "", errors.New("extra")
		}
		if seen || value == "" {
			return "", errors.New("repeat")
		}
		seen = true
		path = value
	}
	return path, nil
}
