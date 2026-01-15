package xmlrpc

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
)

// Feature: process-operation-error-feedback, Property 1: Boolean Response Parsing Round-Trip
// For any valid XML-RPC boolean response, the parseBooleanResponse function SHALL correctly
// extract the boolean value and return it without error.
func TestProperty_BooleanResponseParsing(t *testing.T) {
	// Property: For any boolean value, constructing XML and parsing should return the same value
	property := func(boolValue bool) bool {
		// Construct XML response
		boolStr := "0"
		if boolValue {
			boolStr = "1"
		}
		xml := fmt.Sprintf(`<?xml version="1.0"?>
<methodResponse>
  <params>
    <param>
      <value><boolean>%s</boolean></value>
    </param>
  </params>
</methodResponse>`, boolStr)

		// Parse and verify
		result, err := parseBooleanResponse(xml)
		if err != nil {
			return false
		}
		return result == boolValue
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 1 failed: %v", err)
	}
}

// Feature: process-operation-error-feedback, Property 2: Fault Response Parsing Completeness
// For any valid XML-RPC fault response containing faultCode and faultString, the parseFaultResponse
// function SHALL extract both values correctly.
func TestProperty_FaultResponseParsing(t *testing.T) {
	// Property: For any faultCode and faultString, constructing XML and parsing should return the same values
	property := func(faultCode int, faultString string) bool {
		// Sanitize faultString to avoid XML issues
		faultString = strings.ReplaceAll(faultString, "<", "&lt;")
		faultString = strings.ReplaceAll(faultString, ">", "&gt;")
		faultString = strings.ReplaceAll(faultString, "&", "&amp;")

		// Limit faultCode to reasonable range
		if faultCode < 0 {
			faultCode = -faultCode
		}
		faultCode = faultCode % 1000

		xml := fmt.Sprintf(`<?xml version="1.0"?>
<methodResponse>
  <fault>
    <value>
      <struct>
        <member>
          <name>faultCode</name>
          <value><int>%d</int></value>
        </member>
        <member>
          <name>faultString</name>
          <value><string>%s</string></value>
        </member>
      </struct>
    </value>
  </fault>
</methodResponse>`, faultCode, faultString)

		// Parse and verify
		code, str, isFault := parseFaultResponse(xml)
		if !isFault {
			return false
		}
		if code != faultCode {
			return false
		}
		// faultString should match (after XML escaping was applied)
		return str == faultString
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 2 failed: %v", err)
	}
}

// Feature: process-operation-error-feedback, Property 3: Malformed XML Graceful Handling
// For any malformed or incomplete XML string, the parsing functions SHALL return a descriptive
// error instead of panicking or returning incorrect results.
func TestProperty_MalformedXMLHandling(t *testing.T) {
	// Property: For any random string, parseBooleanResponse should not panic
	property := func(randomData []byte) bool {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseBooleanResponse panicked on input: %s", string(randomData))
			}
		}()

		// Call the function - it should not panic
		_, _ = parseBooleanResponse(string(randomData))
		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 failed: %v", err)
	}
}

// Feature: process-operation-error-feedback, Property 3 (continued): Fault parsing graceful handling
func TestProperty_MalformedXMLFaultHandling(t *testing.T) {
	// Property: For any random string, parseFaultResponse should not panic
	property := func(randomData []byte) bool {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("parseFaultResponse panicked on input: %s", string(randomData))
			}
		}()

		// Call the function - it should not panic
		_, _, _ = parseFaultResponse(string(randomData))
		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (fault) failed: %v", err)
	}
}

// Feature: process-operation-error-feedback, Property 4: StartProcess XML Response Handling
// For any XML response from Supervisor (success, failure, or fault), the StartProcess method
// SHALL correctly interpret the response.
// Note: This is tested via mock responses since we can't call actual Supervisor

