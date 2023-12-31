package common

const Splash = `     ┏━══════━┓       ███▙                 ▅             
┏━━━━┻━━━━━━━━┻━━━━┓  █  █                 █             
┠──────▀▀▀▀▀▀──────┨  ███🭪 ▟██▙ ▟██▙ ▟█▆█▙ ███▙ ▟██▙ █  █
┃ ▟██▙ ▕▚▞▚▞▏ ▟██▙ ┃  █  █ █  █ █  █ █ █ █ █  █ █  █  ▚▞ 
┃ ▜██▛  ○○○○  ▜██▛ ┃  █  █ █  █ █  █ █ █ █ █  █ █  █  ▞▚ 
┗━━━━━━━━━━━━━━━━━━┛  ███▛ ▜██▛ ▜██▛ █ █ █ ███▛ ▜██▛ █  █`

var LogoSprite = []string{
	LogoStyle.Render(`     ┏━══════━┓     
┏━━━━┻━━━━━━━━┻━━━━┓
┠──────▀▀▀▀▀▀──────┨
┃ ▟██▙ ▕▚▞▚▞▏ ▟██▙ ┃
┃ ▜██▛  ○○○○  ▜██▛ ┃
┗━━━━━━━━━━━━━━━━━━┛`),
	LogoActivityStyle.Render(`     ┏━══════━┓     
┏━━━━┻━━━━━━━━┻━━━━┓
┠──────▀▀▀▀▀▀──────┨
┃ ▟██▙ ▕▞▚▞▚▏ ▟██▙ ┃
┃ ▜██▛  ○○○○  ▜██▛ ┃
┗━━━━━━━━━━━━━━━━━━┛`),
	LogoStyle.Render(`     ┏━══════━┓     
┏━━━━┻━━━━━━━━┻━━━━┓
┠──────▀▀▀▀▀▀──────┨
┃ ▟██▙ ▕▚▞▚▞▏ ▟██▙ ┃
┃ ▜██▛  ◙○○○  ▜██▛ ┃
┗━━━━━━━━━━━━━━━━━━┛`),
	LogoActivityStyle.Render(`     ┏━══════━┓     
┏━━━━┻━━━━━━━━┻━━━━┓
┠──────▀▀▀▀▀▀──────┨
┃ ▟██▙ ▕▞▚▞▚▏ ▟██▙ ┃
┃ ▜██▛  ○◙○○  ▜██▛ ┃
┗━━━━━━━━━━━━━━━━━━┛`),
	LogoStyle.Render(`     ┏━══════━┓     
┏━━━━┻━━━━━━━━┻━━━━┓
┠──────▀▀▀▀▀▀──────┨
┃ ▟██▙ ▕▚▞▚▞▏ ▟██▙ ┃
┃ ▜██▛  ○○◙○  ▜██▛ ┃
┗━━━━━━━━━━━━━━━━━━┛`),
	LogoActivityStyle.Render(`     ┏━══════━┓     
┏━━━━┻━━━━━━━━┻━━━━┓
┠──────▀▀▀▀▀▀──────┨
┃ ▟██▙ ▕▞▚▞▚▏ ▟██▙ ┃
┃ ▜██▛  ○○○◙  ▜██▛ ┃
┗━━━━━━━━━━━━━━━━━━┛`),
}

const Banner = `█▀▄ █▀█ █▀█ █▄█ █▀▄ █▀█ █ █
█▀▄ █ █ █ █ █ █ █▀▄ █ █ ▄▀▄
▀▀  ▀▀▀ ▀▀▀ ▀ ▀ ▀▀  ▀▀▀ ▀ ▀`

var MiniLogo = LogoStyle.Render(`     ▁▁     
┏━━━┻━━┻━━━┓
┃▐█▌ ○○ ▐█▌┃
┗━━━━━━━━━━┛`)
