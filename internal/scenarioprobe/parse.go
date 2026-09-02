package scenarioprobe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/blisspixel/fartapp/internal/lawcatalog"
)

const (
	maximumJSONDepth          = 32
	maximumCapabilityRequests = 64
	maximumMemberNameBytes    = 128
)

func Validate(data []byte) Report {
	if len(data) > MaxInputBytes {
		return Failure(Diagnostic{
			Code:       "FART-E-INPUT-0001",
			Stage:      "input",
			Path:       "/",
			ReasonCode: "input_too_large",
		})
	}
	if diagnostic := preflightJSON(data); diagnostic != nil {
		return Failure(*diagnostic)
	}
	document, diagnostic := parseDocument(data)
	if diagnostic != nil {
		return Failure(*diagnostic)
	}
	return resolveDocument(document)
}

func resolveDocument(document Document) Report {
	context := document.LawContextSet.Contexts[0]
	inspection, resolution := lawcatalog.Resolve(context.ID + "@" + context.Version)
	if resolution != lawcatalog.ResolutionFound {
		return catalogFailure(Diagnostic{
			Code:       "FART-E-LAW-0001",
			Stage:      "law-resolution",
			Path:       "/law_context_set/contexts/0",
			ReasonCode: "law_context_not_found",
		})
	}

	capabilities := make([]CapabilityResult, 0, len(document.CapabilityRequests))
	evidenceIDs := make(map[string]struct{})
	for index, request := range document.CapabilityRequests {
		capability, found := findCapability(inspection.CapabilityReport.Capabilities, request.ID)
		if !found {
			return catalogFailure(Diagnostic{
				Code:       "FART-E-CAP-0001",
				Stage:      "capability-resolution",
				Path:       fmt.Sprintf("/capability_requests/%d/id", index),
				ReasonCode: "capability_not_defined",
			})
		}
		capabilities = append(capabilities, capabilityResult(capability))
		for _, reference := range capability.EvidenceReferences {
			evidenceIDs[reference] = struct{}{}
		}
	}
	sort.Slice(capabilities, func(left, right int) bool {
		return capabilities[left].ID < capabilities[right].ID
	})

	report := baseReport("document_bytes", "built_in_law_catalog")
	report.DocumentStatus = "valid"
	report.ValidationStages = successfulValidationStages()
	report.DocumentSchema = document.Schema
	report.LawContext = &lawcatalog.LawContextRef{ID: context.ID, Version: context.Version}
	report.Scope = &Scope{ID: document.Scope.ID}
	report.Capabilities = capabilities
	for _, evidence := range inspection.CapabilityReport.EvidenceRegistry {
		if _, included := evidenceIDs[evidence.ID]; included {
			report.EvidenceRegistry = append(report.EvidenceRegistry, evidence)
		}
	}
	return report
}

func findCapability(capabilities []lawcatalog.Capability, id string) (lawcatalog.Capability, bool) {
	for _, capability := range capabilities {
		if capability.ID == id {
			return capability, true
		}
	}
	return lawcatalog.Capability{}, false
}

func parseDocument(data []byte) (Document, *Diagnostic) {
	root, diagnostic := decodeObject(data, "")
	if diagnostic != nil {
		return Document{}, diagnostic
	}
	schemaRaw, exists := root["schema"]
	if !exists {
		return Document{}, schemaDiagnostic("/schema", "missing_member")
	}
	schema, diagnostic := decodeString(schemaRaw, "/schema")
	if diagnostic != nil {
		return Document{}, diagnostic
	}
	if schema != DocumentSchema {
		return Document{}, schemaDiagnostic("/schema", "unsupported_schema")
	}
	if diagnostic = validateMembers(
		root,
		"",
		[]string{"schema", "law_context_set", "scope", "capability_requests"},
	); diagnostic != nil {
		return Document{}, diagnostic
	}
	lawContextSet, diagnostic := parseLawContextSet(root["law_context_set"])
	if diagnostic != nil {
		return Document{}, diagnostic
	}
	scope, diagnostic := parseScope(root["scope"])
	if diagnostic != nil {
		return Document{}, diagnostic
	}
	requests, diagnostic := parseCapabilityRequests(root["capability_requests"])
	if diagnostic != nil {
		return Document{}, diagnostic
	}
	if lawContextSet.Contexts[0].ScopeID != scope.ID {
		return Document{}, schemaDiagnostic(
			"/law_context_set/contexts/0/scope_id",
			"scope_reference_unresolved",
		)
	}
	return Document{
		Schema:             schema,
		LawContextSet:      lawContextSet,
		Scope:              scope,
		CapabilityRequests: requests,
	}, nil
}

