package model

// Manifest is the root metadata for a normalized SAZ bundle.
type Manifest struct {
	Format                string           `json:"format"`
	FormatVersion         string           `json:"formatVersion"`
	SourceSaz             string           `json:"sourceSaz"`
	SourceSazBaseName     string           `json:"sourceSazBaseName"`
	FiddlerSessionOrder   []int            `json:"fiddlerSessionOrder"`
	SessionCount          int              `json:"sessionCount"`
	Sessions              []SessionSummary `json:"sessions"`
	NormalizationPolicy   Policy           `json:"normalizationPolicy"`
	Notes                 []string         `json:"notes"`
}

// Policy documents what transforms are allowed in normalized output.
type Policy struct {
	PreserveRawFiles         bool     `json:"preserveRawFiles"`
	ReplaceOriginalFiles     bool     `json:"replaceOriginalFiles"`
	NormalizeTransferCoding  bool     `json:"normalizeTransferCoding"`
	NormalizeContentEncoding bool     `json:"normalizeContentEncoding"`
	DecodeCharset            bool     `json:"decodeCharset"`
	PrettyPrintJSON          bool     `json:"prettyPrintJson"`
	SearchLayerIncluded      bool     `json:"searchLayerIncluded"`
	Notes                    []string `json:"notes"`
}

// SessionSummary gives bundle-level indexing for one Fiddler session.
type SessionSummary struct {
	SessionID          int      `json:"sessionId"`
	Ordinal            int      `json:"ordinal"`
	RequestPath        string   `json:"requestPath"`
	ResponsePath       string   `json:"responsePath"`
	MetaPath           string   `json:"metaPath"`
	DecodedBodyPath    string   `json:"decodedBodyPath,omitempty"`
	SearchBodyPath     string   `json:"searchBodyPath,omitempty"`
	URL                string   `json:"url,omitempty"`
	Method             string   `json:"method,omitempty"`
	StatusCode         int      `json:"statusCode,omitempty"`
	ResponseBodyIsText bool     `json:"responseBodyIsText"`
	StatusMarkers      []string `json:"statusMarkers,omitempty"`
}

// SessionMeta records provenance and transforms for a single session.
type SessionMeta struct {
	SessionID              int      `json:"sessionId"`
	Ordinal                int      `json:"ordinal"`
	SourceSaz              string   `json:"sourceSaz"`
	SourceRequestPath      string   `json:"sourceRequestPath"`
	SourceResponsePath     string   `json:"sourceResponsePath"`
	RequestPath            string   `json:"requestPath"`
	ResponsePath           string   `json:"responsePath"`
	DecodedBodyPath        string   `json:"decodedBodyPath,omitempty"`
	SearchBodyPath         string   `json:"searchBodyPath,omitempty"`
	Method                 string   `json:"method,omitempty"`
	URL                    string   `json:"url,omitempty"`
	StatusCode             int      `json:"statusCode,omitempty"`
	ContentType            string   `json:"contentType,omitempty"`
	TransferEncoding       string   `json:"transferEncoding,omitempty"`
	ContentEncoding        string   `json:"contentEncoding,omitempty"`
	Charset                string   `json:"charset,omitempty"`
	Transforms             []string `json:"transforms"`
	DecodedBodyText        bool     `json:"decodedBodyText"`
	BodyExactAfterDecode   bool     `json:"bodyExactAfterDecode"`
	JSONPrettyPrinted      bool     `json:"jsonPrettyPrinted"`
	StatusMarkers          []string `json:"statusMarkers,omitempty"`
	BodyMissing            bool     `json:"bodyMissing"`
	BodyTruncated          bool     `json:"bodyTruncated"`
	DecodeFailed           bool     `json:"decodeFailed"`
	BinaryBodySkipped      bool     `json:"binaryBodySkipped"`
	DecodeFailureReason    string   `json:"decodeFailureReason,omitempty"`
	NormalizationNotes     []string `json:"normalizationNotes"`
}
