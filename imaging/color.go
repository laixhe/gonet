package imaging

import "image/color"

// Transparent 透明
var Transparent = color.RGBA{R: 0, G: 0, B: 0, A: 0}

// White 白色
var White = color.RGBA{R: 255, G: 255, B: 255, A: 255}

// Black 黑色
var Black = color.RGBA{R: 0, G: 0, B: 0, A: 255}

// Red 红色
var Red = color.RGBA{R: 255, G: 0, B: 0, A: 255}

// Green 绿色
var Green = color.RGBA{R: 0, G: 255, B: 0, A: 255}

// Blue 蓝色
var Blue = color.RGBA{R: 0, G: 0, B: 255, A: 255}

// Yellow 黄色
var Yellow = color.RGBA{R: 255, G: 255, B: 0, A: 255}

// Cyan 青色
var Cyan = color.RGBA{R: 0, G: 255, B: 255, A: 255}

// Magenta 品红色（洋红色）
var Magenta = color.RGBA{R: 255, G: 0, B: 255, A: 255}

// Gray 灰色
var Gray = color.RGBA{R: 128, G: 128, B: 128, A: 255}