func parseLawContextSet(raw json.RawMessage) (LawContextSet, *Diagnostic) {
	object, diagnostic := decodeObject(raw, "/law_context_set")
	if diagnostic != nil {
		return LawContextSet{}, diagnostic
	}
	if diagnostic = validateMembers(object, "/law_context_set", []string{"contexts"}); diagnostic != nil {
		return LawContextSet{}, diagnostic
	}
	items, diagnostic := decodeArray(object["contexts"], "/law_context_set/contexts")
	if diagnostic != nil {
		return LawContextSet{}, diagnostic
	}
	if len(items) != 1 {
		if len(items) == 0 {
			return LawContextSet{}, schemaDiagnostic(
				"/law_context_set/contexts",
				"missing_law_context",
			)
		}
		return LawContextSet{}, &Diagnostic{
			Code:       "FART-E-ONTOLOGY-0001",
			Stage:      "schema",
			Path:       "/law_context_set/contexts",
			ReasonCode: "multi_law_not_supported",
		}
	}
	context, diagnostic := parseScopedLawContext(items[0], "/law_context_set/contexts/0")
	if diagnostic != nil {
		return LawContextSet{}, diagnostic
	}
	return LawContextSet{Contexts: []ScopedLawContext{context}}, nil
}

func parseScopedLawContext(raw json.RawMessage, path string) (ScopedLawContext, *Diagnostic) {
	object, diagnostic := decodeObject(raw, path)
	if diagnostic != nil {
		return ScopedLawContext{}, diagnostic
	}
	if diagnostic = validateMembers(object, path, []string{"id", "version", "scope_id"}); diagnostic != nil {
		return ScopedLawContext{}, diagnostic
	}
	id, diagnostic := decodeToken(object["id"], path+"/id")
	if diagnostic != nil {
		return ScopedLawContext{}, diagnostic
	}
	version, diagnostic := decodeToken(object["version"], path+"/version")
	if diagnostic != nil {
		return ScopedLawContext{}, diagnostic
	}
	scopeID, diagnostic := decodeToken(object["scope_id"], path+"/scope_id")
	if diagnostic != nil {
		return ScopedLawContext{}, diagnostic
	}
	return ScopedLawContext{ID: id, Version: version, ScopeID: scopeID}, nil
}

func parseScope(raw json.RawMessage) (Scope, *Diagnostic) {
	object, diagnostic := decodeObject(raw, "/scope")
	if diagnostic != nil {
		return Scope{}, diagnostic
	}
	if diagnostic = validateMembers(object, "/scope", []string{"id"}); diagnostic != nil {
		return Scope{}, diagnostic
	}
	id, diagnostic := decodeToken(object["id"], "/scope/id")
	if diagnostic != nil {
		return Scope{}, diagnostic
	}
	return Scope{ID: id}, nil
}

