package config

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	_ "time/tzdata"
)

const (
	autoFPIPInfoURL            = "https://ip234.in/ip.json"
	autoFPIPAPIURL             = "http://ip-api.com/json/?fields"
	autoFPIPAPICOURL           = "https://ipapi.co/json/"
	autoFPIPWhoIsURL           = "https://ipwho.is/"
	autoFPIPInfoIOURL          = "https://ipinfo.io/json"
	autoFPFreeIPAPIURL         = "https://freeipapi.com/api/json"
	defaultWindowsFontSystem   = "windows"
	defaultWindowsUserAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:147.0) Gecko/20100101 Firefox/151.0"
	defaultWebGLVersion        = "WebGL 1.0 (OpenGL ES 2.0 Chromium)"
	defaultWebGLGLSLVersion    = "WebGL GLSL ES 1.0 (OpenGL ES GLSL ES 1.0 Chromium)"
	defaultWebGLMaxTexture     = 16384
	defaultWebGLMaxCubeMap     = 16384
	defaultWebGLImageUnits     = 32
	defaultWebGLVertexAttr     = 16
	defaultWebGLPointSize      = 1024
	defaultWebGLViewportDim    = 16384
	defaultCanvasSeedMax       = 1_000_000_000
	autoFPIPInfoRequestTimeout = 4 * time.Second
)

var (
	fetchAutoFPIPInfo = fetchAutoFPFingerprintProfile
	autoFPIPProviders = []autoFPIPProvider{
		{name: "ip-api", fetch: fetchIPAPIFingerprintProfile},
		{name: "ipapi", fetch: fetchIPAPICOFingerprintProfile},
		{name: "ipwhois", fetch: fetchIPWhoIsFingerprintProfile},
		{name: "ipinfo", fetch: fetchIPInfoFingerprintProfile},
		{name: "freeipapi", fetch: fetchFreeIPAPIFingerprintProfile},
	}
)

type autoFPIPProvider struct {
	name  string
	fetch func(context.Context, *http.Client) (autoFPIPInfoResponse, error)
}

type autoFPIPProviderResult struct {
	index    int
	name     string
	response autoFPIPInfoResponse
	err      error
}