// TestProperty_BooleanResponseConsistency tests that parsing is consistent
func TestProperty_BooleanResponseConsistency(t *testing.T) {
	// Property: Parsing the same XML multiple times should always return the same result
	property := func(seed int64) bool {
		rng := rand.New(rand.NewSource(seed))
		boolValue := rng.Intn(2) == 1

		boolStr := "0"
		if boolValue {
			boolStr = "1"
		}
		xml := fmt.Sprintf(`<value><boolean>%s</boolean></value>`, boolStr)

		// Parse multiple times
		results := make([]bool, 10)
		for i := 0; i < 10; i++ {
			result, err := parseBooleanResponse(xml)
			if err != nil {
				return false
			}
			results[i] = result
		}

		// All results should be the same
		for i := 1; i < len(results); i++ {
			if results[i] != results[0] {
				return false
			}
		}
		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property consistency failed: %v", err)
	}
}

// TestProperty_FaultDetection tests that fault responses are always detected
func TestProperty_FaultDetection(t *testing.T) {
	// Property: Any XML containing <fault> should be detected as a fault
	property := func(faultCode int, faultString string) bool {
		// Sanitize
		faultString = strings.ReplaceAll(faultString, "<", "")
		faultString = strings.ReplaceAll(faultString, ">", "")
		if faultCode < 0 {
			faultCode = -faultCode
		}

		xml := fmt.Sprintf(`<fault><value><struct>
			<member><name>faultCode</name><value><int>%d</int></value></member>
			<member><name>faultString</name><value><string>%s</string></value></member>
		</struct></value></fault>`, faultCode, faultString)

		_, _, isFault := parseFaultResponse(xml)
		return isFault
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property fault detection failed: %v", err)
	}
}


// Feature: process-operation-error-feedback, Property 4: StartProcess XML Response Handling
// Tests that StartProcess correctly handles various XML response types

// TestProperty_StartProcessSuccessResponse tests that success responses are handled correctly
func TestProperty_StartProcessSuccessResponse(t *testing.T) {
	// This test verifies the parsing logic used by StartProcess
	// We test the underlying parseBooleanResponse since StartProcess uses it

	// Property: For any successful boolean response, parsing should succeed
	successXML := `<?xml version="1.0"?>
<methodResponse>
  <params>
    <param>
      <value><boolean>1</boolean></value>
    </param>
  </params>
</methodResponse>`

	result, err := parseBooleanResponse(successXML)
	if err != nil {
		t.Errorf("Property 4: Success response should not return error: %v", err)
	}
	if !result {
		t.Error("Property 4: Success response should return true")
	}
}

// TestProperty_StartProcessFaultHandling tests that fault responses are handled correctly
func TestProperty_StartProcessFaultHandling(t *testing.T) {
	// Property: ALREADY_STARTED fault should be detected and can be treated as success
	alreadyStartedXML := `<?xml version="1.0"?>
<methodResponse>
  <fault>
    <value>
      <struct>
        <member>
          <name>faultCode</name>
          <value><int>60</int></value>
        </member>
        <member>
          <name>faultString</name>
          <value><string>ALREADY_STARTED: process test-process</string></value>
        </member>
      </struct>
    </value>
  </fault>
</methodResponse>`

	code, faultString, isFault := parseFaultResponse(alreadyStartedXML)
	if !isFault {
		t.Error("Property 4: ALREADY_STARTED should be detected as fault")
	}
	if code != 60 {
		t.Errorf("Property 4: Expected fault code 60, got %d", code)
	}
	if !strings.Contains(faultString, "ALREADY_STARTED") {
		t.Errorf("Property 4: Fault string should contain ALREADY_STARTED, got: %s", faultString)
	}
}

// TestProperty_StartProcessIdempotency tests idempotent behavior
func TestProperty_StartProcessIdempotency(t *testing.T) {
	// Property: For any process name, ALREADY_STARTED fault should be treated as success
	property := func(processName string) bool {
		// Sanitize process name
		processName = strings.ReplaceAll(processName, "<", "")
		processName = strings.ReplaceAll(processName, ">", "")
		if len(processName) > 50 {
			processName = processName[:50]
		}

		faultXML := fmt.Sprintf(`<fault><value><struct>
			<member><name>faultCode</name><value><int>60</int></value></member>
			<member><name>faultString</name><value><string>ALREADY_STARTED: %s</string></value></member>
		</struct></value></fault>`, processName)

		_, faultString, isFault := parseFaultResponse(faultXML)
		if !isFault {
			return false
		}
		// ALREADY_STARTED should be detectable for idempotent handling
		return strings.Contains(faultString, "ALREADY_STARTED")
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 4 idempotency failed: %v", err)
	}
}