func parseCapabilityRequests(raw json.RawMessage) ([]CapabilityRequest, *Diagnostic) {
	items, diagnostic := decodeArray(raw, "/capability_requests")
	if diagnostic != nil {
		return nil, diagnostic
	}
	if len(items) == 0 {
		return nil, schemaDiagnostic("/capability_requests", "missing_capability_request")
	}
	if len(items) > maximumCapabilityRequests {
		return nil, schemaDiagnostic("/capability_requests", "collection_limit_exceeded")
	}
	result := make([]CapabilityRequest, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for index, item := range items {
		path := fmt.Sprintf("/capability_requests/%d", index)
		object, itemDiagnostic := decodeObject(item, path)
		if itemDiagnostic != nil {
			return nil, itemDiagnostic
		}
		if itemDiagnostic = validateMembers(object, path, []string{"id"}); itemDiagnostic != nil {
			return nil, itemDiagnostic
		}
		id, itemDiagnostic := decodeToken(object["id"], path+"/id")
		if itemDiagnostic != nil {
			return nil, itemDiagnostic
		}
		if _, exists := seen[id]; exists {
			return nil, schemaDiagnostic(path+"/id", "duplicate_capability_request")
		}
		seen[id] = struct{}{}
		result = append(result, CapabilityRequest{ID: id})
	}
	return result, nil
}

func validateMembers(
	object map[string]json.RawMessage,
	path string,
	required []string,
) *Diagnostic {
	allowed := make(map[string]struct{}, len(required))
	for _, name := range required {
		allowed[name] = struct{}{}
		if _, exists := object[name]; !exists {
			return schemaDiagnostic(joinJSONPointer(path, name), "missing_member")
		}
	}
	unknown := make([]string, 0)
	for name := range object {
		if _, exists := allowed[name]; !exists {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) != 0 {
		sort.Strings(unknown)
		return schemaDiagnostic(joinJSONPointer(path, unknown[0]), "unknown_member")
	}
	return nil
}

func decodeObject(raw []byte, path string) (map[string]json.RawMessage, *Diagnostic) {
	trimmed := trimJSONSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, schemaDiagnostic(pathOrRoot(path), "wrong_type")
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return nil, syntaxDiagnostic("malformed_json", 0)
	}
	return value, nil
}

func decodeArray(raw []byte, path string) ([]json.RawMessage, *Diagnostic) {
	trimmed := trimJSONSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, schemaDiagnostic(path, "wrong_type")
	}
	var value []json.RawMessage
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return nil, syntaxDiagnostic("malformed_json", 0)
	}
	return value, nil
}

func decodeString(raw []byte, path string) (string, *Diagnostic) {
	trimmed := trimJSONSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '"' {
		return "", schemaDiagnostic(path, "wrong_type")
	}
	var value string
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return "", syntaxDiagnostic("malformed_json", 0)
	}
	return value, nil
}

func decodeToken(raw []byte, path string) (string, *Diagnostic) {
	value, diagnostic := decodeString(raw, path)
	if diagnostic != nil {
		return "", diagnostic
	}
	if err := lawcatalog.ValidateMachineToken(value); err != nil {
		return "", schemaDiagnostic(path, "invalid_token")
	}
	return value, nil
}

