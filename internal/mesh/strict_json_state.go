package mesh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// rejectDuplicateJSONKeys rejects ambiguous persisted state before semantic
// decoding. encoding/json otherwise accepts duplicate object members and keeps a
// later value, which is inappropriate for durable authority/checkpoint state.
func rejectDuplicateJSONKeys(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value is not allowed")
		}
		return err
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	switch delim {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("JSON object key is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate JSON object key %q", key)
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return fmt.Errorf("JSON object was not closed")
		}
		return nil
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return fmt.Errorf("JSON array was not closed")
		}
		return nil
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
}

func (state *durableEventJournalState) UnmarshalJSON(body []byte) error {
	if err := rejectDuplicateJSONKeys(body); err != nil {
		return fmt.Errorf("durable event journal JSON is ambiguous: %w", err)
	}
	type stateAlias durableEventJournalState
	var decoded stateAlias
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	*state = durableEventJournalState(decoded)
	return nil
}

func (state *subscriberCheckpointState) UnmarshalJSON(body []byte) error {
	if err := rejectDuplicateJSONKeys(body); err != nil {
		return fmt.Errorf("subscriber checkpoint JSON is ambiguous: %w", err)
	}
	type stateAlias subscriberCheckpointState
	var decoded stateAlias
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	*state = subscriberCheckpointState(decoded)
	return nil
}