type autoFPIPInfoResponse struct {
	IP          string `json:"ip"`
	City        string `json:"city"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Timezone    string `json:"timezone"`
	Region      string `json:"region"`
}

type autoFPVoiceProfile struct {
	Language    string
	VoiceNames  []string
	VoiceLangs  []string
	DefaultName string
	DefaultLang string
}

type autoFPWebGLProfile struct {
	Vendor   string
	Renderer string
}

type autoFPProxyAuth struct {
	Username       string
	Password       string
	HasCredentials bool
}

var autoFPVoiceProfiles = map[string]autoFPVoiceProfile{
	"AU": {
		Language:    "en-AU,en",
		VoiceNames:  []string{"Microsoft Catherine - English (Australia)"},
		VoiceLangs:  []string{"en-AU"},
		DefaultName: "Microsoft Catherine - English (Australia)",
		DefaultLang: "en-AU",
	},
	"BR": {
		Language:    "pt-BR,pt",
		VoiceNames:  []string{"Microsoft Maria Desktop - Portuguese(Brazil)"},
		VoiceLangs:  []string{"pt-BR"},
		DefaultName: "Microsoft Maria Desktop - Portuguese(Brazil)",
		DefaultLang: "pt-BR",
	},
	"CA": {
		Language:    "en-CA,en",
		VoiceNames:  []string{"Microsoft Linda - English (Canada)"},
		VoiceLangs:  []string{"en-CA"},
		DefaultName: "Microsoft Linda - English (Canada)",
		DefaultLang: "en-CA",
	},
	"CN": {
		Language:    "zh-CN,zh",
		VoiceNames:  []string{"Microsoft Huihui Desktop - Chinese (Simplified)"},
		VoiceLangs:  []string{"zh-CN"},
		DefaultName: "Microsoft Huihui Desktop - Chinese (Simplified)",
		DefaultLang: "zh-CN",
	},
	"DE": {
		Language:    "de-DE,de",
		VoiceNames:  []string{"Microsoft Hedda Desktop - German"},
		VoiceLangs:  []string{"de-DE"},
		DefaultName: "Microsoft Hedda Desktop - German",
		DefaultLang: "de-DE",
	},
	"ES": {
		Language:    "es-ES,es",
		VoiceNames:  []string{"Microsoft Helena Desktop - Spanish"},
		VoiceLangs:  []string{"es-ES"},
		DefaultName: "Microsoft Helena Desktop - Spanish",
		DefaultLang: "es-ES",
	},
	"FR": {
		Language:    "fr-FR,fr",
		VoiceNames:  []string{"Microsoft Hortense Desktop - French"},
		VoiceLangs:  []string{"fr-FR"},
		DefaultName: "Microsoft Hortense Desktop - French",
		DefaultLang: "fr-FR",
	},
	"GB": {
		Language:    "en-GB,en",
		VoiceNames:  []string{"Microsoft Hazel Desktop - English (Great Britain)"},
		VoiceLangs:  []string{"en-GB"},
		DefaultName: "Microsoft Hazel Desktop - English (Great Britain)",
		DefaultLang: "en-GB",
	},
	"HK": {
		Language:    "zh-HK,zh,en-US,en",
		VoiceNames:  []string{"Microsoft Tracy - Chinese (Traditional, Hong Kong SAR)"},
		VoiceLangs:  []string{"zh-HK"},
		DefaultName: "Microsoft Tracy - Chinese (Traditional, Hong Kong SAR)",
		DefaultLang: "zh-HK",
	},
	"IE": {
		Language:    "en-IE,en",
		VoiceNames:  []string{"Microsoft Sean - English (Ireland)"},
		VoiceLangs:  []string{"en-IE"},
		DefaultName: "Microsoft Sean - English (Ireland)",
		DefaultLang: "en-IE",
	},
	"IN": {
		Language:    "en-IN,en",
		VoiceNames:  []string{"Microsoft Heera - English (India)"},
		VoiceLangs:  []string{"en-IN"},
		DefaultName: "Microsoft Heera - English (India)",
		DefaultLang: "en-IN",
	},
	"IT": {
		Language:    "it-IT,it",
		VoiceNames:  []string{"Microsoft Elsa Desktop - Italian"},
		VoiceLangs:  []string{"it-IT"},
		DefaultName: "Microsoft Elsa Desktop - Italian",
		DefaultLang: "it-IT",
	},
	"JP": {
		Language:    "ja-JP,ja",
		VoiceNames:  []string{"Microsoft Haruka Desktop - Japanese", "Microsoft Ichiro Desktop - Japanese", "Microsoft Ayumi - Japanese (Japan)"},
		VoiceLangs:  []string{"ja-JP", "ja-JP", "ja-JP"},
		DefaultName: "Microsoft Haruka Desktop - Japanese",
		DefaultLang: "ja-JP",
	},
	"KR": {
		Language:    "ko-KR,ko",
		VoiceNames:  []string{"Microsoft Heami - Korean (Korea)"},
		VoiceLangs:  []string{"ko-KR"},
		DefaultName: "Microsoft Heami - Korean (Korea)",
		DefaultLang: "ko-KR",
	},
	"MX": {
		Language:    "es-MX,es",
		VoiceNames:  []string{"Microsoft Sabina - Spanish (Mexico)"},
		VoiceLangs:  []string{"es-MX"},
		DefaultName: "Microsoft Sabina - Spanish (Mexico)",
		DefaultLang: "es-MX",
	},
	"NL": {
		Language:    "nl-NL,nl",
		VoiceNames:  []string{"Microsoft Frank - Dutch (Netherlands)"},
		VoiceLangs:  []string{"nl-NL"},
		DefaultName: "Microsoft Frank - Dutch (Netherlands)",
		DefaultLang: "nl-NL",
	},
	"NZ": {
		Language:    "en-NZ,en",
		VoiceNames:  []string{"Microsoft Molly - English (New Zealand)"},
		VoiceLangs:  []string{"en-NZ"},
		DefaultName: "Microsoft Molly - English (New Zealand)",
		DefaultLang: "en-NZ",
	},
	"PT": {
		Language:    "pt-PT,pt",
		VoiceNames:  []string{"Microsoft Helia - Portuguese (Portugal)"},
		VoiceLangs:  []string{"pt-PT"},
		DefaultName: "Microsoft Helia - Portuguese (Portugal)",
		DefaultLang: "pt-PT",
	},
	"RU": {
		Language:    "ru-RU,ru",
		VoiceNames:  []string{"Microsoft Irina Desktop - Russian"},
		VoiceLangs:  []string{"ru-RU"},
		DefaultName: "Microsoft Irina Desktop - Russian",
		DefaultLang: "ru-RU",
	},
	"SG": {
		Language:    "en-SG,en,zh-CN,zh",
		VoiceNames:  []string{"Microsoft Zira Desktop - English (United States)"},
		VoiceLangs:  []string{"en-SG"},
		DefaultName: "Microsoft Zira Desktop - English (United States)",
		DefaultLang: "en-SG",
	},
	"TR": {
		Language:    "tr-TR,tr",
		VoiceNames:  []string{"Microsoft Tolga Desktop - Turkish"},
		VoiceLangs:  []string{"tr-TR"},
		DefaultName: "Microsoft Tolga Desktop - Turkish",
		DefaultLang: "tr-TR",
	},
	"TW": {
		Language:    "zh-TW,zh",
		VoiceNames:  []string{"Microsoft Hanhan Desktop - Chinese (Traditional)"},
		VoiceLangs:  []string{"zh-TW"},
		DefaultName: "Microsoft Hanhan Desktop - Chinese (Traditional)",
		DefaultLang: "zh-TW",
	},
	"US": {
		Language:    "en-US,en",
		VoiceNames:  []string{"Microsoft David Desktop - English (United States)", "Microsoft Zira Desktop - English (United States)"},
		VoiceLangs:  []string{"en-US", "en-US"},
		DefaultName: "Microsoft David Desktop - English (United States)",
		DefaultLang: "en-US",
	},
}

var autoFPVoiceProfileAliases = map[string]string{
	"AD": "ES", "AE": "IN", "AF": "IN", "AG": "US", "AI": "US", "AL": "IT", "AM": "RU",
	"AO": "PT", "AR": "ES", "AS": "US", "AT": "DE", "AW": "NL", "AX": "GB", "AZ": "TR",
	"BA": "DE", "BB": "US", "BD": "IN", "BE": "FR", "BF": "FR", "BG": "DE", "BH": "IN",
	"BI": "FR", "BJ": "FR", "BL": "FR", "BM": "GB", "BN": "SG", "BO": "ES", "BQ": "NL",
	"BS": "US", "BT": "IN", "BV": "NO", "BW": "GB", "BY": "RU", "BZ": "US",
	"CC": "AU", "CD": "FR", "CF": "FR", "CG": "FR", "CH": "DE", "CI": "FR", "CK": "NZ",
	"CL": "ES", "CM": "FR", "CO": "ES", "CR": "ES", "CU": "ES", "CV": "PT", "CW": "NL",
	"CX": "AU", "CY": "GB", "CZ": "DE",
	"DJ": "FR", "DK": "DE", "DM": "US", "DO": "ES", "DZ": "FR",
	"EC": "ES", "EE": "DE", "EG": "TR", "EH": "ES", "ER": "GB", "ET": "GB",
	"FI": "DE", "FJ": "AU", "FK": "GB", "FM": "US", "FO": "GB",
	"GA": "FR", "GD": "US", "GE": "TR", "GF": "FR", "GG": "GB", "GH": "GB", "GI": "GB",
	"GL": "DK", "GM": "GB", "GN": "FR", "GP": "FR", "GQ": "ES", "GR": "IT", "GS": "GB",
	"GT": "ES", "GU": "US", "GW": "PT", "GY": "GB",
	"HM": "AU", "HN": "ES", "HR": "DE", "HT": "FR", "HU": "DE",
	"ID": "SG", "IL": "GB", "IM": "GB", "IO": "GB", "IQ": "TR", "IR": "TR", "IS": "GB",
	"JE": "GB", "JM": "US", "JO": "TR",
	"KE": "GB", "KG": "RU", "KH": "SG", "KI": "AU", "KM": "FR", "KN": "US", "KP": "KR",
	"KW": "IN", "KY": "US", "KZ": "RU",
	"LA": "SG", "LB": "FR", "LC": "US", "LI": "DE", "LK": "IN", "LR": "US", "LS": "GB",
	"LT": "DE", "LU": "FR", "LV": "DE", "LY": "TR",
	"MA": "FR", "MC": "FR", "MD": "RU", "ME": "DE", "MF": "FR", "MG": "FR", "MH": "US",
	"MK": "DE", "ML": "FR", "MM": "SG", "MN": "RU", "MO": "HK", "MP": "US", "MQ": "FR",
	"MR": "FR", "MS": "US", "MT": "GB", "MU": "FR", "MV": "IN", "MW": "GB", "MY": "SG",
	"MZ": "PT",
	"NA": "GB", "NC": "FR", "NE": "FR", "NF": "AU", "NG": "GB", "NI": "ES", "NO": "GB",
	"NP": "IN", "NR": "AU", "NU": "NZ",
	"OM": "IN",
	"PA": "ES", "PE": "ES", "PF": "FR", "PG": "AU", "PH": "US", "PK": "IN", "PL": "DE",
	"PM": "FR", "PN": "NZ", "PR": "US", "PS": "TR", "PW": "US", "PY": "ES",
	"QA": "IN",
	"RE": "FR", "RO": "IT", "RS": "DE", "RW": "FR",
	"SA": "IN", "SB": "AU", "SC": "FR", "SD": "GB", "SE": "GB", "SH": "GB", "SI": "DE",
	"SJ": "NO", "SK": "DE", "SL": "GB", "SM": "IT", "SN": "FR", "SO": "GB", "SR": "NL",
	"SS": "GB", "ST": "PT", "SV": "ES", "SX": "NL", "SY": "TR", "SZ": "GB",
	"TC": "US", "TD": "FR", "TF": "FR", "TG": "FR", "TH": "SG", "TJ": "RU", "TK": "NZ",
	"TL": "PT", "TM": "TR", "TN": "FR", "TO": "AU", "TT": "US", "TV": "AU", "TZ": "GB",
	"UA": "RU", "UG": "GB", "UM": "US", "UY": "ES", "UZ": "RU",
	"VA": "IT", "VC": "US", "VE": "ES", "VG": "US", "VI": "US", "VN": "SG", "VU": "AU",
	"WF": "FR", "WS": "NZ",
	"YE": "IN", "YT": "FR",
	"ZA": "GB", "ZM": "GB", "ZW": "GB",
}

var autoFPCountryNames = map[string]string{
	"AD": "Andorra", "AE": "United Arab Emirates", "AF": "Afghanistan", "AG": "Antigua and Barbuda",
	"AI": "Anguilla", "AL": "Albania", "AM": "Armenia", "AO": "Angola", "AQ": "Antarctica",
	"AR": "Argentina", "AS": "American Samoa", "AT": "Austria", "AU": "Australia", "AW": "Aruba",
	"AX": "Aland Islands", "AZ": "Azerbaijan", "BA": "Bosnia and Herzegovina", "BB": "Barbados",
	"BD": "Bangladesh", "BE": "Belgium", "BF": "Burkina Faso", "BG": "Bulgaria", "BH": "Bahrain",
	"BI": "Burundi", "BJ": "Benin", "BL": "Saint Barthelemy", "BM": "Bermuda", "BN": "Brunei",
	"BO": "Bolivia", "BQ": "Caribbean Netherlands", "BR": "Brazil", "BS": "Bahamas", "BT": "Bhutan",
	"BV": "Bouvet Island", "BW": "Botswana", "BY": "Belarus", "BZ": "Belize", "CA": "Canada",
	"CC": "Cocos (Keeling) Islands", "CD": "Democratic Republic of the Congo",
	"CF": "Central African Republic", "CG": "Republic of the Congo", "CH": "Switzerland",
	"CI": "Cote d'Ivoire", "CK": "Cook Islands", "CL": "Chile", "CM": "Cameroon", "CN": "China",
	"CO": "Colombia", "CR": "Costa Rica", "CU": "Cuba", "CV": "Cape Verde", "CW": "Curacao",
	"CX": "Christmas Island", "CY": "Cyprus", "CZ": "Czechia", "DE": "Germany", "DJ": "Djibouti",
	"DK": "Denmark", "DM": "Dominica", "DO": "Dominican Republic", "DZ": "Algeria",
	"EC": "Ecuador", "EE": "Estonia", "EG": "Egypt", "EH": "Western Sahara", "ER": "Eritrea",
	"ES": "Spain", "ET": "Ethiopia", "FI": "Finland", "FJ": "Fiji", "FK": "Falkland Islands",
	"FM": "Micronesia", "FO": "Faroe Islands", "FR": "France", "GA": "Gabon", "GB": "United Kingdom",
	"GD": "Grenada", "GE": "Georgia", "GF": "French Guiana", "GG": "Guernsey", "GH": "Ghana",
	"GI": "Gibraltar", "GL": "Greenland", "GM": "Gambia", "GN": "Guinea", "GP": "Guadeloupe",
	"GQ": "Equatorial Guinea", "GR": "Greece", "GS": "South Georgia and the South Sandwich Islands",
	"GT": "Guatemala", "GU": "Guam", "GW": "Guinea-Bissau", "GY": "Guyana", "HK": "Hong Kong",
	"HM": "Heard Island and McDonald Islands", "HN": "Honduras", "HR": "Croatia", "HT": "Haiti",
	"HU": "Hungary", "ID": "Indonesia", "IE": "Ireland", "IL": "Israel", "IM": "Isle of Man",
	"IN": "India", "IO": "British Indian Ocean Territory", "IQ": "Iraq", "IR": "Iran",
	"IS": "Iceland", "IT": "Italy", "JE": "Jersey", "JM": "Jamaica", "JO": "Jordan", "JP": "Japan",
	"KE": "Kenya", "KG": "Kyrgyzstan", "KH": "Cambodia", "KI": "Kiribati", "KM": "Comoros",
	"KN": "Saint Kitts and Nevis", "KP": "North Korea", "KR": "South Korea", "KW": "Kuwait",
	"KY": "Cayman Islands", "KZ": "Kazakhstan", "LA": "Laos", "LB": "Lebanon", "LC": "Saint Lucia",
	"LI": "Liechtenstein", "LK": "Sri Lanka", "LR": "Liberia", "LS": "Lesotho", "LT": "Lithuania",
	"LU": "Luxembourg", "LV": "Latvia", "LY": "Libya", "MA": "Morocco", "MC": "Monaco",
	"MD": "Moldova", "ME": "Montenegro", "MF": "Saint Martin", "MG": "Madagascar",
	"MH": "Marshall Islands", "MK": "North Macedonia", "ML": "Mali", "MM": "Myanmar",
	"MN": "Mongolia", "MO": "Macao", "MP": "Northern Mariana Islands", "MQ": "Martinique",
	"MR": "Mauritania", "MS": "Montserrat", "MT": "Malta", "MU": "Mauritius", "MV": "Maldives",
	"MW": "Malawi", "MX": "Mexico", "MY": "Malaysia", "MZ": "Mozambique", "NA": "Namibia",
	"NC": "New Caledonia", "NE": "Niger", "NF": "Norfolk Island", "NG": "Nigeria", "NI": "Nicaragua",
	"NL": "Netherlands", "NO": "Norway", "NP": "Nepal", "NR": "Nauru", "NU": "Niue",
	"NZ": "New Zealand", "OM": "Oman", "PA": "Panama", "PE": "Peru", "PF": "French Polynesia",
	"PG": "Papua New Guinea", "PH": "Philippines", "PK": "Pakistan", "PL": "Poland",
	"PM": "Saint Pierre and Miquelon", "PN": "Pitcairn Islands", "PR": "Puerto Rico",
	"PS": "Palestine", "PT": "Portugal", "PW": "Palau", "PY": "Paraguay", "QA": "Qatar",
	"RE": "Reunion", "RO": "Romania", "RS": "Serbia", "RU": "Russia", "RW": "Rwanda",
	"SA": "Saudi Arabia", "SB": "Solomon Islands", "SC": "Seychelles", "SD": "Sudan",
	"SE": "Sweden", "SG": "Singapore", "SH": "Saint Helena", "SI": "Slovenia", "SJ": "Svalbard and Jan Mayen",
	"SK": "Slovakia", "SL": "Sierra Leone", "SM": "San Marino", "SN": "Senegal", "SO": "Somalia",
	"SR": "Suriname", "SS": "South Sudan", "ST": "Sao Tome and Principe", "SV": "El Salvador",
	"SX": "Sint Maarten", "SY": "Syria", "SZ": "Eswatini", "TC": "Turks and Caicos Islands",
	"TD": "Chad", "TF": "French Southern Territories", "TG": "Togo", "TH": "Thailand",
	"TJ": "Tajikistan", "TK": "Tokelau", "TL": "Timor-Leste", "TM": "Turkmenistan",
	"TN": "Tunisia", "TO": "Tonga", "TR": "Turkey", "TT": "Trinidad and Tobago", "TV": "Tuvalu",
	"TW": "Taiwan", "TZ": "Tanzania", "UA": "Ukraine", "UG": "Uganda",
	"UM": "United States Minor Outlying Islands", "US": "United States", "UY": "Uruguay",
	"UZ": "Uzbekistan", "VA": "Vatican City", "VC": "Saint Vincent and the Grenadines",
	"VE": "Venezuela", "VG": "British Virgin Islands", "VI": "United States Virgin Islands",
	"VN": "Vietnam", "VU": "Vanuatu", "WF": "Wallis and Futuna", "WS": "Samoa", "YE": "Yemen",
	"YT": "Mayotte", "ZA": "South Africa", "ZM": "Zambia", "ZW": "Zimbabwe",
}

var autoFPHardwareConcurrencyProfiles = []int{4, 6, 8, 12, 16}

func (o *FirefoxOptions) withAutoFPFile() (*FirefoxOptions, error) {
	o.ensureDefaults()
	if o.rejectAutoFPMutation("WithAutoFPFile") {
		return o, o.autoFPMutationError()
	}
	width, height := o.resolveAutoFPWindowSize()

	ipInfo, err := fetchAutoFPIPInfo(o.proxy)
	if err != nil {
		return o, err
	}
	proxyAuth, err := extractProxyAuth(o.proxy)
	if err != nil {
		return o, err
	}

	canvasSeed, err := randomPositiveInt(defaultCanvasSeedMax)
	if err != nil {
		return o, err
	}

	webglProfile := pickAutoFPWebGLProfile()
	hardwareConcurrency := pickAutoFPHardwareConcurrency()
	fingerprintText, err := buildAutoFPFingerprintText(ipInfo, proxyAuth, webglProfile, width, height, hardwareConcurrency, canvasSeed)
	if err != nil {
		return o, err
	}

	fpfile, err := os.CreateTemp("", "ruyipage-fp-*.txt")
	if err != nil {
		return o, err
	}

	path := fpfile.Name()
	if _, err := fpfile.WriteString(fingerprintText); err != nil {
		_ = fpfile.Close()
		_ = os.Remove(path)
		return o, err
	}
	if err := fpfile.Close(); err != nil {
		_ = os.Remove(path)
		return o, err
	}

	o.SetManagedFPFile(path)
	o.freezeAutoFP(width, height)
	return o, nil
}

func buildAutoFPFingerprintText(
	ipInfo autoFPIPInfoResponse,
	proxyAuth autoFPProxyAuth,
	webglProfile autoFPWebGLProfile,
	width int,
	height int,
	hardwareConcurrency int,
	canvasSeed int,
) (string, error) {
	normalizedIPInfo, ip, voiceProfile, err := normalizeAutoFPIPInfo(ipInfo)
	if err != nil {
		return "", err
	}

	lines := []string{"webdriver:0"}
	if ipv4 := ip.To4(); ipv4 != nil {
		ipv4Text := ipv4.String()
		lines = append(lines,
			"local_webrtc_ipv4:"+sanitizeFPValue(ipv4Text),
			"public_webrtc_ipv4:"+sanitizeFPValue(ipv4Text),
		)
	} else {
		ipv6Text := ip.String()
		lines = append(lines,
			"local_webrtc_ipv6:"+sanitizeFPValue(ipv6Text),
			"public_webrtc_ipv6:"+sanitizeFPValue(ipv6Text),
		)
	}

	lines = append(lines,
		"timezone:"+sanitizeFPValue(normalizedIPInfo.Timezone),
		"language:"+sanitizeFPValue(voiceProfile.Language),
		"speech.voices.local:"+sanitizeFPValue(strings.Join(voiceProfile.VoiceNames, "|")),
		"speech.voices.local.langs:"+sanitizeFPValue(strings.Join(voiceProfile.VoiceLangs, "|")),
		"speech.voices.default.name:"+sanitizeFPValue(voiceProfile.DefaultName),
		"speech.voices.default.lang:"+sanitizeFPValue(voiceProfile.DefaultLang),
		"font_system:"+defaultWindowsFontSystem,
		"useragent:"+sanitizeFPValue(defaultWindowsUserAgent),
		fmt.Sprintf("hardwareConcurrency:%d", hardwareConcurrency),
		"webgl.vendor:"+sanitizeFPValue(webglProfile.Vendor),
		"webgl.renderer:"+sanitizeFPValue(webglProfile.Renderer),
		"webgl.version:"+defaultWebGLVersion,
		"webgl.glsl_version:"+defaultWebGLGLSLVersion,
		"webgl.unmasked_vendor:"+sanitizeFPValue(webglProfile.Vendor),
		"webgl.unmasked_renderer:"+sanitizeFPValue(webglProfile.Renderer),
		fmt.Sprintf("webgl.max_texture_size:%d", defaultWebGLMaxTexture),
		fmt.Sprintf("webgl.max_cube_map_texture_size:%d", defaultWebGLMaxCubeMap),
		fmt.Sprintf("webgl.max_texture_image_units:%d", defaultWebGLImageUnits),
		fmt.Sprintf("webgl.max_vertex_attribs:%d", defaultWebGLVertexAttr),
		fmt.Sprintf("webgl.aliased_point_size_max:%d", defaultWebGLPointSize),
		fmt.Sprintf("webgl.max_viewport_dim:%d", defaultWebGLViewportDim),
		fmt.Sprintf("width:%d", width),
		fmt.Sprintf("height:%d", height),
		fmt.Sprintf("canvas:%d", canvasSeed),
	)

	if proxyAuth.HasCredentials {
		lines = append(lines, "httpauth.username:"+sanitizeFPValue(proxyAuth.Username))
		lines = append(lines, "httpauth.password:"+sanitizeFPValue(proxyAuth.Password))
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func fetchAutoFPFingerprintProfile(proxy string) (autoFPIPInfoResponse, error) {
	transport, err := buildProxyAwareHTTPTransport(proxy)
	if err != nil {
		return autoFPIPInfoResponse{}, err
	}

	client := &http.Client{
		Timeout:   autoFPIPInfoRequestTimeout,
		Transport: transport,
	}
	ctx, cancel := context.WithTimeout(context.Background(), autoFPIPInfoRequestTimeout)
	defer cancel()

	resultsCh := make(chan autoFPIPProviderResult, len(autoFPIPProviders))
	for index, provider := range autoFPIPProviders {
		go func(index int, provider autoFPIPProvider) {
			response, fetchErr := provider.fetch(ctx, client)
			resultsCh <- autoFPIPProviderResult{
				index:    index,
				name:     provider.name,
				response: response,
				err:      fetchErr,
			}
		}(index, provider)
	}

	results := make([]autoFPIPProviderResult, len(autoFPIPProviders))
	var merged autoFPIPInfoResponse
	var errorsList []string

	for remaining := len(autoFPIPProviders); remaining > 0; remaining-- {
		var result autoFPIPProviderResult
		select {
		case result = <-resultsCh:
		case <-ctx.Done():
			errorsList = append(errorsList, ctx.Err().Error())
			remaining = 0
			continue
		}
		results[result.index] = result
		if result.err != nil {
			errorsList = append(errorsList, fmt.Sprintf("%s: %v", result.name, result.err))
			continue
		}

		normalized, _, _, err := normalizeAutoFPIPInfo(result.response)
		if err == nil {
			cancel()
			return normalized, nil
		}

		errorsList = append(errorsList, fmt.Sprintf("%s: %v", result.name, err))
	}

	for _, result := range results {
		if result.name == "" || result.err != nil {
			continue
		}
		merged = mergeAutoFPIPInfo(merged, result.response)
		if completeAutoFPIPInfo(merged) {
			break
		}
	}

	normalized, _, _, err := normalizeAutoFPIPInfo(merged)
	if err != nil {
		if len(errorsList) == 0 {
			return autoFPIPInfoResponse{}, err
		}
		return autoFPIPInfoResponse{}, fmt.Errorf("自动指纹 IP 信息获取失败：%s；最终结果校验失败：%v", strings.Join(errorsList, "; "), err)
	}
	return normalized, nil
}

func fetchIP234FingerprintProfile(proxy string) (autoFPIPInfoResponse, error) {
	transport, err := buildProxyAwareHTTPTransport(proxy)
	if err != nil {
		return autoFPIPInfoResponse{}, err
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	req, err := http.NewRequest(http.MethodGet, autoFPIPInfoURL, nil)
	if err != nil {
		return autoFPIPInfoResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultWindowsUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return autoFPIPInfoResponse{}, fmt.Errorf("请求 ip234 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return autoFPIPInfoResponse{}, fmt.Errorf("ip234 返回异常状态 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result autoFPIPInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return autoFPIPInfoResponse{}, fmt.Errorf("解析 ip234 响应失败: %w", err)
	}
	if _, _, _, err := normalizeAutoFPIPInfo(result); err != nil {
		return autoFPIPInfoResponse{}, err
	}
	return result, nil
}

func fetchIPAPIFingerprintProfile(ctx context.Context, client *http.Client) (autoFPIPInfoResponse, error) {
	type response struct {
		Status      string  `json:"status"`
		Query       string  `json:"query"`
		City        string  `json:"city"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Timezone    string  `json:"timezone"`
		Region      string  `json:"region"`
		RegionName  string  `json:"regionName"`
		Zip         string  `json:"zip"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Message     string  `json:"message"`
	}

	var result response
	if err := fetchAutoFPJSON(ctx, client, autoFPIPAPIURL, &result); err != nil {
		return autoFPIPInfoResponse{}, err
	}
	if !strings.EqualFold(strings.TrimSpace(result.Status), "success") {
		message := strings.TrimSpace(result.Message)
		if message == "" {
			message = "状态不是 success"
		}
		return autoFPIPInfoResponse{}, fmt.Errorf("返回失败状态: %s", message)
	}
	region := strings.TrimSpace(result.RegionName)
	if region == "" {
		region = strings.TrimSpace(result.Region)
	}
	return autoFPIPInfoResponse{
		IP:          result.Query,
		City:        result.City,
		Country:     result.Country,
		CountryCode: result.CountryCode,
		Timezone:    result.Timezone,
		Region:      region,
	}, nil
}

func fetchIPAPICOFingerprintProfile(ctx context.Context, client *http.Client) (autoFPIPInfoResponse, error) {
	type response struct {
		IP          string  `json:"ip"`
		City        string  `json:"city"`
		Region      string  `json:"region"`
		Country     string  `json:"country"`
		CountryName string  `json:"country_name"`
		CountryCode string  `json:"country_code"`
		Timezone    string  `json:"timezone"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
	}

	var result response
	if err := fetchAutoFPJSON(ctx, client, autoFPIPAPICOURL, &result); err != nil {
		return autoFPIPInfoResponse{}, err
	}
	return autoFPIPInfoResponse{
		IP:          result.IP,
		City:        result.City,
		Country:     firstNonEmpty(result.CountryName, autoFPCountryNameFromCode(result.CountryCode), result.Country),
		CountryCode: firstNonEmpty(result.CountryCode, result.Country),
		Timezone:    result.Timezone,
		Region:      result.Region,
	}, nil
}

