package mcprobot

const JsonRPC = "2.0"

// const ProtocolVersion = "2025-06-18"
const ProtocolVersion = "2025-03-26"

const MethodInitialize = "initialize"
const MethodNotificationsInitialized = "notifications/initialized"

const MethodRootsList = "roots/list"
const MethodNotificationsRootsListChanged = "notifications/roots/list_changed"

const MethodToolsList = "tools/list"
const MethodToolsCall = "tools/call"
const MethodNotificationsToolsListChanged = "notifications/tools/list_changed"

type (
	ID interface {
		string | int
	}

	Request[UID ID] struct {
		Jsonrpc string         `json:"jsonrpc"`
		ID      UID            `json:"id"`
		Method  string         `json:"method"`
		Params  map[string]any `json:"params,omitempty"`
	}

	CapabilityParam struct {
		Subscribe   bool `json:"subscribe"`
		ListChanged bool `json:"listChanged"`
	}

	Info struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	Response[UID ID] struct {
		Jsonrpc      string         `json:"jsonrpc"`
		ID           UID            `json:"id"`
		Result       map[string]any `json:"result"`
		ServerInfo   *Info          `json:"serverInfo,omitempty"`
		Instructions string         `json:"instructions,omitempty"`
		Error        *Error         `json:"error,omitempty"`
	}

	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	Notification struct {
		JsonRPC string         `json:"jsonrpc"`
		Method  string         `json:"method"`
		Params  map[string]any `json:"params,omitempty"`
	}
)

// Params  struct {
// 			ProtocolVersion string `json:"protocolVersion"`
// 			Capabilities    struct {
// 				Roots Capability `json:"roots"`
// 			} `json:"capabilities"`
// 		}

// Result  struct {
// 	ProtocolVersion string `json:"protocolVersion"`
// 	Capabilities    struct {
// 		Logging   Capability `json:"logging"`
// 		Resources Capability `json:"resources"`
// 		Tools     Capability `json:"tools"`
// 	}
// }

type (
	RootCapability struct {
		URI  string `json:"uri"` // MUST be a file://
		Name string `json:"name,omitempty"`
	}
)