func preflightJSON(data []byte) *Diagnostic {
	if len(trimJSONSpace(data)) == 0 {
		return syntaxDiagnostic("empty_input", 0)
	}
	if !utf8.Valid(data) || !validJSONUnicodeEscapes(data) {
		return syntaxDiagnostic("malformed_json", 0)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if diagnostic := scanJSONValue(decoder, "", 0); diagnostic != nil {
		return diagnostic
	}
	offset := decoder.InputOffset()
	if offset < int64(len(data)) && len(trimJSONSpace(data[offset:])) != 0 {
		return syntaxDiagnostic("trailing_json_value", offset)
	}
	return nil
}

func trimJSONSpace(value []byte) []byte {
	start := 0
	for start < len(value) && isJSONSpace(value[start]) {
		start++
	}
	end := len(value)
	for end > start && isJSONSpace(value[end-1]) {
		end--
	}
	return value[start:end]
}

func isJSONSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

func validJSONUnicodeEscapes(data []byte) bool {
	inString := false
	for index := 0; index < len(data); index++ {
		switch data[index] {
		case '"':
			inString = !inString
		case '\\':
			if !inString || index+1 >= len(data) {
				continue
			}
			if data[index+1] != 'u' {
				index++
				continue
			}
			value, valid := decodeHexQuad(data, index+2)
			if !valid {
				continue
			}
			if value >= 0xdc00 && value <= 0xdfff {
				return false
			}
			if value >= 0xd800 && value <= 0xdbff {
				if index+12 > len(data) || data[index+6] != '\\' || data[index+7] != 'u' {
					return false
				}
				low, lowValid := decodeHexQuad(data, index+8)
				if !lowValid || low < 0xdc00 || low > 0xdfff {
					return false
				}
				index += 11
				continue
			}
			index += 5
		}
	}
	return true
}

func decodeHexQuad(data []byte, start int) (uint16, bool) {
	if start+4 > len(data) {
		return 0, false
	}
	var value uint16
	for _, character := range data[start : start+4] {
		value <<= 4
		switch {
		case character >= '0' && character <= '9':
			value += uint16(character - '0')
		case character >= 'a' && character <= 'f':
			value += uint16(character-'a') + 10
		case character >= 'A' && character <= 'F':
			value += uint16(character-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

func scanJSONValue(decoder *json.Decoder, path string, depth int) *Diagnostic {
	if depth > maximumJSONDepth {
		return schemaDiagnostic(pathOrRoot(path), "maximum_nesting_exceeded")
	}
	token, err := decoder.Token()
	if err != nil {
		return syntaxDiagnostic("malformed_json", decoder.InputOffset())
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, keyErr := decoder.Token()
			if keyErr != nil {
				return syntaxDiagnostic("malformed_json", decoder.InputOffset())
			}
			key, ok := keyToken.(string)
			if !ok {
				return syntaxDiagnostic("malformed_json", decoder.InputOffset())
			}
			if len(key) > maximumMemberNameBytes {
				return schemaDiagnostic(pathOrRoot(path), "member_name_too_long")
			}
			childPath := joinJSONPointer(path, key)
			if _, exists := seen[key]; exists {
				return &Diagnostic{
					Code:       "FART-E-SCHEMA-0002",
					Stage:      "schema",
					Path:       childPath,
					ReasonCode: "duplicate_member",
					ByteOffset: decoder.InputOffset(),
				}
			}
			seen[key] = struct{}{}
			if diagnostic := scanJSONValue(decoder, childPath, depth+1); diagnostic != nil {
				return diagnostic
			}
		}
		if _, err = decoder.Token(); err != nil {
			return syntaxDiagnostic("malformed_json", decoder.InputOffset())
		}
	case '[':
		index := 0
		for decoder.More() {
			childPath := joinJSONPointer(path, strconv.Itoa(index))
			if diagnostic := scanJSONValue(decoder, childPath, depth+1); diagnostic != nil {
				return diagnostic
			}
			index++
		}
		if _, err = decoder.Token(); err != nil {
			return syntaxDiagnostic("malformed_json", decoder.InputOffset())
		}
	default:
		return syntaxDiagnostic("malformed_json", decoder.InputOffset())
	}
	return nil
}

func schemaDiagnostic(path, reason string) *Diagnostic {
	code := "FART-E-SCHEMA-0006"
	switch reason {
	case "unsupported_schema":
		code = "FART-E-SCHEMA-0001"
	case "unknown_member":
		code = "FART-E-SCHEMA-0003"
	case "missing_member", "missing_law_context", "missing_capability_request":
		code = "FART-E-SCHEMA-0004"
	case "wrong_type":
		code = "FART-E-SCHEMA-0005"
	case "collection_limit_exceeded", "maximum_nesting_exceeded", "member_name_too_long":
		code = "FART-E-SCHEMA-0007"
	case "scope_reference_unresolved":
		code = "FART-E-SCHEMA-0008"
	case "duplicate_capability_request":
		code = "FART-E-SCHEMA-0009"
	}
	return &Diagnostic{Code: code, Stage: "schema", Path: pathOrRoot(path), ReasonCode: reason}
}

func syntaxDiagnostic(reason string, offset int64) *Diagnostic {
	code := "FART-E-SYNTAX-0001"
	if reason == "trailing_json_value" {
		code = "FART-E-SYNTAX-0002"
	}
	return &Diagnostic{
		Code:       code,
		Stage:      "syntax",
		Path:       "/",
		ReasonCode: reason,
		ByteOffset: offset,
	}
}

func pathOrRoot(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func joinJSONPointer(path, segment string) string {
	escaped := strings.ReplaceAll(strings.ReplaceAll(segment, "~", "~0"), "/", "~1")
	return path + "/" + escaped
}
