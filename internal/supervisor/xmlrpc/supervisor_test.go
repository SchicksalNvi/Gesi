package xmlrpc

import (
	"strings"
	"testing"
)

// TestParseBooleanResponse_Success tests parsing successful boolean responses
func TestParseBooleanResponse_Success(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		expected bool
		wantErr  bool
	}{
		{
			name: "boolean true (1)",
			xml: `<?xml version="1.0"?>
<methodResponse>
  <params>
    <param>
      <value><boolean>1</boolean></value>
    </param>
  </params>
</methodResponse>`,
			expected: true,
			wantErr:  false,
		},
		{
			name: "boolean false (0)",
			xml: `<?xml version="1.0"?>
<methodResponse>
  <params>
    <param>
      <value><boolean>0</boolean></value>
    </param>
  </params>
</methodResponse>`,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "simple boolean true",
			xml:      `<value><boolean>1</boolean></value>`,
			expected: true,
			wantErr:  false,
		},
		{
			name:     "simple boolean false",
			xml:      `<value><boolean>0</boolean></value>`,
			expected: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBooleanResponse(tt.xml)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBooleanResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("parseBooleanResponse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestParseBooleanResponse_Errors tests error handling for invalid responses
func TestParseBooleanResponse_Errors(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		wantErr bool
	}{
		{
			name:    "no boolean value",
			xml:     `<value><string>hello</string></value>`,
			wantErr: true,
		},
		{
			name:    "empty response",
			xml:     ``,
			wantErr: true,
		},
		{
			name:    "malformed boolean",
			xml:     `<value><boolean>invalid</boolean></value>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseBooleanResponse(tt.xml)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBooleanResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestParseFaultResponse tests parsing fault responses
func TestParseFaultResponse(t *testing.T) {
	tests := []struct {
		name            string
		xml             string
		expectedCode    int
		expectedString  string
		expectedIsFault bool
	}{
		{
			name: "ALREADY_STARTED fault",
			xml: `<?xml version="1.0"?>
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
</methodResponse>`,
			expectedCode:    60,
			expectedString:  "ALREADY_STARTED: process test-process",
			expectedIsFault: true,
		},
		{
			name: "NOT_RUNNING fault",
			xml: `<?xml version="1.0"?>
<methodResponse>
  <fault>
    <value>
      <struct>
        <member>
          <name>faultCode</name>
          <value><int>70</int></value>
        </member>
        <member>
          <name>faultString</name>
          <value><string>NOT_RUNNING: process test-process</string></value>
        </member>
      </struct>
    </value>
  </fault>
</methodResponse>`,
			expectedCode:    70,
			expectedString:  "NOT_RUNNING: process test-process",
			expectedIsFault: true,
		},
		{
			name: "BAD_NAME fault",
			xml: `<?xml version="1.0"?>
<methodResponse>
  <fault>
    <value>
      <struct>
        <member>
          <name>faultCode</name>
          <value><int>10</int></value>
        </member>
        <member>
          <name>faultString</name>
          <value><string>BAD_NAME: unknown-process</string></value>
        </member>
      </struct>
    </value>
  </fault>
</methodResponse>`,
			expectedCode:    10,
			expectedString:  "BAD_NAME: unknown-process",
			expectedIsFault: true,
		},
		{
			name:            "not a fault response",
			xml:             `<value><boolean>1</boolean></value>`,
			expectedCode:    0,
			expectedString:  "",
			expectedIsFault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, faultString, isFault := parseFaultResponse(tt.xml)
			if isFault != tt.expectedIsFault {
				t.Errorf("parseFaultResponse() isFault = %v, want %v", isFault, tt.expectedIsFault)
			}
			if code != tt.expectedCode {
				t.Errorf("parseFaultResponse() code = %v, want %v", code, tt.expectedCode)
			}
			if faultString != tt.expectedString {
				t.Errorf("parseFaultResponse() faultString = %v, want %v", faultString, tt.expectedString)
			}
		})
	}
}

// TestParseBooleanResponse_WithFault tests that fault responses are handled correctly
func TestParseBooleanResponse_WithFault(t *testing.T) {
	faultXML := `<?xml version="1.0"?>
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
          <value><string>ALREADY_STARTED: process test</string></value>
        </member>
      </struct>
    </value>
  </fault>
</methodResponse>`

	_, err := parseBooleanResponse(faultXML)
	if err == nil {
		t.Error("parseBooleanResponse() should return error for fault response")
	}
	if !strings.Contains(err.Error(), "ALREADY_STARTED") {
		t.Errorf("error should contain fault string, got: %v", err)
	}
}