func fetchIPWhoIsFingerprintProfile(ctx context.Context, client *http.Client) (autoFPIPInfoResponse, error) {
	type timezoneResponse struct {
		ID string `json:"id"`
	}
	type response struct {
		IP          string           `json:"ip"`
		Success     bool             `json:"success"`
		Message     string           `json:"message"`
		City        string           `json:"city"`
		Region      string           `json:"region"`
		Country     string           `json:"country"`
		CountryCode string           `json:"country_code"`
		Timezone    timezoneResponse `json:"timezone"`
	}

	var result response
	if err := fetchAutoFPJSON(ctx, client, autoFPIPWhoIsURL, &result); err != nil {
		return autoFPIPInfoResponse{}, err
	}
	if !result.Success {
		message := strings.TrimSpace(result.Message)
		if message == "" {
			message = "success=false"
		}
		return autoFPIPInfoResponse{}, fmt.Errorf("返回失败状态: %s", message)
	}
	return autoFPIPInfoResponse{
		IP:          result.IP,
		City:        result.City,
		Country:     result.Country,
		CountryCode: result.CountryCode,
		Timezone:    result.Timezone.ID,
		Region:      result.Region,
	}, nil
}

func fetchIPInfoFingerprintProfile(ctx context.Context, client *http.Client) (autoFPIPInfoResponse, error) {
	type response struct {
		IP       string `json:"ip"`
		City     string `json:"city"`
		Region   string `json:"region"`
		Country  string `json:"country"`
		Timezone string `json:"timezone"`
	}

	var result response
	if err := fetchAutoFPJSON(ctx, client, autoFPIPInfoIOURL, &result); err != nil {
		return autoFPIPInfoResponse{}, err
	}
	return autoFPIPInfoResponse{
		IP:          result.IP,
		City:        result.City,
		Country:     autoFPCountryNameFromCode(result.Country),
		CountryCode: result.Country,
		Timezone:    result.Timezone,
		Region:      result.Region,
	}, nil
}

