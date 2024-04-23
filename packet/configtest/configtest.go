package configtest

var (
	NotchianClientInformation = []byte{
		// locale
		0x05, 0x65, 0x6e, 0x5f, 0x75, 0x73,
		// view distance
		0x0c,
		// chat mode
		0x00,
		// chat colors enabled
		0x01,
		// displayed skin parts
		0x7f,
		// main hand
		0x01,
		// enable test filtering
		0x00,
		// allow server listings
		0x01,
	}
	NotchianClientInformationHeader = []byte{0x0e, 0x00}

	NotchianServerboundPlugin = []byte{
		15, 109, 105, 110, 101, 99, 114, 97, 102, 116, 58, 98,
		114, 97, 110, 100, 7, 118, 97, 110, 105, 108, 108, 97,
	}
	NotchianServerboundPluginHeader = []byte{0x19, 0x01}
)
