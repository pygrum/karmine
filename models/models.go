package models

// Storage container for all KarObject types.
type GenericData struct {
	CmdID  int    `json:"id"` // The ID of the command that requested this data
	UUID   string `json:"uuid"`
	Type   int    `json:"object_type"`
	Object []byte `json:"object"`
}

// This object is received from Karmine C2, parsed and the results sent back as KarResponseObjectCmd
type KarObjectCmd struct {
	Cmd  int         `json:"cmd"`
	Args []MultiType `json:"args"`
}

// This object is only sent after a command has been received and parsed.
// lives only under 'Object' in GenericData object
type KarResponseObjectCmd struct {
	Code int `json:"statuscode"`
	Data struct {
		Result string `json:"result"` // The result of the executed command
		Error  string `json:"error"`
	} `json:"data"`
}

type KarObjectFile struct {
	FileBytes []byte `json:"bytes"`
	FileName  string `json:"filename"`
}

type KarResponseObjectFile struct {
	Error  int    `json:"errors"`
	ErrVal string `json:"errval"`
}

// This object is only sent during credential acquirement
// lives only under 'data' in GenericData object

type KarObjectCred struct {
	Platform string  `json:"platform"`
	Creds    CredObj `json:"creds"`
}

type CredObj struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Custom command argument. Lives under KarObjectCmd, received from Karmine C2

// a struct capable of holding multiple types of data. Used when the response type for requested data is unknown.
type MultiType struct {
	IntValue int    `json:"int"`
	StrValue string `json:"string"`
}

type Error struct {
	Name    string `json:"error"`
	Details string `json:"details"`
}

type TmpConf struct {
	LHost    string `json:"lhost"`
	LPort    string `json:"lport"`
	Endpoint string `json:"endpoint"`
}
