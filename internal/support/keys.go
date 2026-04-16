package support

// KeysValues 表示 W3C WebDriver 特殊键值常量集。
type KeysValues struct {
	NULL       string
	CANCEL     string
	HELP       string
	BACKSPACE  string
	BACK_SPACE string
	TAB        string
	CLEAR      string
	RETURN     string
	ENTER      string

	SHIFT   string
	CONTROL string
	CTRL    string
	ALT     string
	META    string
	COMMAND string

	PAUSE  string
	ESCAPE string
	ESC    string
	SPACE  string

	PAGE_UP   string
	PAGE_DOWN string
	END       string
	HOME      string

	LEFT        string
	ARROW_LEFT  string
	UP          string
	ARROW_UP    string
	RIGHT       string
	ARROW_RIGHT string
	DOWN        string
	ARROW_DOWN  string

	INSERT string
	DELETE string

	SEMICOLON string
	EQUALS    string

	NUMPAD0   string
	NUMPAD1   string
	NUMPAD2   string
	NUMPAD3   string
	NUMPAD4   string
	NUMPAD5   string
	NUMPAD6   string
	NUMPAD7   string
	NUMPAD8   string
	NUMPAD9   string
	MULTIPLY  string
	ADD       string
	SEPARATOR string
	SUBTRACT  string
	DECIMAL   string
	DIVIDE    string

	F1  string
	F2  string
	F3  string
	F4  string
	F5  string
	F6  string
	F7  string
	F8  string
	F9  string
	F10 string
	F11 string
	F12 string
}

// Keys 提供与 Python 版一致的点号访问风格。
var Keys = KeysValues{
	NULL:       "\ue000",
	CANCEL:     "\ue001",
	HELP:       "\ue002",
	BACKSPACE:  "\ue003",
	BACK_SPACE: "\ue003",
	TAB:        "\ue004",
	CLEAR:      "\ue005",
	RETURN:     "\ue006",
	ENTER:      "\ue007",

	SHIFT:   "\ue008",
	CONTROL: "\ue009",
	CTRL:    "\ue009",
	ALT:     "\ue00a",
	META:    "\ue03d",
	COMMAND: "\ue03d",

	PAUSE:  "\ue00b",
	ESCAPE: "\ue00c",
	ESC:    "\ue00c",
	SPACE:  "\ue00d",

	PAGE_UP:   "\ue00e",
	PAGE_DOWN: "\ue00f",
	END:       "\ue010",
	HOME:      "\ue011",

	LEFT:        "\ue012",
	ARROW_LEFT:  "\ue012",
	UP:          "\ue013",
	ARROW_UP:    "\ue013",
	RIGHT:       "\ue014",
	ARROW_RIGHT: "\ue014",
	DOWN:        "\ue015",
	ARROW_DOWN:  "\ue015",

	INSERT: "\ue016",
	DELETE: "\ue017",

	SEMICOLON: "\ue018",
	EQUALS:    "\ue019",

	NUMPAD0:   "\ue01a",
	NUMPAD1:   "\ue01b",
	NUMPAD2:   "\ue01c",
	NUMPAD3:   "\ue01d",
	NUMPAD4:   "\ue01e",
	NUMPAD5:   "\ue01f",
	NUMPAD6:   "\ue020",
	NUMPAD7:   "\ue021",
	NUMPAD8:   "\ue022",
	NUMPAD9:   "\ue023",
	MULTIPLY:  "\ue024",
	ADD:       "\ue025",
	SEPARATOR: "\ue026",
	SUBTRACT:  "\ue027",
	DECIMAL:   "\ue028",
	DIVIDE:    "\ue029",

	F1:  "\ue031",
	F2:  "\ue032",
	F3:  "\ue033",
	F4:  "\ue034",
	F5:  "\ue035",
	F6:  "\ue036",
	F7:  "\ue037",
	F8:  "\ue038",
	F9:  "\ue039",
	F10: "\ue03a",
	F11: "\ue03b",
	F12: "\ue03c",
}