func fetchFreeIPAPIFingerprintProfile(ctx context.Context, client *http.Client) (autoFPIPInfoResponse, error) {
	type response struct {
		IPAddress   string   `json:"ipAddress"`
		CityName    string   `json:"cityName"`
		RegionName  string   `json:"regionName"`
		CountryName string   `json:"countryName"`
		CountryCode string   `json:"countryCode"`
		TimeZones   []string `json:"timeZones"`
	}

	var result response
	if err := fetchAutoFPJSON(ctx, client, autoFPFreeIPAPIURL, &result); err != nil {
		return autoFPIPInfoResponse{}, err
	}
	timezone := ""
	if len(result.TimeZones) > 0 {
		timezone = result.TimeZones[0]
	}
	return autoFPIPInfoResponse{
		IP:          result.IPAddress,
		City:        result.CityName,
		Country:     result.CountryName,
		CountryCode: result.CountryCode,
		Timezone:    timezone,
		Region:      result.RegionName,
	}, nil
}

func fetchAutoFPJSON(ctx context.Context, client *http.Client, endpoint string, target any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if client == nil {
		client = &http.Client{Timeout: autoFPIPInfoRequestTimeout}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultWindowsUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("返回异常状态 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	return nil
}

func buildProxyAwareHTTPTransport(proxy string) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if strings.TrimSpace(proxy) == "" {
		return transport, nil
	}

	proxyURL, err := parseProxyURL(proxy)
	if err != nil {
		return nil, err
	}
	transport.Proxy = http.ProxyURL(proxyURL)
	return transport, nil
}

func parseProxyURL(proxy string) (*url.URL, error) {
	trimmed := strings.TrimSpace(proxy)
	if trimmed == "" {
		return nil, nil
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "http://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("代理地址无效: %w", err)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("代理地址缺少 host: %s", proxy)
	}
	return parsed, nil
}

func extractProxyAuth(proxy string) (autoFPProxyAuth, error) {
	proxyURL, err := parseProxyURL(proxy)
	if err != nil {
		return autoFPProxyAuth{}, err
	}
	if proxyURL == nil || proxyURL.User == nil {
		return autoFPProxyAuth{}, nil
	}

	username := proxyURL.User.Username()
	password, hasPassword := proxyURL.User.Password()
	if username == "" && !hasPassword {
		return autoFPProxyAuth{}, nil
	}

	return autoFPProxyAuth{
		Username:       username,
		Password:       password,
		HasCredentials: true,
	}, nil
}

func resolveAutoFPVoiceProfile(countryCode string) (autoFPVoiceProfile, error) {
	normalizedCountryCode := strings.ToUpper(strings.TrimSpace(countryCode))
	if profile, ok := autoFPVoiceProfiles[normalizedCountryCode]; ok {
		return profile, nil
	}
	if fallbackCountryCode, ok := autoFPVoiceProfileAliases[normalizedCountryCode]; ok {
		if profile, ok := autoFPVoiceProfiles[fallbackCountryCode]; ok {
			return profile, nil
		}
	}
	return autoFPVoiceProfile{}, fmt.Errorf("自动指纹返回了未支持的 country_code: %q", countryCode)
}

func normalizeAutoFPIPInfo(ipInfo autoFPIPInfoResponse) (autoFPIPInfoResponse, net.IP, autoFPVoiceProfile, error) {
	normalized := autoFPIPInfoResponse{
		IP:          strings.TrimSpace(ipInfo.IP),
		City:        strings.TrimSpace(ipInfo.City),
		Country:     strings.TrimSpace(ipInfo.Country),
		CountryCode: strings.ToUpper(strings.TrimSpace(ipInfo.CountryCode)),
		Timezone:    strings.TrimSpace(ipInfo.Timezone),
		Region:      strings.TrimSpace(ipInfo.Region),
	}
	if normalized.Country == "" && normalized.CountryCode != "" {
		normalized.Country = autoFPCountryNameFromCode(normalized.CountryCode)
	}

	if normalized.IP == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("自动指纹 IP 信息缺少 ip 字段")
	}
	ip := net.ParseIP(normalized.IP)
	if ip == nil {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("自动指纹 IP 信息返回了无效 IP: %q", ipInfo.IP)
	}
	if normalized.City == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("自动指纹 IP 信息缺少 city 字段")
	}
	if normalized.Country == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("自动指纹 IP 信息缺少 country 字段")
	}
	if !isValidAutoFPCountryCode(normalized.CountryCode) {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("自动指纹 IP 信息返回了无效 country_code: %q", ipInfo.CountryCode)
	}
	if normalized.Timezone == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("自动指纹 IP 信息缺少 timezone 字段")
	}
	if _, err := time.LoadLocation(normalized.Timezone); err != nil {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("自动指纹无法加载 timezone: %q", ipInfo.Timezone)
	}
	if normalized.Region == "" {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, fmt.Errorf("自动指纹 IP 信息缺少 region 字段")
	}

	voiceProfile, err := resolveAutoFPVoiceProfile(normalized.CountryCode)
	if err != nil {
		return autoFPIPInfoResponse{}, nil, autoFPVoiceProfile{}, err
	}

	return normalized, ip, voiceProfile, nil
}

