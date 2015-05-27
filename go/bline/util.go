package bline

func SplitRGB(color int) (byte, byte, byte) {
	r := byte((color >> 16) & 0xFF)
	g := byte((color >> 8) & 0xFF)
	b := byte(color & 0xFF)
	return r, g, b
}

func CombineRGB(r, g, b byte) int {
	color := int(r) << 16
	color |= int(g) << 8
	color |= int(b)
	return color
}
