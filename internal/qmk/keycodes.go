package qmk

var (
	keycodes = map[string]KC{
		"1": {Default: "1", Shift: "!"},
		"2": {Default: "2", Shift: "@"},
		"3": {Default: "3", Shift: "#"},
		"4": {Default: "4", Shift: "$"},
		"5": {Default: "5", Shift: "%"},
		"6": {Default: "6", Shift: "^"},
		"7": {Default: "7", Shift: "&"},
		"8": {Default: "8", Shift: "*"},
		"9": {Default: "9", Shift: "("},
		"0": {Default: "0", Shift: ")"},

		"MINS":  {Default: "-", Shift: "_"},
		"MINUS": {Default: "-", Shift: "_"},

		"EQUAL": {Default: "=", Shift: "+"},
		"EQL":   {Default: "=", Shift: "+"},

		"LEFT_BRACKET": {Default: "[", Shift: "{"},
		"LBRC":         {Default: "[", Shift: "{"},

		"RIGHT_BRACKET": {Default: "]", Shift: "}"},
		"RBRC":          {Default: "]", Shift: "}"},

		"BACKSLASH": {Default: "\\", Shift: "|"},
		"BSLS":      {Default: "\\", Shift: "|"},

		"NONUS_HASH": {Default: "#", Shift: "~"},
		"NUSH":       {Default: "#", Shift: "~"},

		"SEMICOLON": {Default: ";", Shift: ":"},
		"SCLN":      {Default: ";", Shift: ":"},

		"QUOTE": {Default: "'", Shift: "\""},
		"QUOT":  {Default: "'", Shift: "\""},

		"GRAVE": {Default: "`", Shift: "~"},
		"GRV":   {Default: "`", Shift: "~"},

		"COMMA": {Default: ",", Shift: "<"},
		"COMM":  {Default: ",", Shift: "<"},

		"DOT": {Default: ".", Shift: ">"},

		"SLASH": {Default: "/", Shift: "?"},
		"SLSH":  {Default: "/", Shift: "?"},

		"ENTER": {Default: "enter"},
		"ENT":   {Default: "enter"},

		"ESCAPE": {Default: "esc"},
		"ESC":    {Default: "esc"},

		"BACKSPACE": {Default: "back space"},
		"BSPC":      {Default: "back space"},

		"SPACE": {Default: "space"},
		"SPC":   {Default: "space"},

		"DELETE": {Default: "del"},
		"DEL":    {Default: "del"},

		"TAB": {Default: "tab"},

		"NO": {},

		"TRANSPARENT": {Default: "TRANS"},
		"TRNS":        {Default: "TRANS"},

		"UP":    {Default: "up"},
		"DOWN":  {Default: "down"},
		"LEFT":  {Default: "left"},
		"RIGHT": {Default: "right"},
		"RGHT":  {Default: "right"},

		"HOME": {Default: "home"},
		"END":  {Default: "end"},

		"INSERT": {Default: "ins"},
		"INS":    {Default: "ins"},

		"PAGE_UP": {Default: "page up"},
		"PGUP":    {Default: "page up"},

		"PAGE_DOWN": {Default: "page down"},
		"PGDN":      {Default: "page down"},

		"TILDE": {Default: "~"},
		"TILD":  {Default: "~"},

		"EXCLAIM": {Default: "!"},
		"EXLM":    {Default: "!"},

		"AT": {Default: "@"},

		"HASH": {Default: "#"},

		"DOLLAR": {Default: "$"},
		"DLR":    {Default: "$"},

		"PERCENT": {Default: "%"},
		"PERC":    {Default: "%"},

		"CIRCUMCFLEX": {Default: "^"},
		"CIRC":        {Default: "^"},

		"AMPERSAND": {Default: "&"},
		"AMPR":      {Default: "&"},

		"ASTERISK": {Default: "*"},
		"ASTR":     {Default: "*"},

		"LEFT_PAREN": {Default: "("},
		"LPRN":       {Default: "("},

		"RIGHT_PAREN": {Default: ")"},
		"RPRN":        {Default: ")"},

		"UNDERSCORE": {Default: "_"},
		"UNDS":       {Default: "_"},

		"PLUS": {Default: "+"},

		"LEFT_CURLY_BRACE": {Default: "{"},
		"LCBR":             {Default: "{"},

		"RIGHT_CURLY_BRACE": {Default: "}"},
		"RCBR":              {Default: "}"},

		"PIPE": {Default: "|"},

		"COLON": {Default: ":"},
		"COLN":  {Default: ":"},

		"DOUBLE_QUOTE": {Default: "\""},
		"DQUO":         {Default: "\""},
		"DQT":          {Default: "\""},

		"LEFT_ANGLE_BRACKET": {Default: "<"},
		"LABK":               {Default: "<"},
		"LT":                 {Default: "<"},

		"RIGHT_ANGLE_BRACKET": {Default: ">"},
		"RABK":                {Default: ">"},
		"RT":                  {Default: ">"},

		"QUESTION": {Default: "?"},
		"QUES":     {Default: "?"},

		"KP_COMM": {Default: ","},
		"PCMM":    {Default: ","},

		"KP_SLASH": {Default: "/"},
		"PSLS":     {Default: "/"},

		"KP_ASTERISK": {Default: "*"},
		"PAST":        {Default: "*"},

		"KP_MINUS": {Default: "-"},
		"PMNS":     {Default: "-"},

		"KP_PLUS": {Default: "+"},
		"PPLS":    {Default: "+"},

		"KP_ENTER": {Default: "enter"},
		"PENT":     {Default: "enter"},

		"KP_DOT": {Default: "."},
		"PDOT":   {Default: "."},

		"KP_0": {Default: "0"},
		"P0":   {Default: "0"},

		"KP_1": {Default: "1"},
		"P1":   {Default: "1"},

		"KP_2": {Default: "2"},
		"P2":   {Default: "2"},

		"KP_3": {Default: "3"},
		"P3":   {Default: "3"},

		"KP_4": {Default: "4"},
		"P4":   {Default: "4"},

		"KP_5": {Default: "5"},
		"P5":   {Default: "5"},

		"KP_6": {Default: "6"},
		"P6":   {Default: "6"},

		"KP_7": {Default: "7"},
		"P7":   {Default: "7"},

		"KP_8": {Default: "8"},
		"P8":   {Default: "8"},

		"KP_9": {Default: "9"},
		"P9":   {Default: "9"},

		"LEFT_GUI": {Default: "lgui", Hold: "lgui"},
		"LGUI":     {Default: "lgui", Hold: "lgui"},
		"LCMD":     {Default: "lgui", Hold: "lgui"},
		"LWIN":     {Default: "lgui", Hold: "lgui"},

		"RIGHT_GUI": {Default: "rgui", Hold: "rgui"},
		"RGUI":      {Default: "rgui", Hold: "rgui"},
		"RCMD":      {Default: "rgui", Hold: "rgui"},
		"RWIN":      {Default: "rgui", Hold: "rgui"},

		"LEFT_CTRL": {Hold: "lctrl"},
		"LCTL":      {Hold: "lctrl"},

		"RIGHT_CTRL": {Hold: "rctrl"},
		"RCTL":       {Hold: "rctrl"},

		"LEFT_SHIFT": {Hold: "lsft"},
		"LSFT":       {Hold: "lsft"},

		"RIGHT_SHIFT": {Hold: "rsft"},
		"RSFT":        {Hold: "rsft"},

		"LEFT_ALT": {Hold: "lalt"},
		"LALT":     {Hold: "lalt"},
		"LOPT":     {Hold: "lalt"},

		"RIGHT_ALT": {Hold: "ralt"},
		"RALT":      {Hold: "ralt"},
		"ROPT":      {Hold: "ralt"},
		"RAGR":      {Hold: "ralt"},

		"F1":  {Default: "F1"},
		"F2":  {Default: "F2"},
		"F3":  {Default: "F3"},
		"F4":  {Default: "F4"},
		"F5":  {Default: "F5"},
		"F6":  {Default: "F6"},
		"F7":  {Default: "F7"},
		"F8":  {Default: "F8"},
		"F9":  {Default: "F9"},
		"F10": {Default: "F10"},
		"F11": {Default: "F11"},
		"F12": {Default: "F12"},
	}
)