func mergeAutoFPIPInfo(current autoFPIPInfoResponse, incoming autoFPIPInfoResponse) autoFPIPInfoResponse {
	if strings.TrimSpace(current.IP) == "" {
		current.IP = strings.TrimSpace(incoming.IP)
	}
	if strings.TrimSpace(current.City) == "" {
		current.City = strings.TrimSpace(incoming.City)
	}
	if strings.TrimSpace(current.Country) == "" {
		current.Country = strings.TrimSpace(incoming.Country)
	}
	if strings.TrimSpace(current.CountryCode) == "" {
		current.CountryCode = strings.ToUpper(strings.TrimSpace(incoming.CountryCode))
	}
	if strings.TrimSpace(current.Timezone) == "" {
		current.Timezone = strings.TrimSpace(incoming.Timezone)
	}
	if strings.TrimSpace(current.Region) == "" {
		current.Region = strings.TrimSpace(incoming.Region)
	}
	return current
}

func completeAutoFPIPInfo(value autoFPIPInfoResponse) bool {
	return strings.TrimSpace(value.IP) != "" &&
		strings.TrimSpace(value.City) != "" &&
		strings.TrimSpace(value.Country) != "" &&
		strings.TrimSpace(value.CountryCode) != "" &&
		strings.TrimSpace(value.Timezone) != "" &&
		strings.TrimSpace(value.Region) != ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func autoFPCountryNameFromCode(countryCode string) string {
	return autoFPCountryNames[strings.ToUpper(strings.TrimSpace(countryCode))]
}

func isValidAutoFPCountryCode(countryCode string) bool {
	if len(countryCode) != 2 {
		return false
	}
	for _, char := range countryCode {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

func pickAutoFPWebGLProfile() autoFPWebGLProfile {
	if len(autoFPWindowsWebGLProfiles) == 0 {
		return autoFPWebGLProfile{
			Vendor:   "Google Inc. (AMD)",
			Renderer: "ANGLE (AMD, AMD Radeon RX 6800 XT Direct3D11 vs_5_0 ps_5_0, D3D11)",
		}
	}

	index, err := randomPositiveInt(len(autoFPWindowsWebGLProfiles))
	if err != nil {
		return autoFPWindowsWebGLProfiles[0]
	}
	return autoFPWindowsWebGLProfiles[index-1]
}

func pickAutoFPHardwareConcurrency() int {
	actual := runtime.NumCPU()
	if actual <= 0 {
		return 8
	}

	best := autoFPHardwareConcurrencyProfiles[0]
	bestDistance := absInt(best - actual)
	for _, candidate := range autoFPHardwareConcurrencyProfiles[1:] {
		distance := absInt(candidate - actual)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func randomPositiveInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("随机上限必须大于 0，当前为 %d", max)
	}

	value, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()) + 1, nil
}

func sanitizeFPValue(value string) string {
	replacer := strings.NewReplacer("\r", " ", "\n", " ")
	return strings.TrimSpace(replacer.Replace(value))
}

func sameCleanPath(a string, b string) bool {
	if strings.TrimSpace(a) == "" || strings.TrimSpace(b) == "" {
		return false
	}

	left, err := filepath.Abs(a)
	if err != nil {
		left = a
	}
	right, err := filepath.Abs(b)
	if err != nil {
		right = b
	}
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
